resource "bytebase_setting" "environments" {
  name = "settings/ENVIRONMENT"

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

# Example to upsert single environment.
# resource "bytebase_environment" "test" {
#   depends_on = [
#     bytebase_setting.environments
#   ]
#   resource_id = local.environment_id_test
#   title       = "Staging" // rename to "Staging"
#   order       = 0         // change order to 0
#   protected   = false
# }

# resource "bytebase_environment" "prod" {
#   depends_on = [
#     bytebase_environment.test
#   ]
#   resource_id = local.environment_id_prod
#   title       = "Prod"
#   order       = 1 // change order to 1
#   protected   = true
# }

resource "bytebase_policy" "rollout_policy" {
  depends_on = [bytebase_setting.environments]
  parent     = bytebase_setting.environments.environment_setting[0].environment[0].name
  type       = "ROLLOUT_POLICY"

  rollout_policy {
    automatic = true
    roles = [
      "roles/workspaceAdmin",
      "roles/projectOwner"
    ]
  }
}
