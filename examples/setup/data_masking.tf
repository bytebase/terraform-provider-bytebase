resource "bytebase_database_catalog" "employee_catalog" {
  depends_on = [
    bytebase_instance.test
  ]

  database = "instances/test-sample-instance/databases/employee"

  schemas {
    tables {
      name = "salary"
      columns {
        name           = "amount"
        semantic_type  = "default"
        classification = "1-1-1"
      }
      columns {
        name          = "emp_no"
        semantic_type = "default-partial"
        labels = {
          tenant = "example"
          region = "asia"
        }
      }
    }
  }
}

resource "bytebase_policy" "masking_exception_policy" {
  depends_on = [
    bytebase_project.sample_project,
    bytebase_instance.test
  ]

  parent              = bytebase_project.sample_project.name
  type                = "MASKING_EXCEPTION"
  enforce             = true
  inherit_from_parent = false

  masking_exception_policy {
    exceptions {
      database = "instances/test-sample-instance/databases/employee"
      table    = "salary"
      column   = "amount"
      member   = "user:ed@bytebase.com"
      action   = "EXPORT"
    }
    exceptions {
      database = "instances/test-sample-instance/databases/employee"
      table    = "salary"
      column   = "amount"
      member   = "user:ed@bytebase.com"
      action   = "QUERY"
    }
  }
}
