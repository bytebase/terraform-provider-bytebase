# Create or update sample project, and grant roles for users and groups.
resource "bytebase_project" "sample_project" {
  depends_on = [
    bytebase_instance.test
  ]

  resource_id = local.project_id
  title       = "Sample project"

  databases = bytebase_instance.test.databases
}
