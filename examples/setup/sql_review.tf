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
    type   = "column.no-null"
    engine = "MYSQL"
    level  = "WARNING"
  }
  rules {
    type    = "column.required"
    engine  = "MYSQL"
    level   = "ERROR"
    payload = "{\"list\":[\"id\",\"created_ts\",\"updated_ts\",\"creator_id\",\"updater_id\"]}"
  }
  rules {
    type   = "table.require-pk"
    engine = "MYSQL"
    level  = "ERROR"
  }
  rules {
    type    = "naming.column"
    engine  = "MYSQL"
    level   = "ERROR"
    payload = "{\"format\":\"^[a-z]+(_[a-z]+)*$\",\"maxLength\":64}"
  }
  rules {
    type    = "statement.maximum-limit-value"
    engine  = "MYSQL"
    level   = "ERROR"
    payload = "{\"number\":1000}"
  }
}
