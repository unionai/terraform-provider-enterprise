package provider

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/unionai/cloud/gen/pb-go/common"
	"github.com/unionai/cloud/gen/pb-go/queue"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// defaultClusterPoolName mirrors clusterpool.DefaultPool in the control plane. This pool is
// created on demand; every other pool must already exist before a queue can reference it.
const defaultClusterPoolName = "default"

// wildcardCluster routes a queue to every enabled and healthy cluster in its pool. When present
// it must be the only entry in the cluster list.
const wildcardCluster = "*"

// queueNameReservedNote spells out the consequence shared by every destroy path where the queue
// outlives the Terraform resource: the name stays taken, so recreating it needs an import.
const queueNameReservedNote = "Its name remains reserved, so a later apply of the same configuration will fail with an " +
	"\"already exists\" error — use `terraform import` to adopt it again."

const (
	priorityPrefix = "PRIORITY_"
	fairnessPrefix = "FAIRNESS_ALGORITHM_"
	statePrefix    = "QUEUE_STATE_"
)

// The provider spells enums as lowercase, prefix-stripped strings, matching how roles spell
// their actions.

func queuePriorityFromString(s string) (queue.Priority, bool) {
	v, ok := queue.Priority_value[priorityPrefix+strings.ToUpper(s)]
	if !ok || v == int32(queue.Priority_PRIORITY_UNSPECIFIED) {
		return queue.Priority_PRIORITY_UNSPECIFIED, false
	}
	return queue.Priority(v), true
}

func queueFairnessFromString(s string) (queue.FairnessAlgorithm, bool) {
	v, ok := queue.FairnessAlgorithm_value[fairnessPrefix+strings.ToUpper(s)]
	if !ok || v == int32(queue.FairnessAlgorithm_FAIRNESS_ALGORITHM_UNSPECIFIED) {
		return queue.FairnessAlgorithm_FAIRNESS_ALGORITHM_UNSPECIFIED, false
	}
	return queue.FairnessAlgorithm(v), true
}

// enumToTerraform strips the given prefix and lowercases an enum name. Unspecified and
// unrecognized values become null rather than an empty string.
func enumToTerraform(name string, prefix string) types.String {
	if name == "" || name == prefix+"UNSPECIFIED" {
		return types.StringNull()
	}
	return types.StringValue(strings.ToLower(strings.TrimPrefix(name, prefix)))
}

func queuePriorityToTerraform(p queue.Priority) types.String {
	return enumToTerraform(queue.Priority_name[int32(p)], priorityPrefix)
}

func queueFairnessToTerraform(f queue.FairnessAlgorithm) types.String {
	return enumToTerraform(queue.FairnessAlgorithm_name[int32(f)], fairnessPrefix)
}

func queueStateToTerraform(s queue.QueueState) types.String {
	return enumToTerraform(queue.QueueState_name[int32(s)], statePrefix)
}

func timestampToTerraform(ts *timestamppb.Timestamp) types.String {
	if ts == nil {
		return types.StringNull()
	}
	return types.StringValue(ts.AsTime().UTC().Format(time.RFC3339))
}

// queueCountToUint32 converts one of the concurrency/depth knobs, all of which are uint32 on
// the wire and use 0 to mean unlimited.
func queueCountToUint32(v types.Int64, attribute string) (uint32, diag.Diagnostics) {
	var diags diag.Diagnostics
	if v.IsNull() || v.IsUnknown() {
		return 0, diags
	}
	n := v.ValueInt64()
	if n < 0 || n > math.MaxUint32 {
		diags.AddAttributeError(
			path.Root(attribute),
			"Value out of range",
			fmt.Sprintf("%s must be between 0 and %d, got %d.", attribute, uint32(math.MaxUint32), n),
		)
		return 0, diags
	}
	return uint32(n), diags
}

