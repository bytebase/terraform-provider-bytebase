resource "bytebase_review_config" "sample" {
  depends_on = [
    bytebase_setting.environments,
    bytebase_project.sample_project
  ]

  resource_id = "review-config-sample"
  title       = "Sample SQL Review Config"
  enabled     = true
  resources = toset([
    bytebase_setting.environments.environment_setting[0].environment[0].name,
    bytebase_setting.environments.environment_setting[0].environment[1].name,
    bytebase_project.sample_project.name
  ])
  rules {
    type   = "COLUMN_NO_NULL"
    engine = "MYSQL"
    level  = "WARNING"
  }
  rules {
    type                 = "COLUMN_REQUIRED"
    engine               = "MYSQL"
    level                = "ERROR"
    string_array_payload = ["id", "created_ts", "updated_ts", "creator_id", "updater_id"]
  }
  rules {
    type   = "TABLE_REQUIRE_PK"
    engine = "MYSQL"
    level  = "ERROR"
  }
  rules {
    type   = "NAMING_COLUMN"
    engine = "MYSQL"
    level  = "ERROR"
    naming_payload {
      format = "^[a-z]+(_[a-z]+)*$"
    }
  }
  rules {
    type           = "STATEMENT_MAXIMUM_LIMIT_VALUE"
    engine         = "MYSQL"
    level          = "ERROR"
    number_payload = 1000
  }
}
