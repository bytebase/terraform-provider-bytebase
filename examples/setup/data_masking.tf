resource "bytebase_policy" "masking_policy" {
  depends_on = [
    bytebase_instance.test
  ]

  parent              = "instances/test-sample-instance/databases/employee"
  type                = "MASKING"
  enforce             = true
  inherit_from_parent = false

  masking_policy {
    mask_data {
      table         = "salary"
      column        = "amount"
      masking_level = "FULL"
    }
    mask_data {
      table         = "salary"
      column        = "emp_no"
      masking_level = "NONE"
    }
  }
}

resource "bytebase_policy" "masking_exception_policy" {
  depends_on = [
    bytebase_project.sample_project
  ]

  parent              = bytebase_project.sample_project.name
  type                = "MASKING_EXCEPTION"
  enforce             = true
  inherit_from_parent = false

  masking_exception_policy {
    exceptions {
      database      = "instances/test-sample-instance/databases/employee"
      table         = "salary"
      column        = "amount"
      masking_level = "NONE"
      member        = "user:ed@bytebase.com"
      action        = "EXPORT"
    }
    exceptions {
      database      = "instances/test-sample-instance/databases/employee"
      table         = "salary"
      column        = "amount"
      masking_level = "NONE"
      member        = "user:ed@bytebase.com"
      action        = "QUERY"
    }
  }
}
