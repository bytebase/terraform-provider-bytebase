resource "bytebase_review_config" "sample" {
  depends_on = [
    bytebase_setting.environments
  ]

  resource_id = "review-config-sample"
  title       = "Sample SQL Review Config"
  enabled     = true
  resources = toset([
    bytebase_setting.environments.environment_setting[0].environment[0].name,
    bytebase_setting.environments.environment_setting[0].environment[1].name
  ])
  rules {
    type   = "column.no-null"
    engine = "MYSQL"
    level  = "WARNING"
  }
  rules {
    type   = "table.require-pk"
    engine = "MYSQL"
    level  = "ERROR"
  }
}
