---
page_title: "Provider: Union.ai"
description: |-
  The Union.ai provider is used to interact with Union.ai resources.
---

# Union.ai Provider

The Union.ai provider is used to interact with Union.ai resources. The provider needs to be configured with proper credentials before it can be used.

## Example Usage

```terraform
terraform {
  required_providers {
    unionai = {
      source = "unionai/unionai"
    }
  }
}

provider "unionai" {
  # API key for authentication - can also be set via UNIONAI_API_KEY environment variable
  api_key = "your-api-key"

  # Optional: Override the organization name when it is not encoded in the API key
  # or differs from the control plane URL
  org = "your-org-name"

  # Optional: Restrict operations to specific organizations
  allowed_orgs = [
    "your-org-name",
  ]
}
```

## Authentication

The provider supports authentication via API key. You can obtain an API key using the `flyte` CLI:

```bash
flyte create api-key --name "terraform-api-key"
```

The API key can be provided in two ways:

1. **Via configuration** - Set the `api_key` attribute in the provider block
2. **Via environment variable** - Set the `UNIONAI_API_KEY` environment variable

## Schema

### Required

There are no required arguments for the provider configuration. However, you must provide authentication via either the `api_key` attribute or the `UNIONAI_API_KEY` environment variable.

### Optional

- `api_key` (String, Sensitive) - Union.ai API key for authentication. Can also be set via the `UNIONAI_API_KEY` environment variable.
- `org` (String) - Union.ai organization name. If set, this takes precedence over the organization encoded in the API key or inferred from the API key host. Use this when the control plane's organization name differs from the URL it is served from.
- `allowed_orgs` (Set of String) - List of organization names that this provider is allowed to manage. If specified, the provider will only allow operations on resources belonging to these organizations. This is useful to avoid unintended side effects when using multiple credentials or working with multiple organizations. Can also be set via the `UNIONAI_ALLOWED_ORGS` environment variable (comma-separated list).

## Organization Restriction

The `allowed_orgs` setting provides an additional safety mechanism when working with Terraform. By specifying which organizations the provider can manage, you can:

- Prevent accidental modifications to the wrong organization
- Safely use different credentials in different workspaces
- Add an extra layer of protection in CI/CD pipelines

If the API key's organization is not in the `allowed_orgs` list, the provider will refuse to operate and return an error.

## Useful Links

- [Union.ai Documentation](https://docs.union.ai/)
- [Report Issues](mailto:support+terraform@union.ai)
