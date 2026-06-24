terraform {
  required_version = ">= 1.11"
  required_providers {
    bytebase = {
      version = "3.20.0"
      # For local development, please use "terraform.local/bytebase/bytebase" instead
      source = "registry.terraform.io/bytebase/bytebase"
    }
  }
}

provider "bytebase" {
  # You need to replace the account and key with your Bytebase service account.
  service_account = "terraform@service.bytebase.com"
  service_key     = "bbs_BxVIp7uQsARl8nR92ZZV"
  # The Bytebase service URL. You can use the external URL in production.
  # Check the docs about external URL: https://www.bytebase.com/docs/get-started/install/external-url
  url = "https://bytebase.example.com"
}

# Credentials are write-only. Source them from variables, a tfvars file, or a
# secrets backend; never inline them in committed HCL.

variable "slack_token" {
  type      = string
  sensitive = true
}

variable "feishu_app_id" {
  type = string
}

variable "feishu_app_secret" {
  type      = string
  sensitive = true
}

resource "bytebase_setting" "app_im" {
  name = "settings/APP_IM"

  app_im {
    slack {
      token = var.slack_token
    }

    feishu {
      app_id     = var.feishu_app_id
      app_secret = var.feishu_app_secret
    }

    # Omit a block to disable that IM integration on the next apply.
    # Other supported blocks: wecom, lark, dingtalk, teams.
  }
}

# Data source: surfaces which IM types are configured. Field values are always
# empty strings because the server strips them on GET.
data "bytebase_setting" "app_im" {
  name       = "settings/APP_IM"
  depends_on = [bytebase_setting.app_im]
}

output "configured_im_types" {
  value = {
    slack    = length(data.bytebase_setting.app_im.app_im[0].slack) > 0
    feishu   = length(data.bytebase_setting.app_im.app_im[0].feishu) > 0
    wecom    = length(data.bytebase_setting.app_im.app_im[0].wecom) > 0
    lark     = length(data.bytebase_setting.app_im.app_im[0].lark) > 0
    dingtalk = length(data.bytebase_setting.app_im.app_im[0].dingtalk) > 0
    teams    = length(data.bytebase_setting.app_im.app_im[0].teams) > 0
  }
}
