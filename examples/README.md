# Examples for Bytebase Terraform Porvider

Examples of using the Terraform Bytebase Provider to manage your resource:

- [setup](./setup): Initialize the data.
- [environments](./environments): Use the Bytebase provider to query the environment.
- [instances](./instances): Use the Bytebase provider to query the instance.
- [roles](./roles): Use the Bytebase provider to manage instance role in your instance.
- [projects](./projects): Use the Bytebase provider to query projects in your Bytebase instance.

To run this provider on your local machine:

1. Run your Bytebase service, then you can access the service via `http://localhost:8080`.
1. Create the service account. Docs [Create service account](https://www.bytebase.com/docs/get-started/terraform#create-service-account).
1. Replace the `service_account` and `service_key` with your Bytebase service account, and replace the `url` with your Bytebase service URL.
1. Go to the [setup](./setup) to initialize the data.
1. Go to the `environments`/`instances`/`role`/`projects` folder to query the data.
1. Go to the [setup](./setup) and run `terraform destory` to delete the test resources.

To run this provider for development and testing:

1. Replace the source `registry.terraform.io/bytebase/bytebase` with `terraform.local/bytebase/bytebase`.
1. Run `make install` under the `terraform-provider-bytebase` folder.
1. Go to example folders and run `terraform init`.
