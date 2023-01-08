terraform {
  required_providers {
    bytebase = {
      version = "0.0.6-beta.3"
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
  role_name_dev      = "dev_role_test"
  instance_id_dev    = "dev-instance"
  environment_id_dev = "dev"
}

# Find the instance
data "bytebase_instance" "dev" {
  resource_id = local.instance_id_dev
  environment = local.environment_id_dev
}

# Find the role "dev_role_test" in the instance "dev_instance_test"
data "bytebase_instance_role" "dev" {
  name        = local.role_name_dev
  instance    = data.bytebase_instance.dev.resource_id
  environment = data.bytebase_instance.dev.environment
}

output "dev_role" {
  value = data.bytebase_instance_role.dev
}

# List all roles in the instance "dev_instance_test"
data "bytebase_instance_role_list" "all" {
  instance    = data.bytebase_instance.dev.resource_id
  environment = data.bytebase_instance.dev.environment
}

output "list_role" {
  value = data.bytebase_instance_role_list.all
}
