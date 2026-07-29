package provider

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/unionai/cloud/gen/pb-go/common"
	"github.com/unionai/cloud/gen/pb-go/queue"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// mockQueueClient implements the subset of queue.QueueServiceClient used by QueueResource and
// QueueDataSource. The embedded interface satisfies the rest (ListQueues, WatchQueueMetrics,
// ResolveQueue), which the provider never calls.
type mockQueueClient struct {
	queue.QueueServiceClient
	createFn      func(ctx context.Context, req *queue.CreateQueueRequest) (*queue.CreateQueueResponse, error)
	getFn         func(ctx context.Context, req *queue.GetQueueRequest) (*queue.GetQueueResponse, error)
	updateFn      func(ctx context.Context, req *queue.UpdateQueueRequest) (*queue.UpdateQueueResponse, error)
	updateStateFn func(ctx context.Context, req *queue.UpdateQueueStateRequest) (*queue.UpdateQueueStateResponse, error)
}

func (m *mockQueueClient) CreateQueue(ctx context.Context, in *queue.CreateQueueRequest, opts ...grpc.CallOption) (*queue.CreateQueueResponse, error) {
	return m.createFn(ctx, in)
}

func (m *mockQueueClient) GetQueue(ctx context.Context, in *queue.GetQueueRequest, opts ...grpc.CallOption) (*queue.GetQueueResponse, error) {
	return m.getFn(ctx, in)
}

func (m *mockQueueClient) UpdateQueue(ctx context.Context, in *queue.UpdateQueueRequest, opts ...grpc.CallOption) (*queue.UpdateQueueResponse, error) {
	return m.updateFn(ctx, in)
}

func (m *mockQueueClient) UpdateQueueState(ctx context.Context, in *queue.UpdateQueueStateRequest, opts ...grpc.CallOption) (*queue.UpdateQueueStateResponse, error) {
	return m.updateStateFn(ctx, in)
}

// sampleQueue is a server response for a queue in the shape the tests configure.
func sampleQueue(name string, state queue.QueueState) *queue.Queue {
	return &queue.Queue{
		Id: &common.QueueIdentifier{Organization: "acme", Name: name},
		Spec: &queue.QueueSpec{
			RunConcurrency:    5,
			ActionConcurrency: 0,
			Depth:             100,
			Priority:          queue.Priority_PRIORITY_MAX,
			Fairness:          queue.FairnessAlgorithm_FAIRNESS_ALGORITHM_ROUND_ROBIN,
			Clusters:          []string{"*"},
			ClusterPoolName:   "default",
		},
		Status: &queue.QueueStatus{
			State:             state,
			AvailableClusters: []*common.ClusterIdentifier{{Organization: "acme", Name: "cluster-a"}},
		},
		CreatedAt: timestamppb.New(mustTime("2026-01-02T03:04:05Z")),
		UpdatedAt: timestamppb.New(mustTime("2026-01-02T03:04:05Z")),
	}
}

func TestQueueResource_Metadata(t *testing.T) {
	r := NewQueueResource()
	req := resource.MetadataRequest{ProviderTypeName: "unionai"}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "unionai_queue" {
		t.Errorf("Expected type name 'unionai_queue', got '%s'", resp.TypeName)
	}
}

func TestQueueResource_Schema(t *testing.T) {
	s := queueSchema(t)
	if s.Diagnostics.HasError() {
		t.Fatalf("Schema() returned errors: %v", s.Diagnostics.Errors())
	}

	for _, attr := range []string{
		"id", "name", "cluster_pool_name", "clusters", "run_concurrency",
		"action_concurrency", "depth", "priority", "fairness", "state",
		"available_clusters", "created_at", "updated_at",
	} {
		if _, ok := s.Schema.Attributes[attr]; !ok {
			t.Errorf("Expected '%s' attribute in schema", attr)
		}
	}

	// The server owns these; a practitioner must not be able to set them.
	for _, attr := range []string{"state", "available_clusters", "created_at", "updated_at"} {
		a := s.Schema.Attributes[attr]
		if !a.IsComputed() {
			t.Errorf("Expected '%s' to be computed", attr)
		}
		if a.IsOptional() || a.IsRequired() {
			t.Errorf("Expected '%s' to be read-only, but it is settable", attr)
		}
	}
}

