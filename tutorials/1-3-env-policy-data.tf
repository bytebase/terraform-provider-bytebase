# Restrict SQL Editor data access on production
resource "bytebase_policy" "query_data_policy_prod" {
  depends_on = [bytebase_setting.environments]
  parent     = bytebase_setting.environments.environment_setting[0].environment[1].name
  type       = "DATA_QUERY"

  query_data_policy {
    maximum_result_rows     = 1000
    disable_copy_data       = true
    disable_export          = true
    allow_admin_data_source = false
  }
}
