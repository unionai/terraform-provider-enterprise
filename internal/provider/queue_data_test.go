package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/unionai/cloud/gen/pb-go/queue"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestQueueDataSource_Metadata(t *testing.T) {
	d := NewQueueDataSource()
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "unionai"}, resp)

	if resp.TypeName != "unionai_queue" {
		t.Errorf("Expected type name 'unionai_queue', got '%s'", resp.TypeName)
	}
}

func TestQueueDataSource_Schema(t *testing.T) {
	s := queueDataSourceSchema(t)
	if s.Diagnostics.HasError() {
		t.Fatalf("Schema() returned errors: %v", s.Diagnostics.Errors())
	}

	for _, attr := range []string{
		"id", "cluster_pool_name", "clusters", "run_concurrency", "action_concurrency",
		"depth", "priority", "fairness", "state", "cluster_managed", "available_clusters",
		"created_at", "updated_at", "deleted_at",
	} {
		if _, ok := s.Schema.Attributes[attr]; !ok {
			t.Errorf("Expected '%s' attribute in schema", attr)
		}
	}

	if !s.Schema.Attributes["id"].IsRequired() {
		t.Error("Expected 'id' to be the required lookup key")
	}
}

func TestQueueDataSource_Read(t *testing.T) {
	d := &QueueDataSource{
		org: "acme",
		conn: &mockQueueClient{
			getFn: func(ctx context.Context, req *queue.GetQueueRequest) (*queue.GetQueueResponse, error) {
				if req.GetId().GetName() != "batch" || req.GetId().GetOrganization() != "acme" {
					t.Errorf("unexpected lookup: %+v", req.GetId())
				}
				return &queue.GetQueueResponse{Queue: sampleQueue("batch", queue.QueueState_QUEUE_STATE_DRAINING)}, nil
			},
		},
	}

	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: queueDataSourceSchema(t).Schema}}
	d.Read(context.Background(), datasource.ReadRequest{Config: newQueueDataSourceConfig(t, "batch")}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Read() errors: %v", resp.Diagnostics.Errors())
	}

	var data QueueDataSourceModel
	if diags := resp.State.Get(context.Background(), &data); diags.HasError() {
		t.Fatalf("State.Get() errors: %v", diags.Errors())
	}
	if data.State.ValueString() != "draining" {
		t.Errorf("expected state 'draining', got %q", data.State.ValueString())
	}
	if data.Priority.ValueString() != "max" {
		t.Errorf("expected priority 'max', got %q", data.Priority.ValueString())
	}
	if data.ClusterPoolName.ValueString() != "default" {
		t.Errorf("expected cluster pool 'default', got %q", data.ClusterPoolName.ValueString())
	}
	if len(data.AvailableClusters.Elements()) != 1 {
		t.Errorf("expected one available cluster, got %d", len(data.AvailableClusters.Elements()))
	}
}

func TestQueueDataSource_ReadNotFoundErrors(t *testing.T) {
	d := &QueueDataSource{
		org: "acme",
		conn: &mockQueueClient{
			getFn: func(ctx context.Context, req *queue.GetQueueRequest) (*queue.GetQueueResponse, error) {
				return nil, status.Error(codes.NotFound, "missing")
			},
		},
	}

	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: queueDataSourceSchema(t).Schema}}
	d.Read(context.Background(), datasource.ReadRequest{Config: newQueueDataSourceConfig(t, "batch")}, resp)

	// Unlike a resource, a missing data source is an error rather than a silent state removal.
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error when the queue does not exist")
	}
}

// --- helpers -------------------------------------------------------------

func queueDataSourceSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := NewQueueDataSource()
	resp := datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

func queueDataSourceObjectType() tftypes.Object {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                 tftypes.String,
			"cluster_pool_name":  tftypes.String,
			"clusters":           tftypes.Set{ElementType: tftypes.String},
			"run_concurrency":    tftypes.Number,
			"action_concurrency": tftypes.Number,
			"depth":              tftypes.Number,
			"priority":           tftypes.String,
			"fairness":           tftypes.String,
			"state":              tftypes.String,
			"cluster_managed":    tftypes.Bool,
			"available_clusters": tftypes.Set{ElementType: tftypes.String},
			"created_at":         tftypes.String,
			"updated_at":         tftypes.String,
			"deleted_at":         tftypes.String,
		},
	}
}

func newQueueDataSourceConfig(t *testing.T, name string) tfsdk.Config {
	t.Helper()
	nullSet := tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil)
	return tfsdk.Config{
		Schema: queueDataSourceSchema(t).Schema,
		Raw: tftypes.NewValue(queueDataSourceObjectType(), map[string]tftypes.Value{
			"id":                 tftypes.NewValue(tftypes.String, name),
			"cluster_pool_name":  tftypes.NewValue(tftypes.String, nil),
			"clusters":           nullSet,
			"run_concurrency":    tftypes.NewValue(tftypes.Number, nil),
			"action_concurrency": tftypes.NewValue(tftypes.Number, nil),
			"depth":              tftypes.NewValue(tftypes.Number, nil),
			"priority":           tftypes.NewValue(tftypes.String, nil),
			"fairness":           tftypes.NewValue(tftypes.String, nil),
			"state":              tftypes.NewValue(tftypes.String, nil),
			"cluster_managed":    tftypes.NewValue(tftypes.Bool, nil),
			"available_clusters": nullSet,
			"created_at":         tftypes.NewValue(tftypes.String, nil),
			"updated_at":         tftypes.NewValue(tftypes.String, nil),
			"deleted_at":         tftypes.NewValue(tftypes.String, nil),
		}),
	}
}