func TestQueueResource_Create(t *testing.T) {
	var captured *queue.CreateQueueRequest
	r := &QueueResource{
		org: "acme",
		conn: &mockQueueClient{
			createFn: func(ctx context.Context, req *queue.CreateQueueRequest) (*queue.CreateQueueResponse, error) {
				captured = req
				return &queue.CreateQueueResponse{Queue: sampleQueue("batch", queue.QueueState_QUEUE_STATE_ACTIVE)}, nil
			},
		},
	}

	resp := &resource.CreateResponse{State: tfsdk.State{Schema: queueSchema(t).Schema}}
	r.Create(context.Background(), resource.CreateRequest{
		Plan: newQueuePlan(t, queuePlanValues{
			name: "batch", pool: "default", clusters: []string{"*"},
			runConcurrency: 5, depth: 100, priority: "max", fairness: "round_robin",
		}),
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create() errors: %v", resp.Diagnostics.Errors())
	}
	if captured == nil {
		t.Fatal("CreateQueue was not called")
	}

	// Queues are org-scoped: the control plane rejects an identifier carrying project or domain.
	id := captured.GetId()
	if id.GetOrganization() != "acme" || id.GetName() != "batch" {
		t.Errorf("unexpected queue identifier: %+v", id)
	}
	if id.GetProject() != "" || id.GetDomain() != "" {
		t.Errorf("expected empty project/domain, got project=%q domain=%q", id.GetProject(), id.GetDomain())
	}

	if captured.GetState() != queue.QueueState_QUEUE_STATE_ACTIVE {
		t.Errorf("expected create state ACTIVE, got %v", captured.GetState())
	}

	spec := captured.GetSpec()
	if spec.GetPriority() != queue.Priority_PRIORITY_MAX {
		t.Errorf("expected PRIORITY_MAX, got %v", spec.GetPriority())
	}
	if spec.GetFairness() != queue.FairnessAlgorithm_FAIRNESS_ALGORITHM_ROUND_ROBIN {
		t.Errorf("expected ROUND_ROBIN, got %v", spec.GetFairness())
	}
	if spec.GetClusterPoolName() != "default" {
		t.Errorf("expected cluster pool 'default', got %q", spec.GetClusterPoolName())
	}
	if spec.GetRunConcurrency() != 5 || spec.GetDepth() != 100 || spec.GetActionConcurrency() != 0 {
		t.Errorf("unexpected concurrency knobs: %+v", spec)
	}
}

func TestQueueResource_CreateAlreadyExists(t *testing.T) {
	r := &QueueResource{
		org: "acme",
		conn: &mockQueueClient{
			createFn: func(ctx context.Context, req *queue.CreateQueueRequest) (*queue.CreateQueueResponse, error) {
				return nil, status.Error(codes.AlreadyExists, "queue already exists")
			},
		},
	}

	resp := &resource.CreateResponse{State: tfsdk.State{Schema: queueSchema(t).Schema}}
	r.Create(context.Background(), resource.CreateRequest{
		Plan: newQueuePlan(t, queuePlanValues{name: "batch", pool: "default", clusters: []string{"*"}, priority: "medium", fairness: "round_robin"}),
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error when the queue already exists")
	}
	// Names stay reserved after a drain, so the recovery path is import.
	if detail := resp.Diagnostics.Errors()[0].Detail(); !strings.Contains(detail, "terraform import") {
		t.Errorf("expected the error to point at terraform import, got: %s", detail)
	}
}

func TestQueueResource_CreateRejectsBadEnum(t *testing.T) {
	r := &QueueResource{
		org: "acme",
		conn: &mockQueueClient{
			createFn: func(ctx context.Context, req *queue.CreateQueueRequest) (*queue.CreateQueueResponse, error) {
				t.Fatal("CreateQueue should not be called when the plan is invalid")
				return nil, nil
			},
		},
	}

	resp := &resource.CreateResponse{State: tfsdk.State{Schema: queueSchema(t).Schema}}
	r.Create(context.Background(), resource.CreateRequest{
		Plan: newQueuePlan(t, queuePlanValues{name: "batch", pool: "default", clusters: []string{"*"}, priority: "urgent", fairness: "round_robin"}),
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error for an unknown priority")
	}
}

func TestQueueResource_CreateRejectsWildcardWithOthers(t *testing.T) {
	r := &QueueResource{
		org: "acme",
		conn: &mockQueueClient{
			createFn: func(ctx context.Context, req *queue.CreateQueueRequest) (*queue.CreateQueueResponse, error) {
				t.Fatal("CreateQueue should not be called when the plan is invalid")
				return nil, nil
			},
		},
	}

	resp := &resource.CreateResponse{State: tfsdk.State{Schema: queueSchema(t).Schema}}
	r.Create(context.Background(), resource.CreateRequest{
		Plan: newQueuePlan(t, queuePlanValues{
			name: "batch", pool: "default", clusters: []string{"*", "cluster-a"},
			priority: "medium", fairness: "round_robin",
		}),
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error when '*' is combined with named clusters")
	}
}

func TestQueueResource_ReadNotFoundRemovesState(t *testing.T) {
	r := &QueueResource{
		org: "acme",
		conn: &mockQueueClient{
			getFn: func(ctx context.Context, req *queue.GetQueueRequest) (*queue.GetQueueResponse, error) {
				return nil, status.Error(codes.NotFound, "missing")
			},
		},
	}

	resp := &resource.ReadResponse{State: newQueueState(t, "batch")}
	r.Read(context.Background(), resource.ReadRequest{State: newQueueState(t, "batch")}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Read() errors: %v", resp.Diagnostics.Errors())
	}
	if !resp.State.Raw.IsNull() {
		t.Error("expected the resource to be removed from state when the queue is gone")
	}
}

func TestQueueResource_ReadPopulatesName(t *testing.T) {
	r := &QueueResource{
		org: "acme",
		conn: &mockQueueClient{
			getFn: func(ctx context.Context, req *queue.GetQueueRequest) (*queue.GetQueueResponse, error) {
				return &queue.GetQueueResponse{Queue: sampleQueue("batch", queue.QueueState_QUEUE_STATE_ACTIVE)}, nil
			},
		},
	}

	resp := &resource.ReadResponse{State: newQueueState(t, "batch")}
	r.Read(context.Background(), resource.ReadRequest{State: newQueueState(t, "batch")}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Read() errors: %v", resp.Diagnostics.Errors())
	}

	var data QueueResourceModel
	if diags := resp.State.Get(context.Background(), &data); diags.HasError() {
		t.Fatalf("State.Get() errors: %v", diags.Errors())
	}
	// ImportState only seeds id, so Read must fill in name for import to converge.
	if data.Name.ValueString() != "batch" {
		t.Errorf("expected name 'batch', got %q", data.Name.ValueString())
	}
	if data.State.ValueString() != "active" {
		t.Errorf("expected state 'active', got %q", data.State.ValueString())
	}
	if data.Priority.ValueString() != "max" {
		t.Errorf("expected priority 'max', got %q", data.Priority.ValueString())
	}
	if data.CreatedAt.ValueString() != "2026-01-02T03:04:05Z" {
		t.Errorf("unexpected created_at: %q", data.CreatedAt.ValueString())
	}
}

func TestQueueResource_UpdateSendsClusterPool(t *testing.T) {
	var captured *queue.UpdateQueueRequest
	r := &QueueResource{
		org: "acme",
		conn: &mockQueueClient{
			updateFn: func(ctx context.Context, req *queue.UpdateQueueRequest) (*queue.UpdateQueueResponse, error) {
				captured = req
				return &queue.UpdateQueueResponse{Queue: sampleQueue("batch", queue.QueueState_QUEUE_STATE_ACTIVE)}, nil
			},
		},
	}

	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: queueSchema(t).Schema}}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan: newQueuePlan(t, queuePlanValues{
			name: "batch", pool: "default", clusters: []string{"*"},
			runConcurrency: 5, depth: 100, priority: "max", fairness: "round_robin",
		}),
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Update() errors: %v", resp.Diagnostics.Errors())
	}
	// UpdateQueue rejects an empty cluster pool, and replaces the spec wholesale.
	if captured.GetSpec().GetClusterPoolName() == "" {
		t.Error("expected UpdateQueue to carry a non-empty cluster_pool_name")
	}
	if captured.GetSpec().GetDepth() != 100 {
		t.Errorf("expected the full spec to be sent, got depth %d", captured.GetSpec().GetDepth())
	}
}

func TestQueueResource_ModifyPlanRejectsPoolChange(t *testing.T) {
	r := &QueueResource{org: "acme"}

	resp := &resource.ModifyPlanResponse{}
	r.ModifyPlan(context.Background(), resource.ModifyPlanRequest{
		State: newQueueStateWithPool(t, "batch", "default"),
		Plan: newQueuePlan(t, queuePlanValues{
			name: "batch", pool: "gpu-pool", clusters: []string{"*"},
			priority: "medium", fairness: "round_robin",
		}),
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error when cluster_pool_name changes")
	}
}

func TestQueueResource_ModifyPlanAllowsSamePool(t *testing.T) {
	r := &QueueResource{org: "acme"}

	resp := &resource.ModifyPlanResponse{}
	r.ModifyPlan(context.Background(), resource.ModifyPlanRequest{
		State: newQueueStateWithPool(t, "batch", "default"),
		Plan: newQueuePlan(t, queuePlanValues{
			name: "batch", pool: "default", clusters: []string{"*"},
			priority: "medium", fairness: "round_robin",
		}),
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("ModifyPlan() errors: %v", resp.Diagnostics.Errors())
	}
}

func TestQueueResource_DeleteDrains(t *testing.T) {
	var captured *queue.UpdateQueueStateRequest
	r := &QueueResource{
		org: "acme",
		conn: &mockQueueClient{
			getFn: func(ctx context.Context, req *queue.GetQueueRequest) (*queue.GetQueueResponse, error) {
				return &queue.GetQueueResponse{Queue: sampleQueue("batch", queue.QueueState_QUEUE_STATE_ACTIVE)}, nil
			},
			updateStateFn: func(ctx context.Context, req *queue.UpdateQueueStateRequest) (*queue.UpdateQueueStateResponse, error) {
				captured = req
				return &queue.UpdateQueueStateResponse{Queue: sampleQueue("batch", queue.QueueState_QUEUE_STATE_DRAINING)}, nil
			},
		},
	}

	resp := &resource.DeleteResponse{State: newQueueState(t, "batch")}
	r.Delete(context.Background(), resource.DeleteRequest{State: newQueueState(t, "batch")}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Delete() errors: %v", resp.Diagnostics.Errors())
	}
	if captured == nil {
		t.Fatal("UpdateQueueState was not called")
	}
	if captured.GetState() != queue.QueueState_QUEUE_STATE_DRAINING {
		t.Errorf("expected DRAINING, got %v", captured.GetState())
	}
	if resp.Diagnostics.WarningsCount() == 0 {
		t.Error("expected a warning explaining the queue was drained rather than deleted")
	}
}

func TestQueueResource_DeleteAlreadyDrainedIsNoOp(t *testing.T) {
	r := &QueueResource{
		org: "acme",
		conn: &mockQueueClient{
			getFn: func(ctx context.Context, req *queue.GetQueueRequest) (*queue.GetQueueResponse, error) {
				return &queue.GetQueueResponse{Queue: sampleQueue("batch", queue.QueueState_QUEUE_STATE_DRAINED)}, nil
			},
			updateStateFn: func(ctx context.Context, req *queue.UpdateQueueStateRequest) (*queue.UpdateQueueStateResponse, error) {
				t.Fatal("UpdateQueueState should not be called for an already drained queue")
				return nil, nil
			},
		},
	}

	resp := &resource.DeleteResponse{State: newQueueState(t, "batch")}
	r.Delete(context.Background(), resource.DeleteRequest{State: newQueueState(t, "batch")}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Delete() errors: %v", resp.Diagnostics.Errors())
	}
}

func TestQueueResource_DeleteNotFoundIsNoOp(t *testing.T) {
	r := &QueueResource{
		org: "acme",
		conn: &mockQueueClient{
			getFn: func(ctx context.Context, req *queue.GetQueueRequest) (*queue.GetQueueResponse, error) {
				return nil, status.Error(codes.NotFound, "missing")
			},
		},
	}

	resp := &resource.DeleteResponse{State: newQueueState(t, "batch")}
	r.Delete(context.Background(), resource.DeleteRequest{State: newQueueState(t, "batch")}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Delete() errors: %v", resp.Diagnostics.Errors())
	}
}

func TestQueueResource_DeleteFailedPreconditionExplains(t *testing.T) {
	r := &QueueResource{
		org: "acme",
		conn: &mockQueueClient{
			getFn: func(ctx context.Context, req *queue.GetQueueRequest) (*queue.GetQueueResponse, error) {
				return &queue.GetQueueResponse{Queue: sampleQueue("batch", queue.QueueState_QUEUE_STATE_ACTIVE)}, nil
			},
			updateStateFn: func(ctx context.Context, req *queue.UpdateQueueStateRequest) (*queue.UpdateQueueStateResponse, error) {
				return nil, status.Error(codes.FailedPrecondition, "cannot drain queue [batch]: it is configured as run.default_queue")
			},
		},
	}

	resp := &resource.DeleteResponse{State: newQueueState(t, "batch")}
	r.Delete(context.Background(), resource.DeleteRequest{State: newQueueState(t, "batch")}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error when the control plane refuses to drain")
	}
	if detail := resp.Diagnostics.Errors()[0].Detail(); !strings.Contains(detail, "run.default_queue") {
		t.Errorf("expected the error to mention run.default_queue, got: %s", detail)
	}
}

func TestQueueEnumRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		in  string
		out queue.Priority
	}{
		{"min", queue.Priority_PRIORITY_MIN},
		{"medium", queue.Priority_PRIORITY_MEDIUM},
		{"max", queue.Priority_PRIORITY_MAX},
	} {
		got, ok := queuePriorityFromString(tc.in)
		if !ok || got != tc.out {
			t.Errorf("queuePriorityFromString(%q) = %v, %v", tc.in, got, ok)
		}
		if back := queuePriorityToTerraform(got); back.ValueString() != tc.in {
			t.Errorf("round trip for %q produced %q", tc.in, back.ValueString())
		}
	}

	for _, tc := range []struct {
		in  string
		out queue.FairnessAlgorithm
	}{
		{"round_robin", queue.FairnessAlgorithm_FAIRNESS_ALGORITHM_ROUND_ROBIN},
		{"shuffle_interleave", queue.FairnessAlgorithm_FAIRNESS_ALGORITHM_SHUFFLE_INTERLEAVE},
	} {
		got, ok := queueFairnessFromString(tc.in)
		if !ok || got != tc.out {
			t.Errorf("queueFairnessFromString(%q) = %v, %v", tc.in, got, ok)
		}
		if back := queueFairnessToTerraform(got); back.ValueString() != tc.in {
			t.Errorf("round trip for %q produced %q", tc.in, back.ValueString())
		}
	}

	if _, ok := queuePriorityFromString("urgent"); ok {
		t.Error("expected 'urgent' to be rejected as a priority")
	}
	if _, ok := queuePriorityFromString("unspecified"); ok {
		t.Error("expected 'unspecified' to be rejected as a priority")
	}
	// An unset enum has no meaningful Terraform spelling.
	if v := queuePriorityToTerraform(queue.Priority_PRIORITY_UNSPECIFIED); !v.IsNull() {
		t.Errorf("expected null for an unspecified priority, got %q", v.ValueString())
	}
	if v := queueStateToTerraform(queue.QueueState_QUEUE_STATE_DRAINING); v.ValueString() != "draining" {
		t.Errorf("expected 'draining', got %q", v.ValueString())
	}
}

// --- helpers -------------------------------------------------------------

func queueSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := NewQueueResource()
	resp := resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

func queueObjectType() tftypes.Object {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                 tftypes.String,
			"name":               tftypes.String,
			"cluster_pool_name":  tftypes.String,
			"clusters":           tftypes.Set{ElementType: tftypes.String},
			"run_concurrency":    tftypes.Number,
			"action_concurrency": tftypes.Number,
			"depth":              tftypes.Number,
			"priority":           tftypes.String,
			"fairness":           tftypes.String,
			"state":              tftypes.String,
			"available_clusters": tftypes.Set{ElementType: tftypes.String},
			"created_at":         tftypes.String,
			"updated_at":         tftypes.String,
		},
	}
}

