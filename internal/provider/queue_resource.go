package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/unionai/cloud/gen/pb-go/common"
	"github.com/unionai/cloud/gen/pb-go/queue"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &QueueResource{}
var _ resource.ResourceWithImportState = &QueueResource{}
var _ resource.ResourceWithModifyPlan = &QueueResource{}

func NewQueueResource() resource.Resource {
	return &QueueResource{}
}

// QueueResource defines the resource implementation.
type QueueResource struct {
	conn queue.QueueServiceClient
	org  string
}

// QueueResourceModel describes the resource data model.
type QueueResourceModel struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	ClusterPoolName   types.String `tfsdk:"cluster_pool_name"`
	Clusters          types.Set    `tfsdk:"clusters"`
	RunConcurrency    types.Int64  `tfsdk:"run_concurrency"`
	ActionConcurrency types.Int64  `tfsdk:"action_concurrency"`
	Depth             types.Int64  `tfsdk:"depth"`
	Priority          types.String `tfsdk:"priority"`
	Fairness          types.String `tfsdk:"fairness"`
	Drain             types.Bool   `tfsdk:"drain"`
	State             types.String `tfsdk:"state"`
	ClusterManaged    types.Bool   `tfsdk:"cluster_managed"`
	AvailableClusters types.Set    `tfsdk:"available_clusters"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
	DeletedAt         types.String `tfsdk:"deleted_at"`
}

func (r *QueueResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_queue"
}

func (r *QueueResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Queue resource. Queues control admission and scheduling of runs: concurrency limits, queue depth, scheduling priority, cross-project fairness, and which clusters within a cluster pool work is routed to.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Queue identifier. Same as `name`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Queue name. Must match `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` and be at most 63 characters. Unique within the organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster_pool_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(defaultClusterPoolName),
				MarkdownDescription: "Cluster pool this queue routes into. Defaults to `default`, which is created on demand; any other pool must already exist. Can only be changed while the queue is drained.",
			},
			"clusters": schema.SetAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Clusters this queue routes to. Must be a subset of the assigned cluster pool. `[\"*\"]` routes to every enabled and healthy cluster in the pool, equally weighted, and must be the only element when used. An empty set means actions are not routed to any cluster.",
			},
			"run_concurrency": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: "Maximum number of runs that may be active simultaneously. `0` means unlimited.",
			},
			"action_concurrency": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: "Maximum number of actions (tasks) that may be in flight simultaneously. `0` means unlimited.",
			},
			"depth": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: "Maximum total items the queue can hold (in-progress plus waiting). `0` means unlimited.",
			},
			"priority": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("medium"),
				MarkdownDescription: "Scheduling priority relative to other queues: `min`, `medium`, or `max`. Scheduling is strict — `max` always runs before `medium`, which always runs before `min`.",
			},
			"fairness": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("round_robin"),
				MarkdownDescription: "Cross-project scheduling policy within this queue: `round_robin` or `shuffle_interleave`. The unit of fairness is per-project.",
			},
			"drain": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When true, request that the queue stop accepting new work and drain existing work. Terraform reports progress through the computed `state` attribute.",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Queue lifecycle state: `active`, `draining`, or `drained`. Managed by the control plane.",
			},
			"cluster_managed": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether this queue is the implicit queue owned by a cluster with the same name. Cluster-managed queues cannot move cluster pool or routing independently of that cluster.",
			},
			"available_clusters": schema.SetAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Names of the clusters this queue can currently route to, after filtering `clusters` down to those that are active and healthy.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation timestamp, RFC 3339.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Last update timestamp, RFC 3339.",
			},
			"deleted_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Soft-deletion timestamp, RFC 3339. Null for live queues.",
			},
		},
	}
}

func (r *QueueResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*providerContext)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *providerContext, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.conn = queue.NewQueueServiceClient(client.conn)
	r.org = client.org
}

