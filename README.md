# Terraform Provider Bytebase

This repository is the Terraform provider for Bytebase.

## Development

### Prerequisites

- [Go](https://golang.org/doc/install) (1.19 or later)
- [Air](https://github.com/cosmtrek/air#installation) (must use forked repo 87187cc). This is for backend live reload.
- [Terraform](https://developer.hashicorp.com/terraform/downloads?product_intent=terraform) (1.3.5 or later)

### Prepare

```bash
# clone Bytebase to get the OpenAPI server
git clone git@github.com:bytebase/bytebase.git

git clone git@github.com:bytebase/terraform-provider-bytebase.git
```

```bash
# start Bytebase OpenAPI server
cd bytebase
# check https://github.com/bytebase/bytebase for more information.
air -c scripts/.air.toml
```

### Build and test

```bash
# install the provider in your local machine
cd terraform-provider-bytebase && make install

# initial the terraform for your example
# you may need to change the username and password
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
