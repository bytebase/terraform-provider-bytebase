terraform {
  required_providers {
    bytebase = {
      version = "3.8.7"
      # For local development, please use "terraform.local/bytebase/bytebase" instead
      source = "registry.terraform.io/bytebase/bytebase"
    }
  }
}

# Correspond to the sample data Bytebase generates during onboarding.
locals {
  service_account     = "terraform@service.bytebase.com"
  environment_id_test = "test"
  environment_id_prod = "prod"
  instance_id_test    = "test-sample-instance"
  instance_id_prod    = "prod-sample-instance"
  project_id          = "project-sample"
}


provider "bytebase" {
  # You need to replace the account and key with your Bytebase service account.
  service_account = local.service_account
  service_key     = "bbs_BxVIp7uQsARl8nR92ZZV"
  # The Bytebase service URL. You can use the external URL in production.
  # Check the docs about external URL: https://www.bytebase.com/docs/get-started/install/external-url
  url = "https://bytebase.example.com"
}

resource "bytebase_setting" "workspace_profile" {
  name = "settings/WORKSPACE_PROFILE"

  workspace_profile {
    external_url = "https://bytebase.example.com"
    domains      = ["bytebase.com"]
  }
}
