---
page_title: "unionai_project_domain_attributes Resource - terraform-provider-unionai"
subcategory: ""
description: |-
  Manages cluster resource attributes for a project-domain pair.
---

# unionai_project_domain_attributes (Resource)

Manages the cluster resource attributes (matchable attributes of type `CLUSTER_RESOURCE`) for a project-domain pair.

The attribute map is substituted into the cluster resource templates that Flyte renders for the project-domain namespace. One use case is setting `defaultUserRoleValue` to bind a per-project IAM role to the namespace's default ServiceAccount, giving each project-domain its own scoped cloud identity.

Another use case is setting project-domain resource quotas. Quota values are passed as cluster resource template variables, such as `projectQuotaCpu`, `projectQuotaMemory`, and `projectQuotaNvidiaGpu`.

## Example Usage

### Per-project IAM role

```terraform
resource "unionai_project" "test" {
  name        = "test"
  description = "Test Project"
}

# Bind a per-project IAM role to the project-domain namespace's default
# ServiceAccount by setting the defaultUserRoleValue cluster resource template variable.
resource "unionai_project_domain_attributes" "test" {
  project = unionai_project.test.id
  domain  = "development"

  attributes = {
    defaultUserRoleValue = "arn:aws:iam::123456789012:role/my-project-development-role"
  }
}
```

### Project-domain resource quotas

Set resource quotas for a specific project and domain by passing the quota template variables in `attributes`:

```terraform
resource "unionai_project_domain_attributes" "development_quotas" {
  project = "my-project"
  domain  = "development"

  attributes = {
    projectQuotaCpu       = "2"
    projectQuotaMemory    = "2Gi"
    projectQuotaNvidiaGpu = "1"
  }
}
```

The quota values are strings and should use the quantity formats expected by the cluster resource templates. For example, memory values can use Kubernetes-style quantities such as `2Gi`.

## Schema

### Required

- `project` (String) Project identifier the attributes apply to.
- `domain` (String) Domain the attributes apply to (e.g. `development`, `staging`, `production`).
- `attributes` (Map of String) Cluster resource template variables to substitute, as case-sensitive key/value pairs. Common examples include `defaultUserRoleValue` for a project-domain IAM role and quota keys such as `projectQuotaCpu`, `projectQuotaMemory`, and `projectQuotaNvidiaGpu`.

### Read-Only

- `id` (String) Resource identifier, in the form `{project}/{domain}`.

## Import

Project domain attributes can be imported using `{project}/{domain}`:

```shell
terraform import unionai_project_domain_attributes.test my-project/development
```
