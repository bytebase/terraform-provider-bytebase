terraform {
  required_version = ">= 1.11"
  required_providers {
    bytebase = {
      source  = "bytebase/bytebase"
      version = "3.25.0"
    }
  }
}

provider "bytebase" {}

variable "sql_review_enabled" {
  type        = bool
  default     = true
  description = "Whether the workspace permits AI SQL review."
}

variable "review_policy_content" {
  type        = string
  default     = null
  nullable    = true
  description = "Optional custom AI review instructions. Null creates no policy."
}

variable "review_policy_parent" {
  type        = string
  default     = null
  nullable    = true
  description = "An optional projects/{id} parent. Null selects the workspace."
}

resource "bytebase_setting" "ai" {
  name = "settings/AI"

  ai {
    sql_review_enabled = var.sql_review_enabled
  }
}

resource "bytebase_policy" "ai_review" {
  count = var.review_policy_content != null ? 1 : 0

  parent = var.review_policy_parent
  type   = "REVIEW_AI"

  review_ai_policy {
    content = var.review_policy_content
  }
}

data "bytebase_setting" "ai" {
  name = bytebase_setting.ai.name
}

output "ai_review" {
  value = {
    ai_enabled         = data.bytebase_setting.ai.ai[0].enabled
    sql_review_enabled = data.bytebase_setting.ai.ai[0].sql_review_enabled
    policy_name        = try(bytebase_policy.ai_review[0].name, null)
  }
}
