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
