terraform {
  required_providers {
    bytebase = {
      version = "0.0.5"
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

# Create a new environment named "dev_test"
resource "bytebase_environment" "dev" {
  name                     = local.environment_name_dev
  order                    = 0
  environment_tier_policy  = "UNPROTECTED"
  pipeline_approval_policy = "MANUAL_APPROVAL_NEVER"
  backup_plan_policy       = "UNSET"
}

# Create another environment named "prod_test"
resource "bytebase_environment" "prod" {
  name                     = local.environment_name_prod
  order                    = 1
  environment_tier_policy  = "PROTECTED"
  pipeline_approval_policy = "MANUAL_APPROVAL_BY_WORKSPACE_OWNER_OR_DBA"
  backup_plan_policy       = "DAILY"
}

# List all environment
data "bytebase_environment_list" "all" {
  depends_on = [
    bytebase_environment.dev,
    bytebase_environment.prod
  ]
}

output "all_environments" {
  value = data.bytebase_environment_list.all.environments
}

// Find a specific environment by the name
data "bytebase_environment" "find_dev_env" {
  name = local.environment_name_dev
  depends_on = [
    bytebase_environment.dev
  ]
}

output "dev_environment" {
  value = data.bytebase_environment.find_dev_env
}
