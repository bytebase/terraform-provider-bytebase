terraform {
  required_providers {
    bytebase = {
      version = "0.0.3"
      # For local development, please use "terraform.local/bytebase/bytebase" instead
      source = "registry.terraform.io/bytebase/bytebase"
    }
  }
}

variable "environment_name" {
  type    = string
  default = ""
}

# List all environment
data "bytebase_environment_list" "all" {}

output "all_environments" {
  value = data.bytebase_environment_list.all.environments
}

# Only returns specific environment
output "environment" {
  value = {
    for environment in data.bytebase_environment_list.all.environments :
    environment.id => environment
    if environment.name == var.environment_name
  }
}
