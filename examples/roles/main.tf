terraform {
  required_providers {
    bytebase = {
      version = "0.0.9"
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
  role_name        = "role_test_terraform"
  instance_id_test = "test-sample-instance"
}

# Find the instance
data "bytebase_instance" "test" {
  resource_id = local.instance_id_test
}

# Find the role "role_test_terraform" in the instance "test-sample-instance"
data "bytebase_instance_role" "test" {
  name     = local.role_name
  instance = data.bytebase_instance.test.resource_id
}

output "role" {
  value = data.bytebase_instance_role.test
}

# List all roles in the instance "test-sample-instance"
data "bytebase_instance_role_list" "all" {
  instance = data.bytebase_instance.test.resource_id
}

output "list_role" {
  value = data.bytebase_instance_role_list.all
}
