terraform {
  required_providers {
    bytebase = {
      version = "0.0.3"
      source  = "terraform.local/bytebase/bytebase"
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