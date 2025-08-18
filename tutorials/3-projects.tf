# Project One
resource "bytebase_project" "project-one" {
  depends_on = [
    bytebase_instance.test
  ]
  resource_id = "project-one"
  title       = "Project One"

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
}

# Project Two
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