# Create or update sample project, and grant roles for users and groups.
resource "bytebase_project" "sample_project" {
  depends_on = [
    bytebase_instance.test
  ]

  resource_id            = local.project_id
  title                  = "Sample project"
  allow_modify_statement = true
  auto_resolve_issue     = false
  auto_enable_backup     = false

  databases = bytebase_instance.test.databases

  webhooks {
    title = "Sample webhook 1"
    type  = "SLACK"
    url   = "https://webhook.site/91fcd52a-39f1-4e7b-a43a-ddf72796d6b1"
    notification_types = [
      "NOTIFY_ISSUE_APPROVED",
      "NOTIFY_PIPELINE_ROLLOUT",
      "ISSUE_CREATE",
    ]
  }

  webhooks {
    title = "Sample webhook 2"
    type  = "LARK"
    url   = "https://webhook.site/91fcd52a-39f1-4e7b-a43a-ddf72796d6b1"
    notification_types = [
      "ISSUE_APPROVAL_NOTIFY",
      "ISSUE_PIPELINE_STAGE_STATUS_UPDATE"
    ]
  }
}
