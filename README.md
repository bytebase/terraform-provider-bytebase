# Terraform Provider Bytebase

This repository is the Terraform provider for [Bytebase](https://bytebase.com). A typical setup
involves teams using Cloud vendors' Terraform provider to provision database instances, followed by
using Terraform Bytebase Provider to prepare those instances ready for application use.

![Overview](https://raw.githubusercontent.com/bytebase/terraform-provider-bytebase/main/docs/assets/overview.webp)

## Usage

1. Download [provider](https://registry.terraform.io/providers/bytebase/bytebase).
1. Follow [example](https://www.bytebase.com/docs/get-started/terraform).

## Development

### Prerequisites

- [Go](https://golang.org/doc/install) (1.19 or later)
- [Terraform](https://developer.hashicorp.com/terraform/downloads?product_intent=terraform) (1.3.5 or later)
- [Bytebase](https://github.com/bytebase/bytebase) (3.10.0 or later)

> If you have problems running `terraform` in MacOS with Apple Silicon, you can following https://stackoverflow.com/questions/66281882/how-can-i-get-terraform-init-to-run-on-my-apple-silicon-macbook-pro-for-the-go and use the `tfenv`.

### Prepare Bytebase OpenAPI server

```bash
# clone Bytebase to get the OpenAPI server
git clone git@github.com:bytebase/bytebase.git

git clone git@github.com:bytebase/terraform-provider-bytebase.git
```

```bash
# start Bytebase OpenAPI server
cd bytebase
# check https://github.com/bytebase/bytebase for starting the Bytebase server.
air -c scripts/.air.toml
```

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
terraform destory
```

### Generate docs

> This will generate the doc template in the `docs` folder
>
> Check https://github.com/hashicorp/terraform-plugin-docs and https://github.com/hashicorp/terraform-plugin-docs/issues/141 for details.

```bash
GOOS=darwin GOARCH=amd64 go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs --provider-name=terraform-provider-bytebase
```

## Release

Follow [this doc](https://developer.hashicorp.com/terraform/registry/providers/publishing) to publish the provider.

> Note:
> We need to publish a new tag for a new version, the tag must be a valid [Semantic Version](https://semver.org/) **preceded with a v (for example, v1.2.3)**. There must not be a branch name with the same name as the tag.

1. Develop and merge the feature code.
1. Create a new PR to update the version in [`./VERSION`](./VERSION)
1. After the version is updated, the action [`./.github/workflows/release.yml`](./.github/workflows/release.yml) will use the newest version `x.y.z` to create a new tag `vx.y.z`, then use the tag to create the release.
