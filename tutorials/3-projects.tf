# Project One
resource "bytebase_project" "project-one" {
  depends_on = [
    bytebase_instance.test
  ]
  resource_id = "project-one"
  title       = "Project One"

  databases = bytebase_instance.test.databases
}

# Project Two
resource "bytebase_project" "project-two" {
  depends_on = [
    bytebase_instance.prod
  ]
  resource_id = "project-two"
  title       = "Project Two"

  databases = [
    "instances/prod-sample-instance/databases/hr_prod"
  ]
}