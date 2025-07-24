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
      condition     = "column_name == \"birth_date\""
      id            = "birth-date-mask"
      semantic_type = "date-year-mask"
      title = "Mask Birth Date Year"
    }

    rules {
      condition     = "column_name == \"last_name\""
      id            = "last-name-first-letter-only"
      semantic_type = "name-first-letter-only"
      title = "Last Name Only Show First Letter"
    }
    
    rules {
      condition     = "classification_level in [\"2\"]"
      id            = "classification-level-2"
      semantic_type = "full-mask"
      title = "Full Mask for Classification Level 2"
    }
  }
}