# Read the current workspace.
data "bytebase_workspace" "current" {}

output "workspace_name" {
  value = data.bytebase_workspace.current.name
}

output "workspace_plan" {
  value = data.bytebase_workspace.current.subscription[0].plan
}

# Manage workspace settings including title, logo, and license.
resource "bytebase_workspace" "main" {
  title = "My Workspace"
  # The branding logo as a data URI (e.g. data:image/png;base64,...).
  # logo = "data:image/png;base64,..."

  # Upload a license key to activate a subscription plan.
  # license = "your-license-key"
}
