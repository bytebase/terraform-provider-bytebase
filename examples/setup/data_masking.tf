resource "bytebase_setting" "classification" {
  name = "settings/DATA_CLASSIFICATION"

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

resource "bytebase_setting" "semantic_types" {
  name = "settings/SEMANTIC_TYPES"

  semantic_types {
    id    = "bb.default"
    title = "Default"
  }

  semantic_types {
    id    = "bb.default-partial"
    title = "Default partial"
  }

  semantic_types {
    id    = "9c84e2a6-02e5-4031-89c5-13342b568f7b"
    title = "Full mask"
    algorithm {
      full_mask {
        substitution = "***"
      }
    }
  }

  semantic_types {
    id    = "bb14d0ae-5aff-4f65-9143-bced63f8b054"
    title = "Range mask"
    algorithm {
      range_mask {
        slices {
          start        = 0
          end          = 1
          substitution = "*"
        }

        slices {
          start        = 2
          end          = 3
          substitution = "*"
        }
      }
    }
  }

  semantic_types {
    id    = "9c84e2a6-02e5-4031-89c5-13342b568f8b"
    title = "Inner Outer mask"
    algorithm {
      inner_outer_mask {
        prefix_len   = 1
        suffix_len   = 1
        substitution = "*"
        type         = "INNER"
      }
    }
  }
}

resource "bytebase_policy" "masking_exception_policy" {
  depends_on = [
    bytebase_project.sample_project,
    bytebase_instance.test,
    bytebase_user.project_developer,
    bytebase_user.workspace_dba
  ]

  parent              = bytebase_project.sample_project.name
  type                = "MASKING_EXCEPTION"
  enforce             = true
  inherit_from_parent = false

  masking_exception_policy {
    exceptions {
      database = "instances/test-sample-instance/databases/employee"
      table    = "salary"
      columns  = ["amount", "emp_no"]
      members = [
        format("user:%s", bytebase_user.project_developer.email),
        format("user:%s", bytebase_user.workspace_dba.email),
      ]
      actions = ["QUERY", "EXPORT"]
      reason  = "Grant access"
    }

    exceptions {
      database = "instances/test-sample-instance/databases/employee"
      table    = "employee"
      columns  = ["emp_no"]
      members = [
        format("user:%s", bytebase_user.workspace_dba.email),
      ]
      actions = ["EXPORT"]
      reason  = "Grant export access"
    }

    exceptions {
      members = [
        format("user:%s", bytebase_user.project_developer.email),
      ]
      actions        = ["QUERY"]
      reason         = "Grant query access"
      raw_expression = "resource.instance_id == \"test-sample-instance\" && resource.database_name == \"employee\" && resource.table_name == \"employee\" && resource.column_name in [\"first_name\", \"last_name\", \"gender\"]"
    }
  }
}

resource "bytebase_policy" "global_masking_policy" {
  depends_on = [
    bytebase_instance.prod,
    bytebase_setting.environments
  ]

  parent              = "workspaces/-"
  type                = "MASKING_RULE"
  enforce             = true
  inherit_from_parent = false

  global_masking_policy {
    rules {
      condition     = "resource.environment_id in [\"test\"]"
      id            = "69df1d15-abe5-4bc9-be38-f2a4bef3f7e0"
      semantic_type = "bb.default-partial"
      title         = "Partial masking for test environment"
    }
    rules {
      condition     = "resource.instance_id in [\"prod-sample-instance\"]"
      id            = "90adb734-0808-4c9f-b281-1f76f7a1a29a"
      semantic_type = "bb.default"
      title         = "Default masking for prod instance"
    }
  }
}
