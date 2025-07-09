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
  service_account = "tf@service.example.com"
  service_key     = "bbs_xxx"
  url = "https://xxx.xxx.xxx"
}

# Step 1: Workspace profile configuration
resource "bytebase_setting" "workspace_profile" {
  name = "settings/WORKSPACE_PROFILE"

  workspace_profile {
    disallow_signup           = true
    domains                   = ["example.com"]
    enforce_identity_domain  = true
    external_url             = "https://your-bytebase-instance.com"
    # Optional:
    # disallow_password_signin = false
  }
}

# Step 2: Approval flow settings
resource "bytebase_setting" "approval_flow" {
  name = "settings/WORKSPACE_APPROVAL"

  approval_flow {
    rules {
      flow {
        title       = "Project Owner → DBA → Admin"
        description = "Need DBA and workspace admin approval"

        steps { role = "roles/projectOwner" }
        steps { role = "roles/workspaceDBA" }
        steps { role = "roles/workspaceAdmin" }
      }
      conditions {
        source = "DML"
        level  = "MODERATE"
      }
      conditions {
        source = "DDL"
        level  = "HIGH"
      }
    }
  }
}

# Step 3: Risk management policies
resource "bytebase_risk" "dml_moderate" {
  title     = "DML Moderate Risk"
  source    = "DML"
  level     = 200
  active    = true
  condition = "environment_id == \"prod\" && affected_rows >= 100"
}

resource "bytebase_risk" "ddl_high" {
  title     = "DDL High Risk"
  source    = "DDL"
  level     = 300
  active    = true
  condition = "environment_id == \"prod\""
}
