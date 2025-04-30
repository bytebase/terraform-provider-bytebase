terraform {
  required_providers {
    bytebase = {
      version = "3.6.0"
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

data "bytebase_iam_policy" "workspace_iam" {
  parent = "workspaces/-"
}

output "workspace_iam" {
  value = data.bytebase_iam_policy.workspace_iam
}

data "bytebase_iam_policy" "project_iam" {
  parent = "projects/project-sample"
}

output "project_iam" {
  value = data.bytebase_iam_policy.project_iam
}
