# Create or update sample project, and grant roles for users and groups.
resource "bytebase_project" "sample_project" {
  depends_on = [
    bytebase_instance.test
  ]

  resource_id        = local.project_id
  title              = "Sample project"
  auto_enable_backup = false

  # New project settings
  enforce_sql_review           = true
  require_issue_approval       = true
  require_plan_check_no_error  = true
  allow_request_role           = true
  force_issue_labels           = false

  # Issue labels for categorizing issues
  issue_labels {
    value = "bug"
    color = "#FF0000"
    group = "type"
  }
  issue_labels {
    value = "feature"
    color = "#00FF00"
    group = "type"
  }
  issue_labels {
    value = "urgent"
    color = "#FFA500"
    group = "priority"
  }

  # Project labels
  labels = {
    environment = "production"
    team        = "backend"
  }

  databases = bytebase_instance.test.databases

  webhooks {
    title = "Sample webhook 1"
    type  = "SLACK"
    url   = "https://hooks.slack.com"
    notification_types = [
      "ISSUE_CREATED",
      "ISSUE_APPROVAL_REQUESTED",
      "PIPELINE_COMPLETED",
    ]
  }

  webhooks {
    title = "Sample webhook 2"
    type  = "LARK"
    url   = "https://open.larksuite.com"
    notification_types = [
      "ISSUE_SENT_BACK",
      "PIPELINE_FAILED"
    ]
  }
}
