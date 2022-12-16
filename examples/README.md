# Examples for Bytebase Terraform Porvider

Examples of using the Bytebase Terraform provider to manage your resource:

- [environments](./environments): Use the Bytebase provider to CRUD the environment.
- [instances](./instances): Use the Bytebase provider to CRUD the instance.

To run this provider on your local machine:

1. Run your Bytebase service, then you can access the service via `http://localhost:8080`.
2. Create the service account. Docs [Create service account](https://www.bytebase.com/docs/get-started/terraform#create-service-account).
3. Replace the `service_account` and `service_key` with your Bytebase service account, and replace the `url` with your Bytebase service URL.
4. Go to the `environments` or `instances` folder.
5. Run `terraform init` to install the provider.
6. Run `terraform plan` to check the changes.
7. Run `terraform apply` to apply the changes.
8. Run `terraform destory` to delete the test resources.

To run this provider for development and testing:

1. Replace the source `registry.terraform.io/bytebase/bytebase` with `terraform.local/bytebase/bytebase`.
2. Run `make install` under the `terraform-provider-bytebase` folder.
3. Go to the example folder and run `terraform init`.
