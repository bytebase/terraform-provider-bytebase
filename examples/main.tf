# This is an example for using Bytebase Terraform provider to manage your resource.
# To run this provider in your local machine,
# 1. Run your Bytebase service, then you can access the OpenAPI via http://localhost:8080/v1
# 2. Replace the service_account and service_key with your own Bytebase service account
# 3. Run `make install` under terraform-provider-bytebase folder
# 4. Run `cd examples && terraform init`
# 5. Run `terraform plan` to check the changes
# 6. Run `terraform apply` to apply the changes
# 7. Run `terraform output` to find the outputs
# 8. Run `terraform destory` to delete the test resources
provider "bytebase" {
  # You need to replace the account and key with your Bytebase service account.
  service_account = "ed+dev@service.bytebase.com"
  service_key     = "ed"
  url             = "http://localhost:8080"
}

locals {
  environment_name = "dev"
  instance_name    = "dev instance"
}

# Create a new environment named "dev"
resource "bytebase_environment" "dev" {
  name = local.environment_name
  # You can specific the environment order
  # order = 1
}

# Print the new environment
output "staging_environment" {
  value = bytebase_environment.dev
}

# Create a new instance named "dev instance"
resource "bytebase_instance" "dev_instance" {
  name        = local.instance_name
  engine      = "POSTGRES"
  host        = "127.0.0.1"
  environment = bytebase_environment.dev.name
  # You can also provide the port, username, password
  # port = 5432
  # username = "username"
  # password = "password"
}

# Print the new instance
output "dev_instance" {
  value = bytebase_instance.dev_instance
  # The password in instance is sensitive, so you cannot directly get its value from the output.
  # But we can still print the instance via `terraform output -json dev_instance`
  sensitive = true
}

# import environments module and filter by environment name
module "environment" {
  source           = "./environments"
  environment_name = local.environment_name
  # Make sure the module exec after the "dev" environment is created
  depends_on = [
    bytebase_environment.dev
  ]
}

output "environment" {
  value = module.environment.environment
}

# import instances module and filter by instance name
module "instance" {
  source        = "./instances"
  instance_name = local.instance_name
  # Make sure the module exec after the "dev instance" instance is created
  depends_on = [
    bytebase_instance.dev_instance
  ]
}

output "instance" {
  value = module.instance.instance
}
