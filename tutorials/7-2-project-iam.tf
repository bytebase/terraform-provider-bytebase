resource "bytebase_iam_policy" "project_iam" {
  depends_on = [
    bytebase_group.developers,
    bytebase_user.workspace_dba1,
    bytebase_user.workspace_dba2
  ]

  parent = bytebase_project.project-two.name

  iam_policy {

    binding {
      role = "roles/projectOwner"
      members = [
        format("user:%s", bytebase_user.workspace_dba1.email),
        format("user:%s", bytebase_user.workspace_dba2.email)
      ]
    }

    binding {
      role = "roles/projectDeveloper"
      members = [
        "allUsers",
        format("group:%s", bytebase_group.developers.email)
      ]
    }

    binding {
      role = "roles/sqlEditorUser"
      members = [
        format("group:%s", bytebase_group.developers.email)
      ]
      condition {
        database         = "instances/prod-sample-instance/databases/hr_prod"
        schema           = "public"
        tables           = ["employee","salary"]
        expire_timestamp = "2027-07-10T16:17:49Z"
      }
    }

  }
}