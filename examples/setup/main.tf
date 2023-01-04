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
  environment_id_dev  = "dev"
  environment_id_prod = "prod"
  instance_id_dev     = "dev-instance"
  instance_id_prod    = "prod-instance"
  role_name_dev       = "dev_role_test"
}

# Create a new environment named "dev"
resource "bytebase_environment" "dev" {
  name                     = local.environment_id_dev
  order                    = 0
  environment_tier_policy  = "UNPROTECTED"
  pipeline_approval_policy = "MANUAL_APPROVAL_NEVER"
  backup_plan_policy       = "UNSET"
}

# Create another environment named "prod"
resource "bytebase_environment" "prod" {
  name                     = local.environment_id_prod
  order                    = 1
  environment_tier_policy  = "PROTECTED"
  pipeline_approval_policy = "MANUAL_APPROVAL_BY_WORKSPACE_OWNER_OR_DBA"
  backup_plan_policy       = "DAILY"
}

# Create a new instance named "dev_instance_test"
# You can replace the parameters with your real instance
resource "bytebase_instance" "dev" {
  resource_id = local.instance_id_dev
  environment = bytebase_environment.dev.name
  title       = "dev instance"
  engine      = "POSTGRES"

  # You need to specific the data source
  data_sources {
    title    = "admin data source"
    type     = "ADMIN"
    username = "ecmadao"
    host     = "127.0.0.1"
    port     = "5432"
  }

  # And you can add another data_sources with RO type
  data_sources {
    title    = "read-only data source"
    type     = "READ_ONLY"
    username = "<The connection user name>"
    password = "<The connection user password>"
    host     = "192.168.0.1"
    port     = "1234"
  }
}

# Create a new instance named "prod_instance_test"
resource "bytebase_instance" "prod" {
  resource_id = local.instance_id_prod
  environment = bytebase_environment.prod.name
  title       = "prod instance"
  engine      = "POSTGRES"

  # You need to specific the data source
  data_sources {
    title    = "admin data source"
    type     = "ADMIN"
    username = "<The connection user name>"
    password = "<The connection user password>"
    host     = "127.0.0.1"
    port     = "5432"
  }
}

# Create a new role named "dev_role_test" in the instance "dev_instance_test"
resource "bytebase_database_role" "dev" {
  name        = local.role_name_dev
  instance    = bytebase_instance.dev.resource_id
  environment = bytebase_instance.dev.environment

  password         = "123456"
  connection_limit = 10
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
