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

data "bytebase_environment" "find_env" {
  name = var.environment_name
}

output "environment" {
  value = data.bytebase_environment.find_env
}
