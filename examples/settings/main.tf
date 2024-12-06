terraform {
  required_providers {
    bytebase = {
      version = "1.0.3"
      # For local development, please use "terraform.local/bytebase/bytebase" instead
      source = "terraform.local/bytebase/bytebase"
    }
  }
}

provider "bytebase" {
  # You need to replace the account and key with your Bytebase service account.
  service_account = "ed@bytebase.com"
  service_key     = "12345678A!"
  # The Bytebase service URL. You can use the external URL in production.
  # Check the docs about external URL: https://www.bytebase.com/docs/get-started/install/external-url
  url = "http://localhost:8080"
}

data "bytebase_setting" "approval_flow" {
  name = "bb.workspace.approval"
}
