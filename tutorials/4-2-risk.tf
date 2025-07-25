resource "bytebase_risk" "dml_moderate" {
  title     = "DML Moderate Risk"
  source    = "DML"
  level     = 200
  active    = true
  condition = "environment_id == \"prod\" && affected_rows >= 100"
}

resource "bytebase_risk" "ddl_high" {
  title     = "DDL High Risk"
  source    = "DDL"
  level     = 300
  active    = true
  condition = "environment_id == \"prod\""
}