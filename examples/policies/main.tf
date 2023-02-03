# Examples for query the instances
terraform {
  required_providers {
    bytebase = {
      version = "0.0.7"
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
}

# Find deployment approval policy in test environment.
data "bytebase_policy" "deployment_approval" {
  environment = local.environment_id_test
  type        = "DEPLOYMENT_APPROVAL"
}

output "deployment_approval_policy" {
  value = data.bytebase_policy.deployment_approval
}

# Find SQL review policy in test environment.
data "bytebase_policy" "sql_review" {
  environment = local.environment_id_test
  type        = "SQL_REVIEW"
}

output "sql_review_policy" {
  value = data.bytebase_policy.sql_review
}

# List policies in test environment.
data "bytebase_policy_list" "test_env_policies" {
  environment = local.environment_id_test
}

output "test_env_policies" {
  value = data.bytebase_policy_list.test_env_policies
}
