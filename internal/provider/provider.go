package provider

import (
	"context"
	"crypto/tls"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Ensure UnionProvider satisfies various provider interfaces.
var _ provider.Provider = &UnionaiProvider{}

// UnionaiProvider defines the provider implementation.
type UnionaiProvider struct {
	version string
}

// UnionaiProviderModel describes the provider data model.
type UnionaiProviderModel struct {
	ApiKey      types.String `tfsdk:"api_key"`
	Org         types.String `tfsdk:"org"`
	AllowedOrgs types.Set    `tfsdk:"allowed_orgs"`
}

type providerContext struct {
	conn *grpc.ClientConn
	org  string
	host string
}

func (p *UnionaiProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "unionai"
	resp.Version = p.version
}

func (p *UnionaiProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Unionai API key",
				Optional:            true, // they can be specified by UNIONAI_API_KEY
			},
			"org": schema.StringAttribute{
				MarkdownDescription: "Union.ai organization name. If set, this takes precedence over the organization encoded in the API key or inferred from the API key host. Use this when the control plane's organization name differs from the URL it is served from.",
				Optional:            true,
			},
			"allowed_orgs": schema.SetAttribute{
				MarkdownDescription: "Unionai allowed orgs",
				Optional:            true, // they can be specified by UNIONAI_ALLOWED_ORGS
				ElementType:         types.StringType,
			},
		},
	}
}

func (p *UnionaiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data UnionaiProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("UNIONAI_API_KEY")
	if apiKey == "" {
		apiKey = data.ApiKey.ValueString()
	}
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Union.ai api_key is required",
			"Union.ai api_key can be specified by UNIONAI_API_KEY or api_key attribute.",
		)
		return
	}

	// Get OAuth2 token source using Union auth metadata where available.
	apiTokenConfig, err := GetApiToken(apiKey)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get OAuth2 token", err.Error())
		return
	}
	if data.Org.ValueString() != "" {
		apiTokenConfig.Org = data.Org.ValueString()
	}

	// Create gRPC connection with OAuth2 credentials
	conn, err := grpc.NewClient(
		apiTokenConfig.Host,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(NewTokenSourceCredentials(apiTokenConfig.TokenSource, apiTokenConfig.AuthorizationMetadataKey)),
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to connect to Unionai host", err.Error())
		return
	}

	client := &providerContext{
		conn: conn,
		org:  apiTokenConfig.Org,
		host: apiTokenConfig.Host,
	}
	resp.DataSourceData = client
	resp.ResourceData = client

	if len(data.AllowedOrgs.Elements()) > 0 {
		// Check if our org is allowed
		allowed := false
		for _, org := range data.AllowedOrgs.Elements() {
			if org.(types.String).ValueString() == client.org {
				allowed = true
				break
			}
		}
		if !allowed {
			resp.Diagnostics.AddError(
				"Union.ai org is not allowed",
				"Union.ai org "+client.org+" is not allowed. Please add it to allowed_orgs attribute.",
			)
			return
		}
	}
}

func (p *UnionaiProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewUserResource,
		NewUserAccessResource,
		NewRoleResource,
		NewPolicyResource,
		NewApiKeyResource,
		NewAppResource,
		NewAppAccessResource,
		NewTaskEnvironmentResource,
		NewProjectDomainAttributesResource,
		NewQueueResource,
	}
}

func (p *UnionaiProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectDataSource,
		NewUserDataSource,
		NewUserAccessDataSource,
		NewRoleDataSource,
		NewPolicyDataSource,
		NewApiKeyDataSource,
		NewAppDataSource,
		NewAppAccessDataSource,
		NewDataplaneDataSource,
		NewDataplanesDataSource,
		NewControlplaneDataSource,
		NewQueueDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &UnionaiProvider{
			version: version,
		}
	}
}
