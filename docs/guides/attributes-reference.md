---
page_title: "Attributes Reference Guide"
subcategory: "Reference"
---

# Attributes Reference Guide

This guide provides a comprehensive reference of all attributes available for Union.ai Terraform resources and data sources. Use this as a quick lookup when working with outputs and attribute references.

## Resources

### unionai_api_key

**Inputs:**
- `id` (String, Required) - API key identifier (forces replacement)

**Outputs:**
- `id` (String) - API key identifier
- `secret` (String, Sensitive) - API key secret (only available after creation)

**Example Output Usage:**
```terraform
output "api_key" {
  value     = unionai_api_key.example.secret
  sensitive = true
}
```

---

### unionai_application

**Inputs:**
- `client_id` (String, Required) - OAuth client ID (forces replacement)
- `client_name` (String, Required) - Application name
- `client_uri` (String, Optional) - Application website URI
- `consent_method` (String, Optional) - Consent method
- `grant_types` (Set of String, Optional) - OAuth grant types
- `logo_uri` (String, Optional) - Logo URI
- `policy_uri` (String, Optional) - Privacy policy URI
- `redirect_uris` (Set of String, Optional) - Redirect URIs
- `response_types` (Set of String, Optional) - OAuth response types
- `token_endpoint_auth_method` (String, Optional) - Token endpoint auth method
- `tos_uri` (String, Optional) - Terms of service URI

**Outputs:**
- `id` (String) - Application identifier
- `client_id` (String) - OAuth client ID
- `client_name` (String) - Application name
- `client_uri` (String) - Application website URI
- `consent_method` (String) - Consent method
- `grant_types` (Set of String) - OAuth grant types
- `logo_uri` (String) - Logo URI
- `policy_uri` (String) - Privacy policy URI
- `redirect_uris` (Set of String) - Redirect URIs
- `response_types` (Set of String) - OAuth response types
- `token_endpoint_auth_method` (String) - Token endpoint auth method
- `tos_uri` (String) - Terms of service URI
- `secret` (String, Sensitive) - OAuth client secret (only available after creation)

**Example Output Usage:**
```terraform
output "app_details" {
  value = {
    id          = unionai_application.example.id
    client_id   = unionai_application.example.client_id
    client_name = unionai_application.example.client_name
  }
}

output "app_secret" {
  value     = unionai_application.example.secret
  sensitive = true
}
```

---

### unionai_application_access

**Inputs:**
- `policy` (String, Required) - Policy identifier
- `app` (String, Required) - Application identifier

**Outputs:**
- `policy` (String) - Policy identifier
- `app` (String) - Application identifier

**Example Output Usage:**
```terraform
output "access_info" {
  value = {
    policy = unionai_application_access.example.policy
    app    = unionai_application_access.example.app
  }
}
```

---

### unionai_policy

**Inputs:**
- `name` (String, Required) - Policy name
- `description` (String, Optional) - Policy description
- `organization` (Block Set, Optional) - Organization-level assignments
  - `id` (String, Required) - Organization ID (forces replacement)
  - `role_id` (String, Required) - Role ID (forces replacement)
- `project` (Block Set, Optional) - Project-level assignments
  - `id` (String, Required) - Project ID (forces replacement)
  - `domains` (Set of String, Required) - Domain IDs (forces replacement)
  - `role_id` (String, Required) - Role ID (forces replacement)
- `domain` (Block Set, Optional) - Domain-level assignments
  - `id` (String, Required) - Domain ID (forces replacement)
  - `role_id` (String, Required) - Role ID (forces replacement)

**Outputs:**
- `id` (String) - Policy identifier
- `name` (String) - Policy name
- `description` (String) - Policy description
- `organization` (Set of Object) - Organization-level assignments
- `project` (Set of Object) - Project-level assignments
- `domain` (Set of Object) - Domain-level assignments

**Example Output Usage:**
```terraform
output "policy_id" {
  value = unionai_policy.example.id
}

output "policy_details" {
  value = {
    name        = unionai_policy.example.name
    description = unionai_policy.example.description
  }
}
```

