resource "bytebase_review_config" "sample" {
  depends_on = [
    bytebase_environment.test,
    bytebase_environment.prod
  ]

  resource_id = "review-config-sample"
  title       = "Sample SQL Review Config"
  enabled     = true
  resources = toset([
    bytebase_environment.test.name,
    bytebase_environment.prod.name
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
