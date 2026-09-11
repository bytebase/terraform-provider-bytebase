# Terraform Provider Bytebase

This repository is the Terraform provider for [Bytebase](https://bytebase.com). A typical setup
involves teams using Cloud vendors' Terraform provider to provision database instances, followed by
using Terraform Bytebase Provider to prepare those instances ready for application use.

## Usage

1. Download [provider](https://registry.terraform.io/providers/bytebase/bytebase).
1. Follow [example](https://docs.bytebase.com/integrations/terraform/overview).

Provider configuration can include custom HTTP headers for access gateways:

```hcl
provider "bytebase" {
  url             = "https://bytebase.example.com"
  service_account = "service@example.com"
  service_key     = var.bytebase_service_key

  custom_header {
    name  = "zero_trust_token"
    value = var.zero_trust_token
  }
}
```

The provider can also exchange an external OIDC token for a short-lived
Bytebase access token. An administrator must create the workload identity and
grant its IAM roles before switching Terraform to this authentication mode.

```hcl
provider "bytebase" {
  url                          = "https://bytebase.example.com"
  workload_identity_email      = "terraform@workload.bytebase.com"
  workload_identity_token_file = "/secrets/bytebase.jwt"
}
```

Use exactly one authentication mode: either `service_account` with
`service_key`, or `workload_identity_email` with one of
`workload_identity_token` and `workload_identity_token_file`. The equivalent
workload identity environment variables are:

- `BYTEBASE_WORKLOAD_IDENTITY_EMAIL`
- `BYTEBASE_WORKLOAD_IDENTITY_TOKEN`
- `BYTEBASE_WORKLOAD_IDENTITY_TOKEN_FILE`

The file-backed mode is recommended for workloads such as Nomad because the
provider rereads the file when Bytebase authentication expires. External OIDC
tokens and returned Bytebase tokens are kept out of Terraform state and should
not be written to logs. Remove or unset `BYTEBASE_SERVICE_ACCOUNT` and
`BYTEBASE_SERVICE_KEY` when switching to workload identity authentication.

## Development

### Prerequisites

- [Go](https://go.dev/doc/install) (1.25.0 or later)
- [Terraform](https://developer.hashicorp.com/terraform/downloads?product_intent=terraform) (1.11 or later, required for write-only attributes)
- [Bytebase](https://github.com/bytebase/bytebase) (3.23.0 or later)

> If Terraform has problems on macOS with Apple Silicon, follow this [troubleshooting guide](https://stackoverflow.com/questions/66281882/how-can-i-get-terraform-init-to-run-on-my-apple-silicon-macbook-pro-for-the-go) and use `tfenv`.

### Prepare a Bytebase server

```bash
git clone git@github.com:bytebase/bytebase.git
git clone git@github.com:bytebase/terraform-provider-bytebase.git
```

Start a compatible Bytebase server by following the [Bytebase development instructions](https://github.com/bytebase/bytebase#development).

### Build and test

```bash
# install the provider in your local machine
cd terraform-provider-bytebase && make install

# test
# Any BYTEBASE_SERVICE_ACCOUNT/BYTEBASE_SERVICE_KEY/BYTEBASE_URL value should work since the service is mocked
TF_ACC=1 BYTEBASE_SERVICE_ACCOUNT=test@service.bytebase.com BYTEBASE_SERVICE_KEY=test_secret BYTEBASE_URL=https://bytebase.example.com go test -v ./...

# initialize the terraform for your example
# you need to set the service_account and service_key to your own
cd examples/setup && terraform init

# check the changes
terraform plan

# apply the changes
terraform apply

# print outputs
terraform output

# delete test resources
terraform destroy
```

### Generate docs

> This generates the documentation in the `docs` folder.
>
> Check https://github.com/hashicorp/terraform-plugin-docs and https://github.com/hashicorp/terraform-plugin-docs/issues/141 for details.

```bash
tfplugindocs generate --provider-name=terraform-provider-bytebase
```

## Release

Follow [this doc](https://developer.hashicorp.com/terraform/registry/providers/publishing) to publish the provider.

> Note:
> We need to publish a new tag for a new version, the tag must be a valid [Semantic Version](https://semver.org/) **preceded with a v (for example, v1.2.3)**. There must not be a branch name with the same name as the tag.

1. Develop and merge the feature code.
1. Create a new PR to update the version in [`./VERSION`](./VERSION)
1. After the version is updated, the action [`./.github/workflows/release.yml`](./.github/workflows/release.yml) will use the newest version `x.y.z` to create a new tag `vx.y.z`, then use the tag to create the release.
