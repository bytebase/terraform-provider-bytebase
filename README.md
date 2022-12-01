# Terraform Provider Bytebase

This repository is the Terraform provider for Bytebase.

## Usage

You can download this provider at [registry.terraform.io](https://registry.terraform.io/providers/bytebase/bytebase).

Please follow this docs [Manage Bytebase with Terraform](https://www.bytebase.com/docs/get-started/terraform) to use the provider.

## Development

### Prerequisites

- [Go](https://golang.org/doc/install) (1.19 or later)
- [Terraform](https://developer.hashicorp.com/terraform/downloads?product_intent=terraform) (1.3.5 or later)
- [Bytebase](https://github.com/bytebase/bytebase)

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

# initialize the terraform for your example
# you need to set the service_account and service_key to your own
cd examples && terraform init

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

This will generate the doc template in the `docs` folder

```bash
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs --provider-name=terraform-provider-bytebase
```

## Release

Follow [this doc](https://developer.hashicorp.com/terraform/registry/providers/publishing) to publish the provider.

> Note:
> We need to publish a new tag for a new version, the tag must be a valid [Semantic Version](https://semver.org/) **preceded with a v (for example, v1.2.3)**. There must not be a branch name with the same name as the tag.

1. Develop and merge the feature code.
2. Create a new PR to update the version in [`./VERSION`](./VERSION)
3. After the version is updated, the action [`./.github/workflows/release.yml`](./.github/workflows/release.yml) will use the newest version `x.y.z` to create a new tag `vx.y.z`, then use the tag to create the release.
