terraform {
  required_version = ">= 1.11"
  required_providers {
    bytebase = {
      version = "3.23.0"
      source  = "terraform.local/bytebase/bytebase"
    }
  }
}

variable "bytebase_url" {
  type = string
}

variable "workload_identity_email" {
  type = string
}

variable "workload_identity_token_file" {
  type = string
}

provider "bytebase" {
  url                          = var.bytebase_url
  workload_identity_email      = var.workload_identity_email
  workload_identity_token_file = var.workload_identity_token_file
}
