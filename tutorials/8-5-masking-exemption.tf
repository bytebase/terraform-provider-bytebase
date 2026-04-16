resource "bytebase_policy" "masking_exemption_policy" {
  depends_on = [
    bytebase_project.project-two,
    bytebase_instance.prod
  ]

  parent              = bytebase_project.project-two.name
  type                = "MASKING_EXEMPTION"
  enforce             = true
  inherit_from_parent = false

  masking_exemption_policy {
    exemptions {
      reason           = "Business requirement"
      database         = "instances/prod-sample-instance/databases/hr_prod"
      table            = "employee"
      columns          = ["birth_date", "last_name"]
      members          = ["user:admin@example.com", "user:qa1@example.com"]
      expire_timestamp = "2027-07-30T16:11:49Z"
    }
    exemptions {
      reason         = "Grant query access"
      members        = ["user:dev1@example.com"]
      raw_expression = "resource.instance_id == \"prod-sample-instance\" && resource.database_name == \"hr_prod\" && resource.table_name == \"employee\" && resource.column_name in [\"first_name\", \"last_name\", \"gender\"]"
    }
  }
}