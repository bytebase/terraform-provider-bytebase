resource "bytebase_setting" "environments" {
  name = "settings/ENVIRONMENT"

  environment_setting {
    environment {
      id        = local.environment_id_prod
      title     = "Prod"
      protected = true
    }

    environment {
      id        = local.environment_id_test
      title     = "Test"
      protected = false
    }
  }
}

# Upsert test environment.
resource "bytebase_environment" "test" {
  depends_on = [
    bytebase_setting.environments
  ]
  resource_id = local.environment_id_test
  title       = "Staging" // rename to "Staging"
  order       = 0         // change order to 0
  protected   = false
}

# Upsert prod environment.
resource "bytebase_environment" "prod" {
  depends_on = [
    bytebase_environment.test
  ]
  resource_id = local.environment_id_prod
  title       = "Prod"
  order       = 1 // change order to 1
  protected   = true
}

resource "bytebase_policy" "rollout_policy" {
  depends_on = [bytebase_environment.test]
  parent     = bytebase_environment.test.name
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

resource "bytebase_policy" "disable_copy_data_policy" {
  depends_on = [bytebase_environment.test]
  parent     = bytebase_environment.test.name
  type       = "DISABLE_COPY_DATA"

  disable_copy_data_policy {
    enable = true
  }
}

resource "bytebase_policy" "data_source_query_policy" {
  depends_on = [bytebase_environment.test]
  parent     = bytebase_environment.test.name
  type       = "DATA_SOURCE_QUERY"

  data_source_query_policy {
    restriction  = "FALLBACK"
    disallow_ddl = false
    disallow_dml = false
  }
}
