# Project One - with all new project settings
resource "bytebase_project" "project-one" {
  depends_on = [
    bytebase_instance.test
  ]
  resource_id = "project-one"
  title       = "Project One"

  enforce_sql_review          = true
  require_issue_approval      = true
  require_plan_check_no_error = true
  allow_request_role          = true
  allow_just_in_time_access   = true
  force_issue_labels          = false

  # Issue labels
  issue_labels {
    value = "schema-change"
    color {
      red   = 0
      green = 0.4
      blue  = 0.8
    }
    group = "type"
  }
  issue_labels {
    value = "data-change"
    color {
      red   = 0.8
      green = 0.4
      blue  = 0
    }
    group = "type"
  }

  # Project labels
  labels = {
    environment = "test"
    team        = "platform"
  }

  databases = bytebase_instance.test.databases

  webhooks {
    title = "Sample webhook 1"
    type  = "SLACK"
    url   = "https://webhook.site/91fcd52a-39f1-4e7b-a43a-ddf72796d6b1"
    notification_types = [
      "ISSUE_CREATED",
      "ISSUE_APPROVAL_REQUESTED",
      "PIPELINE_COMPLETED",
    ]
  }
}

# Project Two - minimal configuration
resource "bytebase_project" "project-two" {
  depends_on = [
    bytebase_instance.prod
  ]
  resource_id = "project-two"
  title       = "Project Two"

  databases = [
    "instances/prod-sample-instance/databases/hr_prod"
  ]
}