# Create or update sample project, and grant roles for users and groups.
resource "bytebase_project" "sample_project" {
  depends_on = [
    bytebase_user.workspace_dba,
    bytebase_user.project_developer,
    bytebase_group.developers,
    bytebase_instance.prod
  ]

  resource_id = local.project_id
  title       = "Sample project"
  key         = "SAMM"

  members {
    member = format("user:%s", bytebase_user.workspace_dba.email)
    role   = "roles/projectOwner"
  }

  members {
    member = format("group:%s", bytebase_group.developers.email)
    role   = "roles/projectDeveloper"
  }

  members {
    member = format("user:%s", bytebase_user.project_developer.email)
    role   = "roles/projectExporter"
    condition {
      database         = "instances/test-sample-instance/databases/employee"
      tables           = ["dept_emp", "dept_manager"]
      row_limit        = 10000
      expire_timestamp = "2027-03-09T16:17:49Z"
    }
  }

  databases = bytebase_instance.prod.databases
}
