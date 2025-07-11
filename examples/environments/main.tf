terraform {
  required_providers {
    bytebase = {
      version = "3.8.2"
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

# List all environment
data "bytebase_setting" "environments" {
  name = "settings/ENVIRONMENT"
}

output "all_environments" {
  value = data.bytebase_setting.environments
}

data "bytebase_environment" "prod" {
  resource_id = "prod"
}

output "prod_environment" {
  value = data.bytebase_environment.prod
}
