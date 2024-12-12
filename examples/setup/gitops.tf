# Create GitHub GitOps provider.
resource "bytebase_vcs_provider" "github" {
  resource_id  = "vcs-github"
  title        = "GitHub GitOps"
  type         = "GITHUB"
  access_token = "<github personal token>"
}

# Connect to the GitHub repository.
resource "bytebase_vcs_connector" "github" {
  depends_on = [
    bytebase_project.sample_project,
    bytebase_vcs_provider.github
  ]

  resource_id          = "connector-github"
  title                = "GitHub Connector"
  project              = bytebase_project.sample_project.name
  vcs_provider         = bytebase_vcs_provider.github.name
  repository_id        = "ed-bytebase/gitops"
  repository_path      = "ed-bytebase/gitops"
  repository_directory = "/bytebase"
  repository_branch    = "main"
  repository_url       = "https://github.com/ed-bytebase/gitops"
}