// ModifyPlan rejects cluster pool changes the control plane would reject. A drained queue can
// be moved to another pool, but active/draining queues must finish draining first, and
// cluster-managed queues cannot be repointed independently of their cluster.
func (r *QueueResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		// Create or destroy; nothing to compare.
		return
	}

	var plan, state QueueResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ClusterPoolName.IsUnknown() || state.ClusterPoolName.IsNull() {
		return
	}

	if plan.ClusterPoolName.ValueString() != state.ClusterPoolName.ValueString() {
		if state.ClusterManaged.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				path.Root("cluster_pool_name"),
				"Cluster-managed queue cannot change cluster pool",
				fmt.Sprintf(
					"Queue %q is managed by a cluster with the same name. Move the cluster to change its co-named queue's pool and routing.",
					state.Name.ValueString(),
				),
			)
			return
		}

		if state.State.ValueString() == "drained" {
			return
		}

		resp.Diagnostics.AddAttributeError(
			path.Root("cluster_pool_name"),
			"Queue cluster pool cannot be changed while queue is active",
			fmt.Sprintf(
				"The control plane does not permit moving queue %q from cluster pool %q to %q while it is %s. "+
					"Drain the queue first, then apply the pool change.",
				state.Name.ValueString(), state.ClusterPoolName.ValueString(), plan.ClusterPoolName.ValueString(),
				state.State.ValueString(),
			),
		)
	}
}

