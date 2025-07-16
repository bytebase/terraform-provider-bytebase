resource "bytebase_iam_policy" "workspace_iam" {
  depends_on = [
    bytebase_user.workspace_admin,
    bytebase_user.tf_service_account,
    bytebase_user.workspace_dba1,
    bytebase_user.workspace_dba2,
    bytebase_group.qa
  ]

  parent = "workspaces/-"

  iam_policy {

    binding {
      role = "roles/workspaceAdmin"
      members = [
        format("user:%s", bytebase_user.workspace_admin.email),
        format("user:%s", bytebase_user.tf_service_account.email),
      ]
    }

    binding {
      role = "roles/workspaceDBA"
      members = [
        format("user:%s", bytebase_user.workspace_dba1.email),
        format("user:%s", bytebase_user.workspace_dba2.email)
      ]
    }

    binding {
      role = "roles/workspaceMember"
      members = [
        format("user:%s", bytebase_user.dev1.email),
        format("user:%s", bytebase_user.dev2.email),
        format("user:%s", bytebase_user.dev3.email),
        format("user:%s", bytebase_user.qa1.email),
        format("user:%s", bytebase_user.qa2.email)
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