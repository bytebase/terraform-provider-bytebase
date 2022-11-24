terraform {
  required_providers {
    bytebase = {
      version = "0.0.2"
      source  = "terraform.local/bytebase/bytebase"
    }
  }
}

variable "environment_name" {
  type    = string
  default = ""
}

# List all environment
data "bytebase_environments" "all" {}

output "all_environments" {
  value = data.bytebase_environments.all.environments
}

# Only returns specific environment
output "environment" {
  value = {
    for environment in data.bytebase_environments.all.environments :
    environment.id => environment
    if environment.name == var.environment_name
  }
}
