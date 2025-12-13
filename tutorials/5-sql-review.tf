resource "bytebase_review_config" "sample" {
  depends_on = [
    bytebase_setting.environments
  ]

  resource_id = "review-config-sample"
  title       = "Sample SQL Review Config"
  enabled     = true
  resources = toset([
    bytebase_setting.environments.environment_setting[0].environment[1].name
  ])
  rules {
    type   = "COLUMN_NO_NULL"
    engine = "POSTGRES"
    level  = "WARNING"
  }
  rules {
    type                 = "COLUMN_REQUIRED"
    engine               = "POSTGRES"
    level                = "ERROR"
    string_array_payload = ["id", "created_ts", "updated_ts", "creator_id", "updater_id"]
  }
  rules {
    type   = "TABLE_REQUIRE_PK"
    engine = "POSTGRES"
    level  = "ERROR"
  }
  rules {
    type   = "NAMING_COLUMN"
    engine = "POSTGRES"
    level  = "ERROR"
    naming_payload {
      format     = "^[a-z]+(_[a-z]+)*$"
      max_length = 64
    }
  }
  rules {
    type           = "STATEMENT_MAXIMUM_LIMIT_VALUE"
    engine         = "POSTGRES"
    level          = "ERROR"
    number_payload = 1000
  }
}
