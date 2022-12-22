# Examples for query the instances
terraform {
  required_providers {
    bytebase = {
      version = "0.0.6-alpha"
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
  instance_name_dev  = "dev_instance_test"
  instance_name_prod = "prod_instance_test"
}

# List all instance
data "bytebase_instance_list" "all" {}

output "all_instances" {
  value = data.bytebase_instance_list.all.instances
}

# Find a specific instance by name
data "bytebase_instance" "dev" {
  name = local.instance_name_dev
}


output "dev_instance" {
  value = data.bytebase_instance.dev
}

data "bytebase_instance" "prod" {
  name = local.instance_name_prod
}


output "prod_instance" {
  value = data.bytebase_instance.prod
}
