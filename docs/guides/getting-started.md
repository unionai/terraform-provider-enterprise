---
page_title: "Getting Started with the Union.ai Provider"
subcategory: "Getting Started"
---

# Getting Started with the Union.ai Provider

This guide walks you through setting up the Union.ai Terraform provider and creating your first resources.

## Prerequisites

Before you begin, ensure you have:

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0 installed
- A Union.ai account
- The [Union CLI](https://docs.union.ai/cli/) installed

## Step 1: Create an API Key

First, create an API key using the `flyte` CLI:

```bash
flyte create api-key --name "terraform-api-key"
```

Save the generated API key securely. You'll need it to authenticate the provider.

## Step 2: Configure the Provider

Create a new directory for your Terraform configuration:

```bash
mkdir union-terraform
cd union-terraform
```

Create a `main.tf` file with the provider configuration:

```terraform
terraform {
  required_providers {
    unionai = {
      source  = "unionai/unionai"
      version = "~> 1.0"
    }
  }
}

provider "unionai" {
  api_key = var.unionai_api_key

  # Optional: Override organization when it is not encoded in the API key
  # or differs from the control plane URL
  org = "your-org-name"

  # Optional: Restrict to specific organizations
  allowed_orgs = ["your-org-name"]
}
```

Create a `variables.tf` file:

```terraform
variable "unionai_api_key" {
  description = "Union.ai API key for authentication"
  type        = string
  sensitive   = true
}
```

Create a `terraform.tfvars` file (add this to `.gitignore`):

```terraform
unionai_api_key = "your-api-key-here"
```

## Step 3: Initialize Terraform

Initialize your Terraform workspace:

```bash
terraform init
```

## Step 4: Create Your First Resource

Add a project resource to your `main.tf`:

```terraform
resource "unionai_project" "my_first_project" {
  name        = "my-terraform-project"
  description = "Project managed by Terraform"
}

output "project_id" {
  value = unionai_project.my_first_project.id
}
```

## Step 5: Apply Your Configuration

Preview the changes:

```bash
terraform plan
```

Apply the configuration:

```bash
terraform apply
```

Type `yes` when prompted to confirm.

## Next Steps

Now that you've created your first resource, you can:

- Review the available resources in the provider
- Explore the [Managing Projects and Users](./managing-projects-users.md) guide
- Learn about [Access Control with Policies](./access-control-policies.md)

### Synchronizing your Identity Provider

- Sync [Union policies with Google Workspace](./sync-policy-with-goog-ad.md)
- Sync [Union policies with Microsoft Entra ID](./sync-policy-with-msft-entra.md)

## Cleaning Up

To destroy the resources you created:

```bash
terraform destroy
```
