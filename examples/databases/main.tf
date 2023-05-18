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

# List all databases in all instances
data "bytebase_database_list" "all" {}

output "all_databases" {
  value = data.bytebase_database_list.all
}

# Filter databases by project
data "bytebase_database_list" "sample" {
  project = "project-sample"
}

output "databases_in_sample_project" {
  value = data.bytebase_database_list.sample
}

# List all instances in all environments
data "bytebase_database" "employee" {
  name     = "employee"
  instance = "postgres-sample"
}

output "employee_databases" {
  value = data.bytebase_database.employee
}
