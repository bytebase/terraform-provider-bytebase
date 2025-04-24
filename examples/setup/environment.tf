resource "bytebase_setting" "environments" {
  name = "bb.workspace.environment"
  environment_setting {
    environment {
      id        = local.environment_id_test
      title     = "Test"
      protected = false
    }

    environment {
      id        = local.environment_id_prod
      title     = "Prod"
      protected = true
    }
  }
}
