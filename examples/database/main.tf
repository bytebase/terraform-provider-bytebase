# Examples for query the database
terraform {
  required_providers {
    bytebase = {
      version = "1.0.23"
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

data "bytebase_database_list" "all" {
  parent      = "workspaces/-"
  environment = "environments/test"
  project     = "projects/sample-project"
}

output "all_databases" {
  value = data.bytebase_database_list.all
}

data "bytebase_database_catalog" "employee" {
  database = "instances/test-sample-instance/databases/employee"
}

output "employee_catalog" {
  value = data.bytebase_database_catalog.employee
}
