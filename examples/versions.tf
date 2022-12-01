terraform {
  required_providers {
    bytebase = {
      version = "0.0.3"
      # For local development, please use "terraform.local/bytebase/bytebase" instead
      source = "registry.terraform.io/bytebase/bytebase"
    }
  }
}
