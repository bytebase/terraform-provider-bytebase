terraform {
  required_providers {
    bytebase = {
      version = "3.13.1"
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

data "bytebase_project" "sample_project" {
  resource_id = "sample-project"
}

data "bytebase_database_group_list" "all" {
  depends_on = [data.bytebase_project.sample_project]
  project    = data.bytebase_project.sample_project.name
}

output "database_group_list" {
  value = data.bytebase_database_group_list.all
}

resource "bytebase_database_group" "databases_in_test" {
  depends_on = [data.bytebase_project.sample_project]

  resource_id = "databases-in-test"
  project     = data.bytebase_project.sample_project.name
  title       = "Databases in test env"
  condition   = "resource.environment_id == \"test\""
}

data "bytebase_database_group" "databases_in_test" {
  depends_on = [bytebase_database_group.databases_in_test]

  resource_id = bytebase_database_group.databases_in_test.resource_id
  project     = data.bytebase_project.sample_project.name
}

output "database_group" {
  value = data.bytebase_database_group.databases_in_test
}
