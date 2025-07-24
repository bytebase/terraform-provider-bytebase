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
      reason = "Business requirement"
      database = "instances/prod-sample-instance/databases/hr_prod"
      table    = "employee"
      column   = "birth_date"
      member   = "user:admin@example.com"
      action   = "QUERY"
      expire_timestamp = "2027-07-30T16:11:49Z"
    }
    exceptions {
      reason = "Export data for analysis"
      database = "instances/prod-sample-instance/databases/hr_prod"
      table    = "employee"
      column   = "last_name"
      member   = "user:admin@example.com"
      action   = "EXPORT"
      expire_timestamp = "2027-07-30T16:11:49Z"
    }
  }
}