---

### unionai_project

**Inputs:**
- `name` (String, Required) - Project name
- `description` (String, Optional) - Project description

**Outputs:**
- `id` (String) - Project identifier
- `name` (String) - Project name
- `description` (String) - Project description

**Example Output Usage:**
```terraform
output "project_id" {
  value = unionai_project.example.id
}

output "project_info" {
  value = {
    id          = unionai_project.example.id
    name        = unionai_project.example.name
    description = unionai_project.example.description
  }
}
```

---

### unionai_queue

**Inputs:**
- `name` (String, Required) - Queue name (forces replacement)
- `clusters` (Set of String, Required) - Clusters to route to; `["*"]` means all healthy clusters in the pool
- `cluster_pool_name` (String, Optional) - Cluster pool, defaults to `default` (can only be changed while the queue is drained)
- `run_concurrency` (Number, Optional) - Max concurrent runs, `0` = unlimited
- `action_concurrency` (Number, Optional) - Max in-flight actions, `0` = unlimited
- `depth` (Number, Optional) - Max queued items, `0` = unlimited
- `priority` (String, Optional) - `min`, `medium` (default), or `max`
- `fairness` (String, Optional) - `round_robin` (default) or `shuffle_interleave`
- `drain` (Boolean, Optional) - Request draining before deleting the queue; defaults to `false`

**Outputs:**
- `id` (String) - Queue identifier, same as `name`
- `state` (String) - `active`, `draining`, or `drained`
- `cluster_managed` (Boolean) - Whether this queue is owned by a cluster with the same name
- `available_clusters` (Set of String) - Clusters currently routable, filtered to active and healthy
- `created_at` (String) - Creation timestamp, RFC 3339
- `updated_at` (String) - Last update timestamp, RFC 3339
- `deleted_at` (String) - Soft-deletion timestamp, RFC 3339; null for live queues

**Example Output Usage:**
```terraform
output "queue_state" {
  value = unionai_queue.example.state
}

output "queue_routing" {
  value = {
    pool      = unionai_queue.example.cluster_pool_name
    available = unionai_queue.example.available_clusters
  }
}
```

**Note:** Set `drain = true` and apply before deleting a queue. Once computed `state` reaches `drained`, `terraform destroy` soft-deletes the queue.

---

### unionai_role

**Inputs:**
- `name` (String, Required) - Role name (forces replacement)
- `description` (String, Optional) - Role description (forces replacement)
- `actions` (Set of String, Required) - Policy actions (forces replacement)

**Outputs:**
- `id` (String) - Role identifier
- `name` (String) - Role name
- `description` (String) - Role description
- `actions` (Set of String) - Policy actions

**Example Output Usage:**
```terraform
output "role_id" {
  value = unionai_role.example.id
}

output "role_actions" {
  value = unionai_role.example.actions
}
```

---

### unionai_user

**Inputs:**
- `first_name` (String, Required) - User's first name (forces replacement)
- `last_name` (String, Required) - User's last name (forces replacement)
- `email` (String, Required) - User's email (forces replacement)

**Outputs:**
- `id` (String) - User identifier
- `first_name` (String) - User's first name
- `last_name` (String) - User's last name
- `email` (String) - User's email

**Example Output Usage:**
```terraform
output "user_id" {
  value = unionai_user.example.id
}

output "user_email" {
  value = unionai_user.example.email
}
```

---

### unionai_user_access

**Inputs:**
- `policy` (String, Required) - Policy identifier (forces replacement)
- `user` (String, Required) - User identifier (forces replacement)

**Outputs:**
- `policy` (String) - Policy identifier
- `user` (String) - User identifier

**Example Output Usage:**
```terraform
output "access_assignment" {
  value = {
    policy = unionai_user_access.example.policy
    user   = unionai_user_access.example.user
  }
}
```

---

## Data Sources

### data.unionai_api_key

