# This is an example for using bytebase terraform provider to manage your resource.
# To run this provider in your local machine,
# 1. Run your bytebase service, then you can access the OpenAPI via http://localhost:8080/v1
# 2. Replace the email and password with your account
# 3. Run `make install` under terraform-provider-bytebase folder
# 4. Run `terraform init` under terraform-provider-bytebase/examples folder
# 5. Run terraform plan or terraform apply
terraform {
  required_providers {
    bytebase = {
      version = "0.0.1"
      # The source is only used in this local example.
      source = "registry.terraform.io/bytebase/bytebase"
    }
  }
}

provider "bytebase" {
  # You need to replace the email and password with your own bytebase account.
  email        = "ed+dev@bytebase.com"
  password     = "ed"
  bytebase_url = "http://localhost:8080/v1"
}

# Create a new environment named "dev"
resource "bytebase_environment" "dev" {
  name = "dev"
}

# Print the new environment
output "staging_environment" {
  value = bytebase_environment.dev
}
