# Setup examples

_For a real world example, see https://github.com/bytebase/terraform-example._

This is the setup for examples.
We will create two environments named `test` and `prod`. Each environment contains one instance.

Before you start, please make sure you have running your Bytebase service and have created the service account, and replace the provider initial variables. Check the [README](../README.md) for details.

1. Run `terraform init` to install the provider.
1. Run `terraform plan` to check the changes.
1. Run `terraform apply` to apply the changes.
