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
  }
}
