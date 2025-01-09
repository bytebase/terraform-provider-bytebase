terraform {
  required_providers {
    bytebase = {
      version = "1.0.8"
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

data "bytebase_vcs_provider" "github" {
  resource_id = "vcs-github"
}

data "bytebase_project" "sample_project" {
  resource_id = local.project_id
}

data "bytebase_vcs_connector" "github" {
  depends_on = [
    data.bytebase_project.sample_project
  ]
  resource_id = "connector-github"
  project     = data.bytebase_project.sample_project.name
}

output "vcs_provider_github" {
  value = data.bytebase_vcs_provider.github
}

output "vcs_connector_github" {
  value = data.bytebase_vcs_connector.github
}
