resource "bytebase_risk" "risk" {
  title     = "Risk for prod environment"
  source    = "DML"
  level     = 300
  active    = true
  condition = "environment_id == \"prod\" && affected_rows >= 100"
}
