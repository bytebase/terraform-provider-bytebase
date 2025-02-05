# Examples for query the environments
terraform {
  required_providers {
    bytebase = {
      version = "1.0.9"
      # For local development, please use "terraform.local/bytebase/bytebase" instead
      source = "registry.terraform.io/bytebase/bytebase"
    }
  }
}

provider "bytebase" {
  # You need to replace the account and key with your Bytebase service account.
  service_account = "terraform@service.bytebase.com"
  service_key     = "bbs_BxVIp7uQsARl8nR92ZZV"
  # The Bytebase service URL. You can use the external URL in production.
  # Check the docs about external URL: https://www.bytebase.com/docs/get-started/install/external-url
  url = "https://bytebase.example.com"
}

locals {
  environment_id_test = "test"
  environment_id_prod = "prod"
}

# List all environment
data "bytebase_environment_list" "all" {}

output "all_environments" {
  value = data.bytebase_environment_list.all
}

// Find a specific environment by the name
data "bytebase_environment" "test" {
  resource_id = local.environment_id_test
}

output "test_environment" {
  value = data.bytebase_environment.test
}

data "bytebase_environment" "prod" {
  resource_id = local.environment_id_prod
}

output "prod_environment" {
  value = data.bytebase_environment.prod
}
