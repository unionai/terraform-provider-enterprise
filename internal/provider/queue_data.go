package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/unionai/cloud/gen/pb-go/common"
	"github.com/unionai/cloud/gen/pb-go/queue"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &QueueDataSource{}

func NewQueueDataSource() datasource.DataSource {
	return &QueueDataSource{}
}

// QueueDataSource defines the data source implementation.
type QueueDataSource struct {
	conn queue.QueueServiceClient
	org  string
}

// QueueDataSourceModel describes the data source data model.
type QueueDataSourceModel struct {
	Id                types.String `tfsdk:"id"`
	ClusterPoolName   types.String `tfsdk:"cluster_pool_name"`
	Clusters          types.Set    `tfsdk:"clusters"`
	RunConcurrency    types.Int64  `tfsdk:"run_concurrency"`
	ActionConcurrency types.Int64  `tfsdk:"action_concurrency"`
	Depth             types.Int64  `tfsdk:"depth"`
	Priority          types.String `tfsdk:"priority"`
	Fairness          types.String `tfsdk:"fairness"`
	State             types.String `tfsdk:"state"`
	ClusterManaged    types.Bool   `tfsdk:"cluster_managed"`
	AvailableClusters types.Set    `tfsdk:"available_clusters"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
	DeletedAt         types.String `tfsdk:"deleted_at"`
}

func (d *QueueDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_queue"
}

func (d *QueueDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Queue data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Queue name",
			},
			"cluster_pool_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cluster pool this queue routes into",
			},
			"clusters": schema.SetAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Clusters this queue routes to. `[\"*\"]` means every enabled and healthy cluster in the pool.",
			},
			"run_concurrency": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Maximum number of runs that may be active simultaneously. `0` means unlimited.",
			},
			"action_concurrency": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Maximum number of actions (tasks) that may be in flight simultaneously. `0` means unlimited.",
			},
			"depth": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Maximum total items the queue can hold (in-progress plus waiting). `0` means unlimited.",
			},
			"priority": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Scheduling priority relative to other queues: `min`, `medium`, or `max`.",
			},
			"fairness": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cross-project scheduling policy within this queue: `round_robin` or `shuffle_interleave`.",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Queue lifecycle state: `active`, `draining`, or `drained`.",
			},
			"cluster_managed": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether this queue is the implicit queue owned by a cluster with the same name.",
			},
			"available_clusters": schema.SetAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Names of the clusters this queue can currently route to, after filtering down to those that are active and healthy.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation timestamp, RFC 3339.",
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

func (d *QueueDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.conn = queue.NewQueueServiceClient(client.conn)
	d.org = client.org
}

func (d *QueueDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data QueueDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Id.ValueString()
	got, err := d.conn.GetQueue(ctx, &queue.GetQueueRequest{
		Id: &common.QueueIdentifier{Organization: d.org, Name: name},
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			resp.Diagnostics.AddError("Queue not found", fmt.Sprintf("Queue with name %s not found", name))
			return
		}
		resp.Diagnostics.AddError("Failed to fetch queue", err.Error())
		return
	}
	tflog.Trace(ctx, "GetQueue response", map[string]interface{}{"queue": got.GetQueue()})

	q := got.GetQueue()
	spec := q.GetSpec()

	data.Id = types.StringValue(q.GetId().GetName())
	data.ClusterPoolName = types.StringValue(spec.GetClusterPoolName())
	data.Clusters = convertStringsToSet(spec.GetClusters())
	data.RunConcurrency = types.Int64Value(int64(spec.GetRunConcurrency()))
	data.ActionConcurrency = types.Int64Value(int64(spec.GetActionConcurrency()))
	data.Depth = types.Int64Value(int64(spec.GetDepth()))
	data.Priority = queuePriorityToTerraform(spec.GetPriority())
	data.Fairness = queueFairnessToTerraform(spec.GetFairness())
	data.State = queueStateToTerraform(q.GetStatus().GetState())
	data.ClusterManaged = types.BoolValue(q.GetStatus().GetClusterManaged())
	data.AvailableClusters = convertArrayToSetGetter(
		q.GetStatus().GetAvailableClusters(),
		func(c *common.ClusterIdentifier) string { return c.GetName() },
	)
	data.CreatedAt = timestampToTerraform(q.GetCreatedAt())
	data.UpdatedAt = timestampToTerraform(q.GetUpdatedAt())
	data.DeletedAt = timestampToTerraform(q.GetDeletedAt())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
