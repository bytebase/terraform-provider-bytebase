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
  }
}