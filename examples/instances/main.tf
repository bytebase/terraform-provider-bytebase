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

data "bytebase_instance" "find_instance" {
  name = var.instance_name
}


output "instance" {
  value = data.bytebase_instance.find_instance
}
