# Create users and groups
resource "bytebase_user" "workspace_admin" {
  email = "admin@example.com"
  title = "Workspace Admin"
  type  = "USER"
}

resource "bytebase_user" "tf_service_account" {
  email = "tf@service.bytebase.com"
  title = "Terraform Service Account"
  type  = "SERVICE_ACCOUNT"
}

resource "bytebase_user" "workspace_dba1" {
  email = "dba@example.com"
  title = "Database Administrator 1"
  type  = "USER"
}

resource "bytebase_user" "workspace_dba2" {
  email = "dba2@example.com"
  title = "Database Administrator 2"
  type  = "USER"
}

resource "bytebase_user" "dev1" {
  email = "dev1@example.com"
  title = "Developer 1"
  type  = "USER"
}

resource "bytebase_user" "dev2" {
  email = "dev2@example.com"
  title = "Developer 2"
  type  = "USER"
}

resource "bytebase_user" "dev3" {
  email = "dev3@example.com"
  title = "Developer 3"
  type  = "USER"
}

resource "bytebase_user" "qa1" {
  email = "qa1@example.com"
  title = "QA Tester 1"
  type  = "USER"
}

resource "bytebase_user" "qa2" {
  email = "qa2@example.com"
  title = "QA Tester 2"
  type  = "USER"
}