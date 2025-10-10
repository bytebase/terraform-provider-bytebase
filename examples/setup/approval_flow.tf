resource "bytebase_setting" "approval_flow" {
  name = "settings/WORKSPACE_APPROVAL"
  approval_flow {
    rules {
      flow {
        id          = "bb.project-owner-workspace-dba-workspace-admin"
        title       = "Project Owner -> DBA -> Admin"
        description = "Need DBA and workspace admin approval"

        # Approval flow following the step order.
        roles = [
          "roles/projectOwner",
          "roles/workspaceDBA",
          "roles/workspaceAdmin"
        ]
      }

      # Match any condition will trigger this approval flow.
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
