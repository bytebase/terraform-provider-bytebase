# Examples for query the environments
terraform {
  required_providers {
    bytebase = {
      version = "0.0.6-beta"
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
  environment_name_dev  = "dev_test"
  environment_name_prod = "prod_test"
}

# List all environment
data "bytebase_environment_list" "all" {}

output "all_environments" {
  value = data.bytebase_environment_list.all.environments
}

// Find a specific environment by the name
data "bytebase_environment" "dev" {
  name = local.environment_name_dev
}

output "dev_environment" {
  value = data.bytebase_environment.dev
}

data "bytebase_environment" "prod" {
  name = local.environment_name_prod
}

output "prod_environment" {
  value = data.bytebase_environment.prod
}
