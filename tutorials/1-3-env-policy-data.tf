resource "bytebase_policy" "disable_copy_data_prod" {
  depends_on = [bytebase_setting.environments]
  parent     = bytebase_setting.environments.environment_setting[0].environment[1].name
  type       = "DATA_QUERY"

  query_data_policy {
    disable_copy_data = true
  }
}

resource "bytebase_policy" "data_source_query_policy_prod" {
  depends_on = [bytebase_setting.environments]
  parent     = bytebase_setting.environments.environment_setting[0].environment[1].name
  type       = "DATA_SOURCE_QUERY"

  data_source_query_policy {
    restriction  = "RESTRICTION_UNSPECIFIED" # or DISALLOW or FALLBACK
    disallow_ddl = true
    disallow_dml = true
  }
}
