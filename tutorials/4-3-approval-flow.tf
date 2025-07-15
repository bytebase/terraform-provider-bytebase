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