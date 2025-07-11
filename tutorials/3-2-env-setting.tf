resource "bytebase_policy" "rollout_policy_test" {
  depends_on = [bytebase_setting.environments]
  parent     = bytebase_setting.environments.environment_setting[0].environment[0].name
  type       = "ROLLOUT_POLICY"

  rollout_policy {
    automatic = true
    roles = [
      "roles/workspaceAdmin",
      "roles/projectOwner",
      "roles/LAST_APPROVER",
      "roles/CREATOR"
    ]
  }
}

resource "bytebase_policy" "rollout_policy_prod" {
  depends_on = [bytebase_setting.environments]
  parent     = bytebase_setting.environments.environment_setting[0].environment[1].name
  type       = "ROLLOUT_POLICY"

  rollout_policy {
    automatic = false
    roles = [
      "roles/workspaceAdmin",
      "roles/projectOwner",
      "roles/LAST_APPROVER",
      "roles/CREATOR"
    ]
  }
}

resource "bytebase_policy" "disable_copy_data_policy_prod" {
  depends_on = [bytebase_setting.environments]
  parent     = bytebase_setting.environments.environment_setting[0].environment[1].name
  type       = "DISABLE_COPY_DATA"

  disable_copy_data_policy {
    enable = true
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