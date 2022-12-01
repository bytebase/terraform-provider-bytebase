terraform {
  required_providers {
    bytebase = {
      version = "0.0.3"
      # The source is only used for local development.
      # To use it in production, please replace it with "registry.terraform.io/bytebase/bytebase"
      source = "terraform.local/bytebase/bytebase"
    }
  }
}

variable "instance_name" {
  type    = string
  default = ""
}

# List all instance
data "bytebase_instances" "all" {}

output "all_instances" {
  value = data.bytebase_instances.all.instances
}

# Only returns specific instance
output "instance" {
  value = {
    for instance in data.bytebase_instances.all.instances :
    instance.id => instance
    if instance.name == var.instance_name
  }
}
