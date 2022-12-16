# Setup the data for examples
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
  environment_name_dev  = "dev_test"
  environment_name_prod = "prod_test"
  instance_name_dev     = "dev_instance_test"
  instance_name_prod    = "prod_instance_test"
}

# Create a new environment named "dev_test"
resource "bytebase_environment" "dev" {
  name                     = local.environment_name_dev
  order                    = 0
  environment_tier_policy  = "UNPROTECTED"
  pipeline_approval_policy = "MANUAL_APPROVAL_NEVER"
  backup_plan_policy       = "UNSET"
}

# Create another environment named "prod_test"
resource "bytebase_environment" "prod" {
  name                     = local.environment_name_prod
  order                    = 1
  environment_tier_policy  = "PROTECTED"
  pipeline_approval_policy = "MANUAL_APPROVAL_BY_WORKSPACE_OWNER_OR_DBA"
  backup_plan_policy       = "DAILY"
}

# Create a new instance named "dev_instance_test"
resource "bytebase_instance" "dev" {
  name        = local.instance_name_dev
  engine      = "POSTGRES"
  host        = "127.0.0.1"
  port        = 5432
  environment = bytebase_environment.dev.name

  # You need to specific the data source
  data_source_list {
    name = "admin data source"
    type = "ADMIN"
  }

  # And you can add another data_source_list with RO type
  data_source_list {
    name          = "read-only data source"
    type          = "RO"
    username      = "<The connection user name>"
    password      = "<The connection user password>"
    host_override = "192.168.0.1"
    port_override = "1234"
  }
}

# Create a new instance named "prod_instance_test"
resource "bytebase_instance" "prod" {
  name        = local.instance_name_prod
  engine      = "POSTGRES"
  host        = "127.0.0.1"
  port        = 5432
  environment = bytebase_environment.prod.name

  # You need to specific the data source
  data_source_list {
    name     = "admin data source"
    type     = "ADMIN"
    username = "<The connection user name>"
    password = "<The connection user password>"
  }
}
