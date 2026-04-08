# List all identity providers.
data "bytebase_idp_list" "all" {}

output "all_idps" {
  value = data.bytebase_idp_list.all.identity_providers
}

# Example: OAuth2 identity provider (GitHub).
resource "bytebase_idp" "github" {
  resource_id = "github"
  title       = "GitHub"
  domain      = "github.com"
  type        = "OAUTH2"

  oauth2_config {
    auth_url      = "https://github.com/login/oauth/authorize"
    token_url     = "https://github.com/login/oauth/access_token"
    user_info_url = "https://api.github.com/user"
    client_id     = "your-client-id"
    client_secret = "your-client-secret"
    scopes        = ["user"]

    field_mapping {
      identifier   = "email"
      display_name = "name"
    }
  }
}

# Example: OIDC identity provider (Okta).
# resource "bytebase_idp" "okta" {
#   resource_id = "okta"
#   title       = "Okta"
#   domain      = "example.okta.com"
#   type        = "OIDC"
#
#   oidc_config {
#     issuer        = "https://example.okta.com"
#     client_id     = "your-client-id"
#     client_secret = "your-client-secret"
#     scopes        = ["openid", "profile", "email"]
#
#     field_mapping {
#       identifier   = "email"
#       display_name = "name"
#     }
#   }
# }

# Example: LDAP identity provider.
# resource "bytebase_idp" "ldap" {
#   resource_id = "company-ldap"
#   title       = "Company LDAP"
#   domain      = "example.com"
#   type        = "LDAP"
#
#   ldap_config {
#     host              = "ldap.example.com"
#     port              = 636
#     security_protocol = "LDAPS"
#     bind_dn           = "cn=admin,dc=example,dc=com"
#     bind_password     = "admin-password"
#     base_dn           = "ou=users,dc=example,dc=com"
#     user_filter       = "(uid=%s)"
#
#     field_mapping {
#       identifier   = "uid"
#       display_name = "cn"
#     }
#   }
# }
