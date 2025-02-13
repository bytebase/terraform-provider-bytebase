terraform {
  required_providers {
    bytebase = {
      version = "1.0.16"
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

data "bytebase_setting" "workspace_profile" {
  name = "bb.workspace.profile"
}

data "bytebase_setting" "classification" {
  name = "bb.workspace.data-classification"
}

data "bytebase_setting" "semantic_types" {
  name = "bb.workspace.semantic-types"
}

output "approval_flow" {
  value = data.bytebase_setting.approval_flow
}

output "external_approval" {
  value = data.bytebase_setting.external_approval
}

output "workspace_profile" {
  value = data.bytebase_setting.workspace_profile
}

output "classification" {
  value = data.bytebase_setting.classification
}

output "semantic_types" {
  value = data.bytebase_setting.semantic_types
}