**Inputs:**
- `id` (String, Required) - API key identifier

**Outputs:**
- `id` (String) - API key identifier
- `secret` (String, Sensitive) - Empty for existing keys (only available during creation)

**Example Usage:**
```terraform
data "unionai_api_key" "example" {
  id = "my-api-key"
}

output "api_key_id" {
  value = data.unionai_api_key.example.id
}
```

---

### data.unionai_application

**Inputs:**
- `id` (String, Required) - Application identifier (client ID)

**Outputs:**
- `client_id` (String) - OAuth client ID
- `client_name` (String) - Application name
- `client_uri` (String) - Application website URI
- `grant_types` (Set of String) - OAuth grant types
- `logo_uri` (String) - Logo URI
- `policy_uri` (String) - Privacy policy URI
- `redirect_uris` (Set of String) - Redirect URIs
- `response_types` (Set of String) - OAuth response types
- `token_endpoint_auth_method` (String) - Token endpoint auth method
- `tos_uri` (String) - Terms of service URI
- `secret` (String, Sensitive) - Empty for existing applications

**Example Usage:**
```terraform
data "unionai_application" "example" {
  id = "my-app-client-id"
}

output "app_name" {
  value = data.unionai_application.example.client_name
}
```

---

### data.unionai_application_access

**Inputs:**
- `app_id` (String, Required) - Application identifier
- `policy_id` (String, Required) - Policy identifier

**Outputs:**
- `app_id` (String) - Application identifier
- `policy_id` (String) - Policy identifier

**Example Usage:**
```terraform
data "unionai_application_access" "example" {
  app_id    = "my-app"
  policy_id = "my-policy"
}
```

---

### data.unionai_controlplane

**Inputs:**
None (reads current controlplane configuration)

**Outputs:**
- `endpoint` (String) - Controlplane endpoint
- `host` (String) - Controlplane host
- `organization` (String) - Organization identifier

**Example Usage:**
```terraform
data "unionai_controlplane" "current" {}

output "org_id" {
  value = data.unionai_controlplane.current.organization
}
```

---

### data.unionai_dataplane

**Inputs:**
- `id` (String, Required) - Dataplane/cluster identifier

**Outputs:**
- `id` (String) - Cluster identifier
- `state` (String) - Cluster state
- `health` (String) - Dataplane health status

**Example Usage:**
```terraform
data "unionai_dataplane" "example" {
  id = "my-dataplane"
}

output "dataplane_health" {
  value = data.unionai_dataplane.example.health
}
```

---

### data.unionai_dataplanes

**Inputs:**
None (lists all dataplanes)

**Outputs:**
- `ids` (Set of String) - List of all dataplane IDs

**Example Usage:**
```terraform
data "unionai_dataplanes" "all" {}

output "all_dataplane_ids" {
  value = data.unionai_dataplanes.all.ids
}
```

---

### data.unionai_policy

**Inputs:**
- `id` (String, Required) - Policy identifier

**Outputs:**
- `description` (String) - Policy description
- `roles` (List of Object) - Policy role assignments
  - `role_id` (String) - Role ID
  - `resource` (Object) - Resource assignment
    - `org_id` (String) - Organization ID (if org-level)
    - `domain_id` (String) - Domain ID (if domain-level)
    - `project_id` (String) - Project ID (if project-level)

**Example Usage:**
```terraform
data "unionai_policy" "example" {
  id = "my-policy"
}

output "policy_description" {
  value = data.unionai_policy.example.description
}

output "policy_roles" {
  value = data.unionai_policy.example.roles
}
```

---

### data.unionai_project

**Inputs:**
- `id` (String, Required) - Project identifier

**Outputs:**
- `name` (String) - Project name
- `description` (String) - Project description
- `domain_ids` (Set of String) - Associated domain IDs
- `state` (String) - Project state

**Example Usage:**
```terraform
data "unionai_project" "example" {
  id = "my-project"
}

output "project_domains" {
  value = data.unionai_project.example.domain_ids
}
```

