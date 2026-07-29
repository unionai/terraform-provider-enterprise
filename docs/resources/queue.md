---
page_title: "unionai_queue Resource - terraform-provider-unionai"
subcategory: ""
description: |-
  Manages a Union.ai queue.
---

# unionai_queue (Resource)

Manages a Union.ai queue. Queues control admission and scheduling of runs: how many runs and actions may be in flight, how deep the backlog may get, scheduling priority relative to other queues, how work from different projects is interleaved, and which clusters within a cluster pool the work is routed to.

Queues are organization-scoped — a queue name is unique within the organization and is not qualified by project or domain.

~> **Destroying a queue drains it.** The Union API has no delete operation for queues. `terraform destroy` moves the queue to the `draining` state (which the control plane advances to `drained` once in-flight work finishes) and removes it from Terraform state. **The queue and its name persist.** A later `terraform apply` of the same configuration will fail with an "already exists" error; use `terraform import` to adopt the existing queue instead.

~> **The `default` queue cannot be drained**, and neither can a queue referenced as `run.default_queue` in settings at any scope. Update or unset those settings before destroying such a queue.

## Example Usage

### Basic queue

```terraform
resource "unionai_queue" "batch" {
  name     = "batch"
  clusters = ["*"]
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

- `name` (String) Queue name. Must match `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` and be at most 63 characters. Unique within the organization. Changing this forces replacement — note that the old queue is only drained, so its name stays reserved.
- `clusters` (Set of String) Clusters this queue routes to. Must be a subset of the assigned cluster pool. `["*"]` routes to every enabled and healthy cluster in the pool, equally weighted, and must be the only element when used. An empty set means actions are not routed to any cluster.

### Optional

- `cluster_pool_name` (String) Cluster pool this queue routes into. Defaults to `default`, which the control plane creates on demand; any other pool must already exist. **Cannot be changed after creation** — the provider rejects a change at plan time.
- `run_concurrency` (Number) Maximum number of runs that may be active simultaneously. `0` (the default) means unlimited.
- `action_concurrency` (Number) Maximum number of actions (tasks) that may be in flight simultaneously. `0` (the default) means unlimited.
- `depth` (Number) Maximum total items the queue can hold, in-progress plus waiting. `0` (the default) means unlimited.
- `priority` (String) Scheduling priority relative to other queues: `min`, `medium` (the default), or `max`. Scheduling is strict — `max` always runs before `medium`, which always runs before `min`.
- `fairness` (String) Cross-project scheduling policy within this queue: `round_robin` (the default) or `shuffle_interleave`. The unit of fairness is per-project.

### Read-Only

- `id` (String) Queue identifier. Same as `name`.
- `state` (String) Queue lifecycle state: `active`, `draining`, or `drained`. Managed by the control plane.
- `available_clusters` (Set of String) Names of the clusters this queue can currently route to, after filtering `clusters` down to those that are active and healthy.
- `created_at` (String) Creation timestamp, RFC 3339.
- `updated_at` (String) Last update timestamp, RFC 3339.

## Import

Queues can be imported using the queue name:

```shell
terraform import unionai_queue.batch batch
```

This is also the way to recover after a `terraform destroy`, since destroying a queue only drains it and leaves the name reserved.
