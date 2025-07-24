
# Create a new instance named "test instance"
# You can replace the parameters with your real instance
resource "bytebase_instance" "test" {
  depends_on = [
    bytebase_setting.environments
  ]

  resource_id = local.instance_id_test
  environment = bytebase_setting.environments.environment_setting[0].environment[0].name
  title       = "test instance"
  engine      = "MYSQL"
  activation  = true

  # You need to specific the data source
  data_sources {
    id   = "admin data source"
    type = "ADMIN"
    host = "127.0.0.1"
    port = "3366"

    username = "bytebase"
    external_secret {
      vault {
        url               = "http://127.0.0.1:8200"
        token             = "<root token>"
        engine_name       = "secret"
        secret_name       = "bytebase"
        password_key_name = "database_pwd"
      }
    }
  }

  # And you can add another data_sources with RO type
  data_sources {
    id       = "read-only data source"
    type     = "READ_ONLY"
    username = "bytebase"
    password = "YOUR_DB_PWD"
    host     = "127.0.0.1"
    port     = "3366"
  }
}

# Create a new instance named "prod instance"
resource "bytebase_instance" "prod" {
  depends_on = [
    bytebase_setting.environments
  ]

  resource_id = local.instance_id_prod
  environment = bytebase_setting.environments.environment_setting[0].environment[1].name
  title       = "prod instance"
  engine      = "POSTGRES"

  # You need to specific the data source
  data_sources {
    id       = "admin data source"
    type     = "ADMIN"
    username = "bytebase"
    password = "YOUR_DB_PWD"
    host     = "127.0.0.1"
    port     = "54321"
  }
}

# Instance with external_secret
#
# Option 1, use external_secret (recommend)
# resource "bytebase_instance" "aws" {
#   resource_id    = "instance-with-aws"
#   environment    = "environments/test"
#   title          = "Instance with AWS"
#   engine         = "POSTGRES"
#   activation     = true # activation is required to be true to enable the external_secret feature

#   data_sources {
#     id       = "admin data source"
#     type     = "ADMIN"
#     host     = "127.0.0.1"
#     port     = "54321"
#     username = "bytebase"
#     external_secret {
#       aws_secrets_manager {
#         secret_name       = "bytebase-external-secret"
#         password_key_name = "db_password"
#       }
#     }
#   }
# }
#
# Option 2, work with aws provider
# terraform {
#   required_providers {
#     aws = {
#       source  = "hashicorp/aws"
#       version = "6.4.0"
#     }
#   }
# }
#
# Retrieve the secret
# data "aws_secretsmanager_secret_version" "bytebase" {
#   secret_id = "bytebase-external-secret"
# }

# resource "bytebase_instance" "aws" {
#   resource_id    = "instance-with-aws"
#   environment    = "environments/test"
#   title          = "Instance with AWS"
#   engine         = "POSTGRES"

#   data_sources {
#     id       = "admin data source"
#     type     = "ADMIN"
#     host     = "127.0.0.1"
#     port     = "54321"
#     username = "bytebase"
#     password = jsondecode(data.aws_secretsmanager_secret_version.bytebase.secret_string)["db_password"] # db_password is the secret key name in AWS Secret Manager
#   }
# }
