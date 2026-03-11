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
        title = "DBA -> Admin"

        # Approval flow following the step order.
        roles = [
          "roles/workspaceDBA",
          "roles/workspaceAdmin"
        ]
      }

      source    = "CREATE_DATABASE"
      condition = "resource.environment_id == \"prod\" || resource.instance_id == \"prod-sample-instance\""
    }

    rules {
      flow {
        title = "Project Owner review"

        # Approval flow following the step order.
        roles = [
          "roles/projectOwner"
        ]
      }

      source    = "REQUEST_ROLE"
      condition = "resource.role == \"roles/projectOwner\""
    }

    rules {
      flow {
        title = "Project Owner review"

        # Approval flow following the step order.
        roles = [
          "roles/projectOwner"
        ]
      }

      source    = "REQUEST_ACCESS"
      condition = "resource.unmask == true"
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
        title = "Fallback rule"

        # Approval flow following the step order.
        roles = [
          "roles/workspaceDBA"
        ]
      }

      condition = "true"
    }
  }
}
