
resource "bytebase_setting" "external_approval" {
  name = "bb.workspace.approval.external"

  external_approval_nodes {
    nodes {
      id       = "9e150339-f014-4835-83d7-123aeb1895ba"
      title    = "Example node"
      endpoint = "https://example.com"
    }

    nodes {
      id       = "49a976be-50de-4541-b2d3-f2e32f8e41ef"
      title    = "Example node 2"
      endpoint = "https://example.com"
    }
  }
}

resource "bytebase_setting" "approval_flow" {
  name = "bb.workspace.approval"
  approval_flow {
    rules {
      flow {
        title       = "DBA -> OWNER"
        description = "Need DBA and workspace owner approval"
        creator     = "users/support@bytebase.com"

        # Approval flow following the step order.
        steps {
          type = "GROUP"
          node = "WORKSPACE_DBA"
        }

        steps {
          type = "GROUP"
          node = "WORKSPACE_OWNER"
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
