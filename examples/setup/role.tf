resource "bytebase_role" "auditor" {
  resource_id = "auditor-role"
  title       = "Auditor role"
  description = "This role can only list audit logs"
  permissions = [
    "bb.auditLogs.search"
  ]
}
