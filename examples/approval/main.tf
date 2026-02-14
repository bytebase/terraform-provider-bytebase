terraform {
  required_providers {
    bytebase = {
      version = "3.15.2"
      # For local development, please use "terraform.local/bytebase/bytebase" instead
      source = "terraform.local/bytebase/bytebase"
    }
  }
}

provider "bytebase" {
  # You need to replace the account and key with your Bytebase service account.
  service_account = "ed@bytebase.com"
  service_key     = "12345678"
  # The Bytebase service URL. You can use the external URL in production.
  # Check the docs about external URL: https://www.bytebase.com/docs/get-started/install/external-url
  url = "http://localhost:8080"
}

resource "bytebase_setting" "approval_flow" {
  name = "settings/WORKSPACE_APPROVAL"
  approval_flow {
    rules {
      flow {
        title       = "Project Owner -> DBA -> Admin"
        description = "Need DBA and workspace admin approval"

        # Approval flow following the step order.
        roles = [
          "roles/projectOwner",
          "roles/workspaceDBA",
          "roles/workspaceAdmin"
        ]
      }

      source    = "CHANGE_DATABASE"
      condition = "resource.environment_id == \"prod\" && statement.affected_rows >= 100"
    }

    rules {
      flow {
        title = "Project Owner review"

        # Approval flow following the step order.
        roles = [
          "roles/projectOwner"
        ]
      }

      source    = "EXPORT_DATA"
      condition = "resource.environment_id == \"prod\" && resource.table_name == \"employee\""
    }

    rules {
      flow {
        title = "Fallback rule 1"

        # Approval flow following the step order.
        roles = [
          "roles/projectOwner"
        ]
      }

      condition = "resource.project_id == \"new-project\""
    }

    rules {
      flow {
        title = "Fallback rule 2"

        # Approval flow following the step order.
        roles = [
          "roles/workspaceDBA"
        ]
      }

      condition = "true"
    }
  }
}
