# This is an example for using Bytebase Terraform provider to manage your resource.
# Docs: https://www.bytebase.com/docs/get-started/work-with-terraform/overview
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
  service_account = "terraform@service.bytebase.com"
  service_key     = "bbs_qHX6CswQ1nMMELSCc2lk"
  # The Bytebase service URL. You can use the external URL in production.
  # Check the docs about external URL: https://www.bytebase.com/docs/get-started/install/external-url
  url = "https://bytebase.example.com"
}

locals {
  environment_name_dev  = "dev"
  environment_name_prod = "prod"
  instance_name         = "dev instance"
}

# Create a new environment named "dev"
resource "bytebase_environment" "dev" {
  name  = local.environment_name_dev
  order = 0
}

# Create another environment named "prod"
resource "bytebase_environment" "prod" {
  name  = local.environment_name_prod
  order = 1
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
  environment_name = local.environment_name_dev
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

# find single environment named "dev"
data "bytebase_environment" "dev" {
  name = local.environment_name_dev
  depends_on = [
    bytebase_environment.dev
  ]
}

output "dev_env" {
  value = data.bytebase_environment.dev
}
