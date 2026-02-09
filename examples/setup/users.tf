# Create or update the user.
resource "bytebase_user" "workspace_dba" {
  title = "DBA"
  email = "dba@bytebase.com"
}

resource "bytebase_user" "project_owner" {
  title = "Project Owner"
  email = "project-owner@bytebase.com"
}

# Create or update the user.
resource "bytebase_user" "workspace_auditor" {
  title = "Auditor"
  email = "auditor@bytebase.com"
}

# Create or update the user.
resource "bytebase_user" "project_developer" {
  title = "Developer"
  email = "developer@bytebase.com"
}

# Create or update the service account.
resource "bytebase_service_account" "ci_bot" {
  parent             = "workspaces/-"
  service_account_id = "terraform"
  title              = "CI Bot"
}

# Create or update the workload identity for GitHub Actions CI/CD.
resource "bytebase_workload_identity" "github_ci" {
  parent               = "workspaces/-"
  workload_identity_id = "github-ci"
  title                = "GitHub CI"

  workload_identity_config {
    provider_type   = "GITHUB"
    subject_pattern = "repo:bytebase/sample:ref:refs/heads/main"
  }
}

# Create or update the group.
resource "bytebase_group" "developers" {
  depends_on = [
    bytebase_user.workspace_dba,
    bytebase_user.project_developer,
    # group requires the domain.
    bytebase_setting.workspace_profile
  ]

  email = "developers+dba@bytebase.com"
  title = "Bytebase Developers"

  members {
    member = format("users/%s", bytebase_user.workspace_dba.email)
    role   = "OWNER"
  }

  members {
    member = format("users/%s", bytebase_user.project_developer.email)
    role   = "MEMBER"
  }
}

resource "bytebase_group" "project_owners" {
  depends_on = [
    bytebase_user.project_owner,
    bytebase_user.workspace_dba,
    # group requires the domain.
    bytebase_setting.workspace_profile
  ]

  email = "owner+dba@bytebase.com"
  title = "Bytebase Project Owners"

  members {
    member = format("users/%s", bytebase_user.project_owner.email)
    role   = "OWNER"
  }

  members {
    member = format("users/%s", bytebase_user.workspace_dba.email)
    role   = "MEMBER"
  }
}