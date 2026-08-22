---
page_title: "unionai_queue Resource - terraform-provider-unionai"
subcategory: ""
description: |-
  Manages a Union.ai queue.
---

# unionai_queue (Resource)

Manages a Union.ai queue. Queues control admission and scheduling of runs: how many runs and actions may be in flight, how deep the backlog may get, scheduling priority relative to other queues, how work from different projects is interleaved, and which clusters within a cluster pool the work is routed to.

Queues are organization-scoped — a queue name is unique within the organization and is not qualified by project or domain.

~> **Destroying a queue soft-deletes it.** The control plane only accepts deletion once the queue is `drained`. Set `drain = true` and apply first to request draining. After the computed `state` reaches `drained`, run destroy to soft-delete it. A soft-deleted queue no longer accepts work, but its name remains reserved until the queue is undeleted.

~> A queue referenced as `run.default_queue` in settings at any scope cannot be deleted. Update or unset those settings before destroying such a queue.

### What `terraform destroy` reports

A queue must be drained before it can be deleted. Destroy behavior depends on the server's current state:

| Situation | Outcome |
| --- | --- |
| Queue is `active` and `drain = false` | Terraform fails before mutating the queue; set `drain = true` and apply first |
| Queue is `active` and `drain = true` | Terraform requests `draining`, then fails with retry guidance so state is retained |
| Queue is `draining` | Terraform fails with retry guidance so state is retained |
| Queue is `drained` | Terraform calls `DeleteQueue`, removes state on success, and warns that the queue was soft-deleted |
| Queue is already deleted or missing | Terraform removes state |

Destroy fails, and the resource stays in state, when the queue has not drained yet or when the control plane refuses deletion because the queue is still referenced as `run.default_queue`.


## Example Usage

### Basic queue

```terraform
resource "unionai_queue" "batch" {
  name     = "batch"
  clusters = ["*"]
}
```

### Pre-drain before deletion

```terraform
resource "unionai_queue" "batch" {
  name     = "batch"
  clusters = ["*"]

  drain = true
}
```

### Constrained production queue

```terraform
resource "unionai_queue" "production" {
  name              = "production"
  cluster_pool_name = "gpu-pool"
  clusters          = ["gpu-us-east-1", "gpu-us-west-2"]

  run_concurrency    = 50
  action_concurrency = 500
  depth              = 10000

  priority = "max"
  fairness = "shuffle_interleave"
}
```

## Schema

### Required

- `name` (String) Queue name. Must match `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` and be at most 63 characters. Unique within the organization. Changing this forces replacement.
- `clusters` (Set of String) Clusters this queue routes to. Must be a subset of the assigned cluster pool. `["*"]` routes to every enabled and healthy cluster in the pool, equally weighted, and must be the only element when used. An empty set means actions are not routed to any cluster.

### Optional

- `cluster_pool_name` (String) Cluster pool this queue routes into. Defaults to `default`, which the control plane creates on demand; any other pool must already exist. Can only be changed while the queue is `drained`; cluster-managed queues cannot change pool independently of their cluster.
- `run_concurrency` (Number) Maximum number of runs that may be active simultaneously. `0` (the default) means unlimited.
- `action_concurrency` (Number) Maximum number of actions (tasks) that may be in flight simultaneously. `0` (the default) means unlimited.
- `depth` (Number) Maximum total items the queue can hold, in-progress plus waiting. `0` (the default) means unlimited.
- `priority` (String) Scheduling priority relative to other queues: `min`, `medium` (the default), or `max`. Scheduling is strict — `max` always runs before `medium`, which always runs before `min`.
- `fairness` (String) Cross-project scheduling policy within this queue: `round_robin` (the default) or `shuffle_interleave`. The unit of fairness is per-project.
- `drain` (Boolean) When true, request that the queue stop accepting new work and drain existing work. Use this before destroying a queue. Defaults to `false`.

### Read-Only

- `id` (String) Queue identifier. Same as `name`.
- `state` (String) Queue lifecycle state: `active`, `draining`, or `drained`. Managed by the control plane.
- `cluster_managed` (Boolean) Whether this queue is the implicit queue owned by a cluster with the same name.
- `available_clusters` (Set of String) Names of the clusters this queue can currently route to, after filtering `clusters` down to those that are active and healthy.
- `created_at` (String) Creation timestamp, RFC 3339.
- `updated_at` (String) Last update timestamp, RFC 3339.
- `deleted_at` (String) Soft-deletion timestamp, RFC 3339. Null for live queues.

## Import

Queues can be imported using the queue name:

```shell
terraform import unionai_queue.batch batch
```

This is also the way to adopt an existing live queue. Soft-deleted queues keep their names reserved until they are undeleted.
