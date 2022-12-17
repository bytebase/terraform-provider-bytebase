terraform {
  required_providers {
    bytebase = {
      version = "0.0.5"
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
  instance_name_dev = "dev_instance_test"
}

# Find the instance
data "bytebase_instance" "dev" {
  name = local.instance_name_dev
}

# Create a new role in the instance
resource "bytebase_database_role" "test_role" {
  name             = "test_role"
  instance         = data.bytebase_instance.dev.name
  password         = "123456"
  connection_limit = 99
  valid_until      = "2022-12-31T00:00:00+08:00"

  attribute {
    super_user  = true
    no_inherit  = true
    create_role = true
    create_db   = false
    can_login   = true
    replication = true
    bypass_rls  = true
  }
}

output "test_role" {
  value = bytebase_database_role.test_role
}
