resource "bytebase_setting" "approval_flow" {
  name = "bb.workspace.approval"
  approval_flow {
    rules {
      flow {
        title       = "Project Owner -> DBA -> Admin"
        description = "Need DBA and workspace admin approval"
        creator     = "users/support@bytebase.com"

        # Approval flow following the step order.
        steps {
          role = "roles/projectOwner"
        }

        steps {
          role = "roles/workspaceDBA"
        }

        steps {
          role = "roles/workspaceAdmin"
        }
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
