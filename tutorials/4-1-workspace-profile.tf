resource "bytebase_setting" "workspace_profile" {
  name = "settings/WORKSPACE_PROFILE"

  workspace_profile {
    disallow_signup          = true
    domains                  = ["example.com"]
    enforce_identity_domain  = false
    external_url             = "https://example.com"
  }
}