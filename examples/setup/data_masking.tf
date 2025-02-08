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

resource "bytebase_policy" "global_masking_policy" {
  depends_on = [
    bytebase_instance.prod,
    bytebase_environment.test
  ]

  parent              = ""
  type                = "MASKING_RULE"
  enforce             = true
  inherit_from_parent = false

  global_masking_policy {
    rules {
      condition     = "environment_id in [\"test\"]"
      id            = "69df1d15-abe5-4bc9-be38-f2a4bef3f7e0"
      semantic_type = "bb.default-partial"
    }
    rules {
      condition     = "instance_id in [\"prod-sample-instance\"]"
      id            = "90adb734-0808-4c9f-b281-1f76f7a1a29a"
      semantic_type = "bb.default"
    }
  }
}
