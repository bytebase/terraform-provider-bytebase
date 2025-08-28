resource "bytebase_policy" "masking_exception_policy" {
  depends_on = [
    bytebase_project.project-two,
    bytebase_instance.prod
  ]

  parent              = bytebase_project.project-two.name
  type                = "MASKING_EXCEPTION"
  enforce             = true
  inherit_from_parent = false

  masking_exception_policy {
    exceptions {
      reason           = "Business requirement"
      database         = "instances/prod-sample-instance/databases/hr_prod"
      table            = "employee"
      columns          = ["birth_date", "last_name"]
      members          = ["user:admin@example.com"]
      actions          = ["QUERY", "EXPORT"]
      expire_timestamp = "2027-07-30T16:11:49Z"
    }
     exceptions {
      reason           = "Export data for analysis"
      members          = ["user:qa1@example.com"]
      actions          = ["EXPORT"]
      expire_timestamp = "2027-07-30T16:11:49Z"
    }
    exceptions {
      reason         = "Grant query access"
      members = ["user:dev1@example.com"]
      actions        = ["QUERY"]
      raw_expression = "resource.instance_id == \"prod-sample-instance\" && resource.database_name == \"hr_prod\" && resource.table_name == \"employee\" && resource.column_name in [\"first_name\", \"last_name\", \"gender\"]"
    }
  }
}