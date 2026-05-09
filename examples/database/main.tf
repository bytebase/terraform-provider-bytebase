# Examples for query the database
terraform {
  required_version = ">= 1.11"
  required_providers {
    bytebase = {
      version = "3.18.0"
      # For local development, please use "terraform.local/bytebase/bytebase" instead
      source = "registry.terraform.io/bytebase/bytebase"
    }
  }
}

provider "bytebase" {
  # You need to replace the account and key with your Bytebase service account.
  service_account = "terraform@service.bytebase.com"
  service_key     = "bbs_BxVIp7uQsARl8nR92ZZV"
  # The Bytebase service URL. You can use the external URL in production.
  # Check the docs about external URL: https://www.bytebase.com/docs/get-started/install/external-url
  url = "https://bytebase.example.com"
}

data "bytebase_database_list" "all" {
  # parent defaults to workspace when not specified.
  environment = "environments/test"
  project     = "projects/sample-project"
}

output "all_databases" {
  value = data.bytebase_database_list.all
}

# Example: OpenSearch / document-DB nested masking via object_schema_json.
# The JSON must match the v1.ObjectSchema proto shape.
# Replace <uuid-from-ui> with real semantic type IDs from the Bytebase
# UI at Settings -> Data Masking -> Semantic Types.
#
# resource "bytebase_database" "opensearch_users" {
#   name        = "instances/opensearch-cluster/databases/node-1"
#   project     = "projects/sample-project"
#   environment = "environments/test"
#
#   catalog {
#     schemas {
#       name = ""
#       tables {
#         name = "users_index"
#         object_schema_json = jsonencode({
#           type = "OBJECT"
#           structKind = {
#             properties = {
#               email = { type = "STRING", semanticType = "<uuid-from-ui>" }
#               contact = {
#                 type = "OBJECT"
#                 structKind = {
#                   properties = {
#                     phone = { type = "STRING", semanticType = "<uuid-from-ui>" }
#                   }
#                 }
#               }
#             }
#           }
#         })
#       }
#     }
#   }
# }
