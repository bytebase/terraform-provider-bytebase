# Create or update the user.
resource "bytebase_user" "workspace_dba" {
  title = "DBA"
  email = "dba@bytebase.com"

  # Grant workspace level roles.
  roles = ["roles/workspaceDBA"]
}

resource "bytebase_user" "service_account" {
  title = "CI Bot"
  email = "ci-bot@service.bytebase.com"
  type  = "SERVICE_ACCOUNT"
  roles = ["roles/workspaceDBA"]
}

# Create or update the user.
resource "bytebase_user" "workspace_auditor" {
  depends_on = [
    bytebase_user.workspace_dba,
    bytebase_role.auditor
  ]
  title = "Auditor"
  email = "auditor@bytebase.com"

  # Grant workspace level roles.
  roles = [bytebase_role.auditor.name]
}

# Create or update the user.
resource "bytebase_user" "project_developer" {
  depends_on = [
    bytebase_user.workspace_auditor
  ]

  title = "Developer"
  email = "developer@bytebase.com"

  # Grant workspace level roles, will grant projectViewer for this user in all
  roles = ["roles/projectViewer"]
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