// buildQueueSpec assembles the complete spec from the plan. UpdateQueue replaces the spec
// wholesale rather than merging, so create and update both send every field.
func buildQueueSpec(ctx context.Context, data *QueueResourceModel) (*queue.QueueSpec, diag.Diagnostics) {
	var diags diag.Diagnostics

	runConcurrency, d := queueCountToUint32(data.RunConcurrency, "run_concurrency")
	diags.Append(d...)
	actionConcurrency, d := queueCountToUint32(data.ActionConcurrency, "action_concurrency")
	diags.Append(d...)
	depth, d := queueCountToUint32(data.Depth, "depth")
	diags.Append(d...)

	priority, ok := queuePriorityFromString(data.Priority.ValueString())
	if !ok {
		diags.AddAttributeError(
			path.Root("priority"),
			"Invalid priority",
			fmt.Sprintf("Unknown queue priority %q. Valid values are: min, medium, max.", data.Priority.ValueString()),
		)
	}

	fairness, ok := queueFairnessFromString(data.Fairness.ValueString())
	if !ok {
		diags.AddAttributeError(
			path.Root("fairness"),
			"Invalid fairness algorithm",
			fmt.Sprintf("Unknown queue fairness algorithm %q. Valid values are: round_robin, shuffle_interleave.", data.Fairness.ValueString()),
		)
	}

	var clusters []string
	if !data.Clusters.IsNull() && !data.Clusters.IsUnknown() {
		diags.Append(data.Clusters.ElementsAs(ctx, &clusters, false)...)
	}
	for _, c := range clusters {
		if c == wildcardCluster && len(clusters) != 1 {
			diags.AddAttributeError(
				path.Root("clusters"),
				"Invalid cluster list",
				fmt.Sprintf("When %q is used it must be the only entry in clusters, got %d entries.", wildcardCluster, len(clusters)),
			)
			break
		}
	}

	if diags.HasError() {
		return nil, diags
	}

	return &queue.QueueSpec{
		RunConcurrency:    runConcurrency,
		ActionConcurrency: actionConcurrency,
		Depth:             depth,
		Priority:          priority,
		Fairness:          fairness,
		Clusters:          clusters,
		ClusterPoolName:   data.ClusterPoolName.ValueString(),
	}, diags
}

// applyQueueToModel copies a server-returned queue back over the model, so that computed
// attributes and any server-side normalization land in state.
func applyQueueToModel(data *QueueResourceModel, q *queue.Queue) diag.Diagnostics {
	var diags diag.Diagnostics

	if q == nil {
		diags.AddError("Client Error", "The control plane returned an empty queue.")
		return diags
	}

	data.Id = types.StringValue(q.GetId().GetName())
	data.Name = types.StringValue(q.GetId().GetName())

	spec := q.GetSpec()
	data.ClusterPoolName = types.StringValue(spec.GetClusterPoolName())
	data.Clusters = convertStringsToSet(spec.GetClusters())
	data.RunConcurrency = types.Int64Value(int64(spec.GetRunConcurrency()))
	data.ActionConcurrency = types.Int64Value(int64(spec.GetActionConcurrency()))
	data.Depth = types.Int64Value(int64(spec.GetDepth()))
	data.Priority = queuePriorityToTerraform(spec.GetPriority())
	data.Fairness = queueFairnessToTerraform(spec.GetFairness())
	if data.Drain.IsNull() || data.Drain.IsUnknown() {
		data.Drain = types.BoolValue(false)
	}

	data.State = queueStateToTerraform(q.GetStatus().GetState())
	data.ClusterManaged = types.BoolValue(q.GetStatus().GetClusterManaged())
	data.AvailableClusters = convertArrayToSetGetter(
		q.GetStatus().GetAvailableClusters(),
		func(c *common.ClusterIdentifier) string { return c.GetName() },
	)

	data.CreatedAt = timestampToTerraform(q.GetCreatedAt())
	data.UpdatedAt = timestampToTerraform(q.GetUpdatedAt())
	data.DeletedAt = timestampToTerraform(q.GetDeletedAt())

	return diags
}
