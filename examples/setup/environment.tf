resource "bytebase_setting" "environments" {
  name = "ENVIRONMENT"
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
