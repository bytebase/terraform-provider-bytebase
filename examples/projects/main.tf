# Examples for query the project
terraform {
  required_providers {
    bytebase = {
      version = "3.8.1"
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

locals {
  project_id = "project-sample"
}

# List all projects
data "bytebase_project_list" "all" {
  query           = "sample"
  exclude_default = true
}

output "all_projects" {
  value = data.bytebase_project_list.all
}

# Find a specific project by name
data "bytebase_project" "sample_project" {
  resource_id = local.project_id
}

output "sample_project" {
  value = data.bytebase_project.sample_project
}
