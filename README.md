# Union Terraform Provider

The Union.ai Terraform provider allows you to manage Union.ai resources using Infrastructure as Code.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for development)

## Using the Provider

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
  # API key for authentication - can also be set via UNIONAI_API_KEY environment variable
  api_key = var.unionai_api_key

  # Optional: Override the organization name if it is not encoded in the API key
  # or differs from the control plane URL
  org = "your-org-name"

  # Optional: Restrict operations to specific organizations
  allowed_orgs = ["your-org-name"]
}
```

## Authentication

The provider requires authentication via API key. You can obtain an API key using the Flyte CLI:

```bash
flyte create api-key admin --name "terraform-api-key"
```

The API key can be provided in two ways:
1. Set the `api_key` attribute in the provider configuration
2. Set the `UNIONAI_API_KEY` environment variable

If the organization name is not encoded in the API key, or if the control plane
is served from a URL that differs from the organization name, set the optional
`org` provider attribute. This value takes precedence over the API key org and
the host-derived fallback.

## Documentation

Full documentation is available on the [Terraform Registry](https://registry.terraform.io/providers/unionai/unionai/latest/docs).

## Allowed Organizations

To avoid unintended side effects by mixing credentials, you can specify the
organizations in the provider configuration. The provider will only allow
operations on resources in the allowed organizations.

## Available Resources

- `unionai_project` - Manage Union.ai projects
- `unionai_user` - Manage users
- `unionai_role` - Manage roles and permissions
- `unionai_policy` - Manage access policies
- `unionai_api_key` - Manage API keys
- `unionai_application` - Manage OAuth applications
- `unionai_user_access` - Assign policies to users
- `unionai_application_access` - Assign policies to applications

## Available Data Sources

- `unionai_project` - Read project information
- `unionai_user` - Read user information
- `unionai_role` - Read role information
- `unionai_policy` - Read policy information
- `unionai_api_key` - Read API key information
- `unionai_application` - Read application information
- `unionai_dataplane` - Read dataplane information
- `unionai_dataplanes` - List all dataplanes
- `unionai_controlplane` - Read controlplane information

## Developer Setup

### Prerequisites

0.  Have a shared folder for repos

        mkdir src
        cd src

1.  Clone this repository

        git clone git@github.com:unionai/terraform-provider-enterprise.git
        cd terraform-provider-enterprise

2.  Clone the `cloud` repo as a sibling directory (required for local development)

        cd ../
        git clone git@github.com:unionai/cloud.git

### Building

    cd terraform-provider-enterprise
    go build

## Testing the Provider Locally

When testing changes on the terraform provider you want to point to a local
version of the provider instead of getting it from the offical terraform
registry

- Run `go build` in the root directory of this repo -> creates
  `terraform-provider-enterprise`

- Create a `.terraformrc` file in the examples dir or any directory:

      provider_installation {
        dev_overrides {
          "unionai/unionai" = "<path-to-the-binary-built>"
        }
        direct {}
      }

- Set the `TF_CLI_CONFIG_FILE` env var to this file

      export TF_CLI_CONFIG_FILE="<path-to-the-terraformrc>"/.terraformrc

- By unsetting this env var you can switch to getting it from the official
  registry again

## Publishing to Terraform Registry

See [PUBLISHING.md](./PUBLISHING.md) for detailed instructions on publishing this provider to the Terraform Registry.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `go test ./...`
5. Submit a pull request

## License

This project is licensed under the Mozilla Public License Version 2.0 - see the [LICENSE](LICENSE) file for details.