---

### data.unionai_queue

**Inputs:**
- `id` (String, Required) - Queue name

**Outputs:**
- `cluster_pool_name` (String) - Cluster pool this queue routes into
- `clusters` (Set of String) - Clusters this queue routes to
- `run_concurrency` (Number) - Max concurrent runs, `0` = unlimited
- `action_concurrency` (Number) - Max in-flight actions, `0` = unlimited
- `depth` (Number) - Max queued items, `0` = unlimited
- `priority` (String) - `min`, `medium`, or `max`
- `fairness` (String) - `round_robin` or `shuffle_interleave`
- `state` (String) - `active`, `draining`, or `drained`
- `available_clusters` (Set of String) - Clusters currently routable, filtered to active and healthy
- `created_at` (String) - Creation timestamp, RFC 3339
- `updated_at` (String) - Last update timestamp, RFC 3339

**Example Output Usage:**
```terraform
output "queue_state" {
  value = data.unionai_queue.example.state
}
```

---

### data.unionai_role

**Inputs:**
- `id` (String, Required) - Role identifier

**Outputs:**
- `actions` (Set of String) - Role actions/permissions

**Example Usage:**
```terraform
data "unionai_role" "example" {
  id = "my-role"
}

output "role_permissions" {
  value = data.unionai_role.example.actions
}
```

---

### data.unionai_user

**Inputs:**
- `id` (String, Optional) - User identifier
- `email` (String, Optional) - User email

**Note:** Either `id` or `email` must be specified.

**Outputs:**
- `id` (String) - User identifier
- `first_name` (String) - User's first name
- `last_name` (String) - User's last name
- `email` (String) - User's email

**Example Usage:**
```terraform
# Look up by ID
data "unionai_user" "by_id" {
  id = "user-123"
}

# Look up by email
data "unionai_user" "by_email" {
  email = "user@example.com"
}

output "user_full_name" {
  value = "${data.unionai_user.by_email.first_name} ${data.unionai_user.by_email.last_name}"
}
```

---

### data.unionai_user_access

**Inputs:**
- `user_id` (String, Required) - User identifier
- `policy_id` (String, Required) - Policy identifier

**Outputs:**
- `user_id` (String) - User identifier
- `policy_id` (String) - Policy identifier

**Example Usage:**
```terraform
data "unionai_user_access" "example" {
  user_id   = "user-123"
  policy_id = "policy-456"
}
```

---

## Common Patterns

### Chaining Resources with Outputs

```terraform
# Create a project
resource "unionai_project" "ml_project" {
  name        = "ml-workflows"
  description = "Machine learning project"
}

# Create a role
resource "unionai_role" "developer" {
  name        = "developer"
  description = "Developer role"
  actions     = ["read", "execute"]
}

# Create a policy using outputs from above
resource "unionai_policy" "dev_policy" {
  name = "dev-ml-policy"

  project {
    id      = unionai_project.ml_project.id
    role_id = unionai_role.developer.id
    domains = ["development"]
  }
}

# Assign to user
resource "unionai_user_access" "dev_access" {
  policy = unionai_policy.dev_policy.id
  user   = unionai_user.developer.id
}
```

### Using Data Sources for Reference

```terraform
# Reference existing resources
data "unionai_project" "production" {
  id = "prod-project-id"
}

data "unionai_user" "admin" {
  email = "admin@example.com"
}

# Use their attributes
output "prod_domains" {
  value = data.unionai_project.production.domain_ids
}

output "admin_id" {
  value = data.unionai_user.admin.id
}
```

### Handling Sensitive Outputs

```terraform
# Mark sensitive outputs appropriately
output "api_credentials" {
  value = {
    api_key_id = unionai_api_key.ci.id
    api_secret = unionai_api_key.ci.secret
  }
  sensitive = true
}

output "app_oauth" {
  value = {
    client_id     = unionai_application.app.client_id
    client_secret = unionai_application.app.secret
  }
  sensitive = true
}
```
