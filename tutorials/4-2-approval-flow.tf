resource "bytebase_setting" "approval_flow" {
  name = "settings/WORKSPACE_APPROVAL"

  approval_flow {
    rules {
      flow {
        title       = "Project Owner → DBA → Admin"
        description = "Need DBA and workspace admin approval"
        roles = [
          "roles/projectOwner",
          "roles/workspaceDBA",
          "roles/workspaceAdmin"
        ]
      }
      source    = "CHANGE_DATABASE"
      condition = "request.risk >= 100"
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
