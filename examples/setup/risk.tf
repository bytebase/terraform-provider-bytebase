resource "bytebase_risk" "risk" {
  title     = "Risk for prod environment"
  source    = "DML"
  level     = "HIGH"
  active    = true
  condition = "resource.environment_id == \"prod\" && statement.affected_rows >= 100"
}
