# Create users
resource "bytebase_user" "workspace_admin" {
  email = "admin@example.com"
  title = "Workspace Admin"
}

resource "bytebase_user" "workspace_dba1" {
  email = "dba@example.com"
  title = "Database Administrator 1"
}

resource "bytebase_user" "workspace_dba2" {
  email = "dba2@example.com"
  title = "Database Administrator 2"
}

resource "bytebase_user" "dev1" {
  email = "dev1@example.com"
  title = "Developer 1"
}

resource "bytebase_user" "dev2" {
  email = "dev2@example.com"
  title = "Developer 2"
}

resource "bytebase_user" "dev3" {
  email = "dev3@example.com"
  title = "Developer 3"
}

resource "bytebase_user" "qa1" {
  email = "qa1@example.com"
  title = "QA Tester 1"
}

resource "bytebase_user" "qa2" {
  email = "qa2@example.com"
  title = "QA Tester 2"
}

# Create service account for Terraform automation
resource "bytebase_service_account" "tf_service_account" {
  parent             = "workspaces/-"
  service_account_id = "tf"
  title              = "Terraform Service Account"
}

# Create workload identity for GitHub Actions CI/CD
resource "bytebase_workload_identity" "github_ci" {
  parent               = "workspaces/-"
  workload_identity_id = "github-ci"
  title                = "GitHub CI"

  workload_identity_config {
    provider_type   = "GITHUB"
    subject_pattern = "repo:example/repo:ref:refs/heads/main"
  }
}
