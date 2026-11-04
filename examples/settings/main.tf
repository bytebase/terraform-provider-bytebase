terraform {
  required_version = ">= 1.11"
  required_providers {
    bytebase = {
      version = "3.24.0"
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

variable "ai_api_key" {
  type      = string
  sensitive = true
}

variable "ai_model" {
  type        = string
  description = "An OpenAI model that supports the Chat Completions endpoint."
}

# Self-hosted AI configuration. The API key is sensitive but stored in state.
# On Bytebase SaaS, omit this resource and the AI input variables.
resource "bytebase_setting" "ai" {
  name = "settings/AI"

  ai {
    enabled  = true
    provider = "OPEN_AI"
    endpoint = "https://api.openai.com/v1/chat/completions"
    api_key  = var.ai_api_key
    model    = var.ai_model
  }
}

# READ_ONLY allows metadata inspection and read-only SQL within the user's RBAC.
# Other supported modes are DISABLED and READ_WRITE.
resource "bytebase_setting" "mcp" {
  name = "settings/MCP"

  mcp {
    capability = "READ_ONLY"
  }
}

data "bytebase_setting" "ai" {
  # On Bytebase SaaS, use "settings/AI" to read the server-managed settings.
  name = bytebase_setting.ai.name
}

data "bytebase_setting" "mcp" {
  name = bytebase_setting.mcp.name
}

data "bytebase_setting" "approval_flow" {
  name = "settings/WORKSPACE_APPROVAL"
}

data "bytebase_setting" "workspace_profile" {
  name = "settings/WORKSPACE_PROFILE"
}

data "bytebase_setting" "classification" {
  name = "settings/DATA_CLASSIFICATION"
}

data "bytebase_setting" "semantic_types" {
  name = "settings/SEMANTIC_TYPES"
}

output "approval_flow" {
  value = data.bytebase_setting.approval_flow
}

output "workspace_profile" {
  value = data.bytebase_setting.workspace_profile
}

# Configure this timeout in ../setup/main.tf; zero disables the timeout.
output "inactive_session_timeout_in_seconds" {
  value = data.bytebase_setting.workspace_profile.workspace_profile[0].inactive_session_timeout_in_seconds
}

output "ai" {
  value = {
    enabled  = data.bytebase_setting.ai.ai[0].enabled
    provider = data.bytebase_setting.ai.ai[0].provider
    model    = data.bytebase_setting.ai.ai[0].model
    version  = data.bytebase_setting.ai.ai[0].version
  }
}

output "mcp_capability" {
  value = data.bytebase_setting.mcp.mcp[0].capability
}

output "classification" {
  value = data.bytebase_setting.classification
}

output "semantic_types" {
  value = data.bytebase_setting.semantic_types
}