type queuePlanValues struct {
	name              string
	pool              string
	clusters          []string
	runConcurrency    int64
	actionConcurrency int64
	depth             int64
	priority          string
	fairness          string
}

func stringSetValue(items []string) tftypes.Value {
	vals := make([]tftypes.Value, 0, len(items))
	for _, c := range items {
		vals = append(vals, tftypes.NewValue(tftypes.String, c))
	}
	return tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, vals)
}

func newQueuePlan(t *testing.T, v queuePlanValues) tfsdk.Plan {
	t.Helper()
	return tfsdk.Plan{
		Schema: queueSchema(t).Schema,
		Raw: tftypes.NewValue(queueObjectType(), map[string]tftypes.Value{
			"id":                 tftypes.NewValue(tftypes.String, nil),
			"name":               tftypes.NewValue(tftypes.String, v.name),
			"cluster_pool_name":  tftypes.NewValue(tftypes.String, v.pool),
			"clusters":           stringSetValue(v.clusters),
			"run_concurrency":    tftypes.NewValue(tftypes.Number, v.runConcurrency),
			"action_concurrency": tftypes.NewValue(tftypes.Number, v.actionConcurrency),
			"depth":              tftypes.NewValue(tftypes.Number, v.depth),
			"priority":           tftypes.NewValue(tftypes.String, v.priority),
			"fairness":           tftypes.NewValue(tftypes.String, v.fairness),
			"state":              tftypes.NewValue(tftypes.String, nil),
			"available_clusters": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
			"created_at":         tftypes.NewValue(tftypes.String, nil),
			"updated_at":         tftypes.NewValue(tftypes.String, nil),
		}),
	}
}

