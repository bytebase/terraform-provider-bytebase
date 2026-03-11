resource "bytebase_iam_policy" "workspace_iam" {
  depends_on = [
    bytebase_user.workspace_admin,
    bytebase_service_account.tf_service_account,
    bytebase_workload_identity.github_ci,
    bytebase_user.workspace_dba1,
    bytebase_user.workspace_dba2,
    bytebase_group.qa
  ]

  # parent defaults to workspace when not specified.

  iam_policy {

    binding {
      role = "roles/workspaceAdmin"
      members = [
        format("user:%s", bytebase_user.workspace_admin.email),
      ]
    }

    binding {
      role = "roles/workspaceDBA"
      members = [
        format("user:%s", bytebase_user.workspace_dba1.email),
        format("user:%s", bytebase_user.workspace_dba2.email),
        format("serviceAccount:%s", bytebase_service_account.tf_service_account.email),
        format("workloadIdentity:%s", bytebase_workload_identity.github_ci.email),
      ]
    }

    binding {
      role = "roles/workspaceMember"
      members = [
        "allUsers"
      ]
    }

    binding {
      role = "roles/projectViewer"
      members = [
        format("group:%s", bytebase_group.qa.email),
      ]
    }
  }
}
