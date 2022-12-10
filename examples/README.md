# Examples for Bytebase Terraform Porvider

This is an example for using Bytebase Terraform provider to manage your resource.

To run this provider in your local machine:

1. Run your Bytebase service, then you can access the service via `http://localhost:8080`
2. Create the service account. Docs: https://www.bytebase.com/docs/get-started/work-with-terraform/overview
3. Replace the `service_account` and `service_key` with your own Bytebase service account, replace the `url` with your Bytebase service URL
4. Run `cd examples && terraform init`
5. Run `terraform plan` to check the changes
6. Run `terraform apply` to apply the changes
7. Run `terraform output` to find the outputs
8. Run `terraform destory` to delete the test resources

To run this provider for development and test:

1. Replace the source with `terraform.local/bytebase/bytebase` in `./versions.tf`, `./environments/main.tf` and `./instances/main.tf`
2. Run `make install` under the `terraform-provider-bytebase` folder
3. Run `cd examples && terraform init`
