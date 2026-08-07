package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/unionai/cloud/gen/pb-go/common"
	"github.com/unionai/cloud/gen/pb-go/identity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

// UserResource defines the resource implementation.
type UserResource struct {
	conn identity.UserServiceClient
	org  string
}

// UserResourceModel describes the resource data model.
type UserResourceModel struct {
	Id        types.String `tfsdk:"id"`
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	Email     types.String `tfsdk:"email"`
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "User resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "User identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"first_name": schema.StringAttribute{
				MarkdownDescription: "User first name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"last_name": schema.StringAttribute{
				MarkdownDescription: "User last name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "User email",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*providerContext)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *providerContext, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.conn = identity.NewUserServiceClient(client.conn)
	if r.conn == nil {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *identity.UserServiceClient, got: %T. Please report this issue to the provider developers.", r.conn),
		)
		return
	}
	r.org = client.org
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.conn.CreateUser(ctx, &identity.CreateUserRequest{
		Spec: &common.UserSpec{
			Organization: r.org,
			FirstName:    data.FirstName.ValueString(),
			LastName:     data.LastName.ValueString(),
			Email:        data.Email.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating User",
			fmt.Sprintf("Could not create user, unexpected error: %s", err.Error()),
		)
		return
	}

	data.Id = types.StringValue(user.Id.Subject)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.conn.GetUser(ctx, &identity.GetUserRequest{
		Id: &common.UserIdentifier{
			Subject: data.Id.ValueString(),
		},
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// User has been deleted outside of Terraform, remove from state
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading User",
			fmt.Sprintf("Could not read user, unexpected error: %s", err.Error()),
		)
		return
	}

	data.FirstName = types.StringValue(user.User.Spec.FirstName)
	data.LastName = types.StringValue(user.User.Spec.LastName)
	data.Email = types.StringValue(user.User.Spec.Email)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.conn.DeleteUser(ctx, &identity.DeleteUserRequest{
		Id: &common.UserIdentifier{
			Subject: data.Id.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting User",
			fmt.Sprintf("Could not delete user, unexpected error: %s", err.Error()),
		)
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	users, err := r.conn.ListUsers(ctx, &identity.ListUsersRequest{
		Organization: r.org,
		Request: &common.ListRequest{
			Filters: []*common.Filter{
				{
					Field:    "email",
					Function: common.Filter_EQUAL,
					Values:   []string{req.ID},
				},
			},
		},
		IncludeSupportStaff: true,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch user", err.Error())
		return
	}
	if len(users.Users) == 0 {
		resp.Diagnostics.AddError("User not found", fmt.Sprintf("User %s not found", req.ID))
		return
	}

	user := users.Users[0]

	resp.State.Set(ctx, &UserResourceModel{
		Id:        types.StringValue(user.Id.Subject),
		FirstName: types.StringValue(user.Spec.FirstName),
		LastName:  types.StringValue(user.Spec.LastName),
		Email:     types.StringValue(user.Spec.Email),
	})
}
