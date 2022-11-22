# This is an example for using Bytebase Terraform provider to manage your resource.
# To run this provider in your local machine,
# 1. Run your Bytebase service, then you can access the OpenAPI via http://localhost:8080/v1
# 2. Replace the email and password with your own Bytebase account
# 3. Run `make install` under terraform-provider-bytebase folder
# 4. Run `cd examples && terraform init`
# 5. Run `terraform plan` to check the changes
# 6. Run `terraform apply` to apply the changes
# 7. Run `terraform output` to find the outputs
# 8. Run `terraform destory` to delete the test resources
terraform {
  required_providers {
    bytebase = {
      version = "0.0.1"
      # The source is only used in the local example.
      source = "registry.terraform.io/bytebase/bytebase"
    }
  }
}

provider "bytebase" {
  # You need to replace the email and password with your own Bytebase account.
  email        = "ed+dev@bytebase.com"
  password     = "ed"
  bytebase_url = "http://localhost:8080/v1"
}

# Create a new environment named "dev"
resource "bytebase_environment" "dev" {
  name = "dev"
  # You can specific the environment order
  # order = 1
}

# Print the new environment
output "staging_environment" {
  value = bytebase_environment.dev
}

# Create a new instance named "dev instance"
resource "bytebase_instance" "dev_instance" {
  name        = "dev instance"
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

# List data source
data "bytebase_environments" "all" {}
data "bytebase_instances" "all" {}

output "all_environments" {
  value = data.bytebase_environments.all.environments
}

output "all_instances" {
  value = data.bytebase_instances.all.instances
}
