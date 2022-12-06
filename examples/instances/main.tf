terraform {
  required_providers {
    bytebase = {
      version = "0.0.3"
      # For local development, please use "terraform.local/bytebase/bytebase" instead
      source = "registry.terraform.io/bytebase/bytebase"
    }
  }
}

variable "instance_name" {
  type    = string
  default = ""
}

# List all instance
data "bytebase_instance_list" "all" {}

output "all_instances" {
  value = data.bytebase_instance_list.all.instances
}

# Only returns specific instance
output "instance" {
  value = {
    for instance in data.bytebase_instance_list.all.instances :
    instance.id => instance
    if instance.name == var.instance_name
  }
}
