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

data "bytebase_setting" "approval_flow" {
  name = "bb.workspace.approval"
}

data "bytebase_setting" "external_approval" {
  name = "bb.workspace.approval.external"
}

output "approval_flow" {
  value = data.bytebase_setting.approval_flow
}

output "external_approval" {
  value = data.bytebase_setting.external_approval
}