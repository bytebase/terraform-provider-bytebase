provider "bytebase" {
  # You need to replace the account and key with your Bytebase service account.
  service_account = "terraform@service.bytebase.com"
  service_key     = "bbs_qHX6CswQ1nMMELSCc2lk"
  # The Bytebase service URL. You can use the external URL in production.
  # Check the docs about external URL: https://www.bytebase.com/docs/get-started/install/external-url
  url = "https://bytebase.example.com"
}

locals {
  environment_name_dev  = "dev"
  environment_name_prod = "prod"
  instance_name         = "dev-instance"
}

# Create a new environment named "dev"
resource "bytebase_environment" "dev" {
  name                     = local.environment_name_dev
  order                    = 0
  environment_tier_policy  = "UNPROTECTED"
  pipeline_approval_policy = "MANUAL_APPROVAL_NEVER"
  backup_plan_policy       = "UNSET"
}

# Create another environment named "prod"
resource "bytebase_environment" "prod" {
  name                     = local.environment_name_prod
  order                    = 1
  environment_tier_policy  = "PROTECTED"
  pipeline_approval_policy = "MANUAL_APPROVAL_BY_WORKSPACE_OWNER_OR_DBA"
  backup_plan_policy       = "DAILY"
}

# Print the new environment
output "staging_environment" {
  value = bytebase_environment.dev
}

# Create a new instance named "dev-instance"
resource "bytebase_instance" "dev_instance" {
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

# Print the new instance
output "dev_instance" {
  value = bytebase_instance.dev_instance
}

# import environments module and filter by environment name
module "environment" {
  source           = "./environments"
  environment_name = local.environment_name_dev
  # Make sure the module exec after the "dev" environment is created
  depends_on = [
    bytebase_environment.dev
  ]
}

output "environment" {
  value = module.environment.environment
}

# import instances module and filter by instance name
module "instance" {
  source        = "./instances"
  instance_name = local.instance_name
  # Make sure the module exec after the "dev-instance" instance is created
  depends_on = [
    bytebase_instance.dev_instance
  ]
}

output "instance" {
  value = module.instance.instance
}
