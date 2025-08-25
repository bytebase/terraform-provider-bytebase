# Workspace level IAM.
resource "bytebase_iam_policy" "workspace_iam" {
  depends_on = [
    bytebase_user.workspace_dba,
    bytebase_user.workspace_auditor,
    bytebase_user.project_developer,
    bytebase_user.service_account,
    bytebase_group.developers,
    bytebase_role.auditor
  ]

  parent = "workspaces/-"

  iam_policy {
    binding {
      role = "roles/workspaceAdmin"
      members = [
        format("user:%s", bytebase_user.service_account.email),
      ]
    }

    binding {
      role = "roles/workspaceDBA"
      members = [
        format("user:%s", bytebase_user.workspace_dba.email),
      ]
    }

    binding {
      role = bytebase_role.auditor.name
      members = [
        format("user:%s", bytebase_user.workspace_auditor.email)
      ]
    }

    binding {
      role = "roles/projectViewer"
      members = [
        format("user:%s", bytebase_user.project_developer.email),
        format("group:%s", bytebase_group.developers.email),
      ]
    }
  }
}

# Project level IAM
resource "bytebase_iam_policy" "project_iam" {
  depends_on = [
    bytebase_project.sample_project,
    bytebase_user.workspace_dba,
    bytebase_user.project_developer,
    bytebase_group.developers,
    bytebase_group.project_owners
  ]

  parent = bytebase_project.sample_project.name

  iam_policy {
    binding {
      role = "roles/projectOwner"
      members = [
        format("user:%s", bytebase_user.workspace_dba.email),
        format("group:%s", bytebase_group.project_owners.email),
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
      role = "roles/projectExporter"
      members = [
        format("user:%s", bytebase_user.project_developer.email)
      ]
      condition {
        database         = "instances/test-sample-instance/databases/employee"
        tables           = ["dept_emp", "dept_manager"]
        row_limit        = 10000
        expire_timestamp = "2027-03-09T16:17:49Z"
      }
    }
  }
}
