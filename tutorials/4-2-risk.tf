resource "bytebase_risk" "dml_moderate" {
  title     = "DML Moderate Risk"
  source    = "DML"
  level     = "MODERATE"
  active    = true
  condition = "resource.environment_id == \"prod\" && statement.affected_rows >= 100"
}

resource "bytebase_risk" "ddl_high" {
  title     = "DDL High Risk"
  source    = "DDL"
  level     = "HIGH"
  active    = true
  condition = "resource.environment_id == \"prod\""
}
