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
  environment_name_dev = "dev_env_test"
  instance_name        = "dev_instance_test"
}

# Create a new environment named "dev"
resource "bytebase_environment" "dev" {
  name                     = local.environment_name_dev
  order                    = 0
  environment_tier_policy  = "UNPROTECTED"
  pipeline_approval_policy = "MANUAL_APPROVAL_NEVER"
  backup_plan_policy       = "UNSET"
}

# Create a new instance named "dev_instance_test"
resource "bytebase_instance" "dev" {
  name        = local.instance_name
  engine      = "POSTGRES"
  host        = "127.0.0.1"
  port        = 5432
  environment = bytebase_environment.dev.name

  # You need to specific the data source
  data_source_list {
    name     = "admin data source"
    type     = "ADMIN"
    username = "<The connection user name>"
    password = "<The connection user password>"
  }

  # And you can add another data_source_list with RO type
  data_source_list {
    name     = "read-only data source"
    type     = "RO"
    username = "<The connection user name>"
    password = "<The connection user password>"
  }
}


# List all instance
data "bytebase_instance_list" "all" {
  depends_on = [
    bytebase_instance.dev
  ]
}

output "all_instances" {
  value = data.bytebase_instance_list.all.instances
}

# Find a specific instance by name
data "bytebase_instance" "find_instance" {
  name = local.instance_name
  depends_on = [
    bytebase_instance.dev
  ]
}


output "instance" {
  value = data.bytebase_instance.find_instance
}
