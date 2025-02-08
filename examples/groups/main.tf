terraform {
  required_providers {
    bytebase = {
      version = "1.0.14"
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

data "bytebase_group_list" "all" {
}

output "all_groups" {
  value = data.bytebase_group_list.all
}

data "bytebase_group" "sample" {
  name = "groups/group@bytebase.com"
}

output "sample_group" {
  value = data.bytebase_group.sample
}