func newQueueState(t *testing.T, name string) tfsdk.State {
	t.Helper()
	return newQueueStateWithPool(t, name, "default")
}

func newQueueStateWithPool(t *testing.T, name, pool string) tfsdk.State {
	t.Helper()
	return tfsdk.State{
		Schema: queueSchema(t).Schema,
		Raw: tftypes.NewValue(queueObjectType(), map[string]tftypes.Value{
			"id":                 tftypes.NewValue(tftypes.String, name),
			"name":               tftypes.NewValue(tftypes.String, name),
			"cluster_pool_name":  tftypes.NewValue(tftypes.String, pool),
			"clusters":           stringSetValue([]string{"*"}),
			"run_concurrency":    tftypes.NewValue(tftypes.Number, 5),
			"action_concurrency": tftypes.NewValue(tftypes.Number, 0),
			"depth":              tftypes.NewValue(tftypes.Number, 100),
			"priority":           tftypes.NewValue(tftypes.String, "max"),
			"fairness":           tftypes.NewValue(tftypes.String, "round_robin"),
			"state":              tftypes.NewValue(tftypes.String, "active"),
			"available_clusters": stringSetValue([]string{"cluster-a"}),
			"created_at":         tftypes.NewValue(tftypes.String, "2026-01-02T03:04:05Z"),
			"updated_at":         tftypes.NewValue(tftypes.String, "2026-01-02T03:04:05Z"),
		}),
	}
}

func mustTime(s string) time.Time {
	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return ts
}
