terraform {
  required_providers {
    bytebase = {
      version = "0.0.1"
      source  = "registry.terraform.io/bytebase/bytebase"
    }
  }
}

provider "bytebase" {
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
