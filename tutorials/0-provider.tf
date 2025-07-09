terraform {
  required_providers {
    bytebase = {
      version = "3.8.0"
      # For local development, please use "terraform.local/bytebase/bytebase" instead
      source = "registry.terraform.io/bytebase/bytebase"
    }
  }
}

provider "bytebase" {
  service_account = "tf@service.bytebase.com"
  service_key     = "bbs_EttGNSWLLGCqWB9Yev0I"
  url = "https://valid-just-tadpole.ngrok-free.app"
}