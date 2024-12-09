terraform {
  required_providers {
    bytebase = {
      version = "1.0.4"
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

data "bytebase_policy" "masking_policy" {
  parent = "instances/test-sample-instance/databases/employee"
  type   = "MASKING"
}

data "bytebase_policy" "masking_exception_policy" {
  parent = "projects/project-sample"
  type   = "MASKING_EXCEPTION"
}

output "masking_policy" {
  value = data.bytebase_policy.masking_policy
}

output "masking_exception_policy" {
  value = data.bytebase_policy.masking_exception_policy
}