func (r *QueueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data QueueResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = data.Name

	spec, diags := buildQueueSpec(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := &queue.CreateQueueRequest{
		Id:    r.queueId(data.Name.ValueString()),
		Spec:  spec,
		State: queue.QueueState_QUEUE_STATE_ACTIVE,
	}
	if data.Drain.ValueBool() {
		createRequest.State = queue.QueueState_QUEUE_STATE_DRAINED
	}

	tflog.Debug(ctx, "CreateQueue request", map[string]interface{}{
		"queue(create)": createRequest,
	})
	created, err := r.conn.CreateQueue(ctx, createRequest)
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			name := data.Name.ValueString()
			resp.Diagnostics.AddError(
				"Queue already exists",
				fmt.Sprintf(
					"Queue %q already exists in organization %q. Queue names stay reserved even after a queue is drained, "+
						"so this can happen after a previous destroy. Import it instead:\n\n"+
						"  terraform import unionai_queue.example %s",
					name, r.org, name,
				),
			)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create queue, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(applyQueueToModel(&data, created.GetQueue())...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *QueueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data QueueResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	got, err := r.conn.GetQueue(ctx, &queue.GetQueueRequest{
		Id: r.queueId(data.Id.ValueString()),
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read queue, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(applyQueueToModel(&data, got.GetQueue())...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *QueueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data QueueResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = data.Name

	// UpdateQueue fully replaces the spec rather than merging set fields, so always send
	// every field from the plan — including cluster_pool_name, which the server requires.
	spec, diags := buildQueueSpec(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateRequest := &queue.UpdateQueueRequest{
		Id:   r.queueId(data.Name.ValueString()),
		Spec: spec,
	}

	tflog.Debug(ctx, "UpdateQueue request", map[string]interface{}{
		"queue(update)": updateRequest,
	})
	updated, err := r.conn.UpdateQueue(ctx, updateRequest)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			resp.Diagnostics.AddError(
				"Queue not found",
				fmt.Sprintf("Queue %q no longer exists in organization %q. Run `terraform refresh` and apply again.", data.Name.ValueString(), r.org),
			)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update queue, got error: %s", err))
		return
	}

	q, err := r.applyDrainIntent(ctx, data.Name.ValueString(), updated.GetQueue(), data.Drain.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	resp.Diagnostics.Append(applyQueueToModel(&data, q)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete drains the queue, then soft-deletes it once the control plane reports DRAINED. The
// queue name remains reserved until the queue is undeleted.
func (r *QueueResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data QueueResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Id.ValueString()

	// Check the current state first: draining an already-drained queue is a FailedPrecondition,
	// which we want to treat as a no-op rather than a failure.
	got, err := r.conn.GetQueue(ctx, &queue.GetQueueRequest{Id: r.queueId(name)})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read queue before draining, got error: %s", err))
		return
	}

	if got.GetQueue().GetDeletedAt() != nil {
		return
	}

	state := got.GetQueue().GetStatus().GetState()
	if state == queue.QueueState_QUEUE_STATE_ACTIVE {
		if !data.Drain.ValueBool() {
			resp.Diagnostics.AddError(
				"Queue must be drained before destroy",
				fmt.Sprintf(
					"Queue %q is active. Set `drain = true` on the resource and apply first. "+
						"Once the computed `state` reaches `drained`, run destroy again.",
					name,
				),
			)
			return
		}

		updated, err := r.conn.UpdateQueueState(ctx, &queue.UpdateQueueStateRequest{
			Id:    r.queueId(name),
			State: queue.QueueState_QUEUE_STATE_DRAINING,
		})
		if err != nil {
			switch status.Code(err) {
			case codes.NotFound:
				return
			case codes.FailedPrecondition:
				resp.Diagnostics.AddError(
					"Unable to drain queue",
					fmt.Sprintf(
						"The control plane refused to drain queue %q: %s\n\n"+
							"This usually means the queue is referenced as run.default_queue in settings at one or more scopes. "+
							"Update or unset those settings first.",
						name, err,
					),
				)
				return
			default:
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to drain queue, got error: %s", err))
				return
			}
		}
		state = updated.GetQueue().GetStatus().GetState()
	}

	if state != queue.QueueState_QUEUE_STATE_DRAINED {
		resp.Diagnostics.AddError(
			"Queue is not drained",
			fmt.Sprintf(
				"Queue %q is %s. The control plane only deletes drained queues. Wait until "+
					"the computed `state` reaches `drained`, then run destroy again.",
				name, queueStateToTerraform(state).ValueString(),
			),
		)
		return
	}

	_, err = r.conn.DeleteQueue(ctx, &queue.DeleteQueueRequest{Id: r.queueId(name)})
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			return
		case codes.FailedPrecondition:
			resp.Diagnostics.AddError(
				"Unable to delete queue",
				fmt.Sprintf(
					"The control plane refused to delete queue %q: %s\n\n"+
						"Queues can only be deleted after they are drained, and queues referenced as run.default_queue "+
						"must be unset from settings first.",
					name, err,
				),
			)
			return
		default:
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete queue, got error: %s", err))
			return
		}
	}

	resp.Diagnostics.AddWarning(
		"Queue soft-deleted",
		fmt.Sprintf(
			"Queue %q was soft-deleted in organization %q. %s",
			name, r.org, queueNameReservedNote,
		),
	)
}

func (r *QueueResource) applyDrainIntent(ctx context.Context, name string, q *queue.Queue, drain bool) (*queue.Queue, error) {
	if q == nil {
		return nil, fmt.Errorf("the control plane returned an empty queue")
	}

	current := q.GetStatus().GetState()
	if drain {
		if current == queue.QueueState_QUEUE_STATE_ACTIVE {
			resp, err := r.conn.UpdateQueueState(ctx, &queue.UpdateQueueStateRequest{
				Id:    r.queueId(name),
				State: queue.QueueState_QUEUE_STATE_DRAINING,
			})
			if err != nil {
				return nil, fmt.Errorf("unable to drain queue, got error: %w", err)
			}
			return resp.GetQueue(), nil
		}
		return q, nil
	}

	if current == queue.QueueState_QUEUE_STATE_DRAINING || current == queue.QueueState_QUEUE_STATE_DRAINED {
		resp, err := r.conn.UpdateQueueState(ctx, &queue.UpdateQueueStateRequest{
			Id:    r.queueId(name),
			State: queue.QueueState_QUEUE_STATE_ACTIVE,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to activate queue, got error: %w", err)
		}
		return resp.GetQueue(), nil
	}

	return q, nil
}

func (r *QueueResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// queueId builds the identifier for a queue. Queues are organization-scoped: the control plane
// currently rejects any identifier carrying a project or domain.
func (r *QueueResource) queueId(name string) *common.QueueIdentifier {
	return &common.QueueIdentifier{
		Organization: r.org,
		Name:         name,
	}
}
