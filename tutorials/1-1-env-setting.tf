# Environment Settings
resource "bytebase_setting" "environments" {
  name = "settings/ENVIRONMENT"

  environment_setting {
    environment {
      id        = "test"
      title     = "Test"
      protected = false
    }
    environment {
      id        = "prod"
      title     = "Prod"
      protected = true
    }
  }
}