terraform {
  required_providers {
    bytebase = {
      version = "3.13.0"
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

    password_restriction {
      min_length                             = 8
      require_number                         = true
      require_reset_password_for_first_login = true
    }
  }
}

resource "bytebase_policy" "query_data_policy" {
  parent = "workspaces/-"
  type   = "DATA_QUERY"
  query_data_policy {
    maximum_result_size = 200 * 1024 * 1024 # 200MB
    maximum_result_rows = 100
    disable_export      = false
    timeout_in_seconds  = 60 # 60 seconds
  }
}
