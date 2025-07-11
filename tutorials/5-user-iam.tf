# Create users and groups
resource "bytebase_user" "workspace_admin" {
  email = "admin@example.com"
  title = "Workspace Admin"
  type  = "USER"
}

resource "bytebase_user" "tf_service_account" {
  email = "tf@service.bytebase.com"
  title = "Terraform Service Account"
  type  = "SERVICE_ACCOUNT"
}

resource "bytebase_user" "workspace_dba1" {
  email = "dba@example.com"
  title = "Database Administrator 1"
  type  = "USER"
}

resource "bytebase_user" "workspace_dba2" {
  email = "dba2@example.com"
  title = "Database Administrator 2"
  type  = "USER"
}

resource "bytebase_user" "dev1" {
  email = "dev1@example.com"
  title = "Developer 1"
  type  = "USER"
}

resource "bytebase_user" "dev2" {
  email = "dev2@example.com"
  title = "Developer 2"
  type  = "USER"
}

resource "bytebase_user" "dev3" {
  email = "dev3@example.com"
  title = "Developer 3"
  type  = "USER"
}

resource "bytebase_user" "qa1" {
  email = "qa1@example.com"
  title = "QA Tester 1"
  type  = "USER"
}

resource "bytebase_user" "qa2" {
  email = "qa2@example.com"
  title = "QA Tester 2"
  type  = "USER"
}

# Create groups
resource "bytebase_group" "developers" {
  email       = "developers@example.com"
  title       = "Developer Team"
  description = "Group for all developers"
  
  members {
    member = "users/${bytebase_user.dev1.email}"
    role   = "OWNER"
  }
  
  members {
    member = "users/${bytebase_user.dev2.email}"
    role   = "MEMBER"
  }
  
  members {
    member = "users/${bytebase_user.dev3.email}"
    role   = "MEMBER"
  }
}

resource "bytebase_group" "qa" {
  email       = "qa@example.com"
  title       = "QA Team"
  description = "Group for all QA testers"
  
  members {
    member = "users/${bytebase_user.qa1.email}"
    role   = "OWNER"
  }
  
  members {
    member = "users/${bytebase_user.qa2.email}"
    role   = "MEMBER"
  }
}

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
