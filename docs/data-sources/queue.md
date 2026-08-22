---
page_title: "unionai_queue Data Source - terraform-provider-unionai"
subcategory: ""
description: |-
  Reads a Union.ai queue.
---

# unionai_queue (Data Source)

Reads a Union.ai queue by name. Use this to reference an existing queue's configuration, its lifecycle state, or the clusters it can currently route to, without managing the queue with Terraform.

## Example Usage

```terraform
data "unionai_queue" "batch" {
  id = "batch"
}

output "batch_queue_state" {
  value = data.unionai_queue.batch.state
}

output "batch_queue_clusters" {
  value = data.unionai_queue.batch.available_clusters
}
```

## Schema

### Required

- `id` (String) Queue name. Queues are organization-scoped, so the name alone identifies the queue.

### Read-Only

- `cluster_pool_name` (String) Cluster pool this queue routes into.
- `clusters` (Set of String) Clusters this queue routes to. `["*"]` means every enabled and healthy cluster in the pool.
- `run_concurrency` (Number) Maximum number of runs that may be active simultaneously. `0` means unlimited.
- `action_concurrency` (Number) Maximum number of actions (tasks) that may be in flight simultaneously. `0` means unlimited.
- `depth` (Number) Maximum total items the queue can hold, in-progress plus waiting. `0` means unlimited.
- `priority` (String) Scheduling priority relative to other queues: `min`, `medium`, or `max`.
- `fairness` (String) Cross-project scheduling policy within this queue: `round_robin` or `shuffle_interleave`.
- `state` (String) Queue lifecycle state: `active`, `draining`, or `drained`.
- `cluster_managed` (Boolean) Whether this queue is the implicit queue owned by a cluster with the same name.
- `available_clusters` (Set of String) Names of the clusters this queue can currently route to, after filtering down to those that are active and healthy.
- `created_at` (String) Creation timestamp, RFC 3339.
- `updated_at` (String) Last update timestamp, RFC 3339.
- `deleted_at` (String) Soft-deletion timestamp, RFC 3339. Null for live queues.
