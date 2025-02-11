resource "bytebase_database" "database" {
  depends_on = [
    bytebase_instance.test,
    bytebase_project.sample_project,
    bytebase_environment.test
  ]

  name        = "instances/test-sample-instance/databases/employee"
  project     = bytebase_project.sample_project.name
  environment = bytebase_environment.test.name

  catalog {
    schemas {
      tables {
        name = "salary"
        columns {
          name          = "amount"
          semantic_type = "bb.default"
        }
        columns {
          name           = "emp_no"
          semantic_type  = "bb.default-partial"
          classification = "1-1"
          labels = {
            tenant = "example"
            region = "asia"
          }
        }
      }
    }
  }
}
