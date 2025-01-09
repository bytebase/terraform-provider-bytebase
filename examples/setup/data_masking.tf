resource "bytebase_setting" "classification" {
  name = "bb.workspace.data-classification"

  classification {
    id    = "unique-id"
    title = "Classification Example"

    levels {
      id    = "1"
      title = "Level 1"
    }
    levels {
      id    = "2"
      title = "Level 2"
    }

    classifications {
      id    = "1"
      title = "Basic"
    }
    classifications {
      id    = "1-1"
      title = "User basic info"
      level = "2"
    }
    classifications {
      id    = "1-2"
      title = "User contact info"
      level = "2"
    }
    classifications {
      id    = "2"
      title = "Relationship"
    }
    classifications {
      id    = "2-1"
      title = "Social info"
      level = "2"
    }
  }
}

resource "bytebase_database_catalog" "employee_catalog" {
  depends_on = [
    bytebase_instance.test,
    bytebase_setting.classification
  ]

  database = "instances/test-sample-instance/databases/employee"

  schemas {
    tables {
      name = "salary"
      columns {
        name          = "amount"
        semantic_type = "default"
      }
      columns {
        name           = "emp_no"
        semantic_type  = "default-partial"
        classification = "1-1"
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
