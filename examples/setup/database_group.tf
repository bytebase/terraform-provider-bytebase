resource "bytebase_database_group" "databases_in_test" {
  depends_on = [bytebase_project.sample_project]

  resource_id = "databases-in-test"
  project     = bytebase_project.sample_project.name
  title       = "Databases in test env"
  condition   = "resource.environment_name == \"test\""
}
