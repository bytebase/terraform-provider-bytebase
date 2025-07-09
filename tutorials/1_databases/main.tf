terraform {
  required_providers {
    bytebase = {
      version = "3.8.0"
      # For local development, please use "terraform.local/bytebase/bytebase" instead
      source = "registry.terraform.io/bytebase/bytebase"
    }
  }
}

provider "bytebase" {
  service_account = "tf@service.example.com"
  service_key     = "bbs_xxxx"
  url = "https://xxx.xxx.xxx"
}

# Local environment IDs
locals {
  environment_id_test = "test"
  environment_id_prod = "prod"
}

# Environment configuration (step 4a)
resource "bytebase_setting" "environments" {
  name = "settings/ENVIRONMENT"

  environment_setting {
    environment {
      id        = local.environment_id_test
      title     = "Test"
      protected = false
    }
    environment {
      id        = local.environment_id_prod
      title     = "Prod"
      protected = true
    }
  }
}

# MySQL test instance
resource "bytebase_instance" "test" {
  depends_on  = [bytebase_setting.environments]
  resource_id = "mysql-test"
  environment = "environments/${local.environment_id_test}"
  title       = "MySQL test"
  engine      = "MYSQL"
  activation  = false

  data_sources {
    id       = "admin data source mysql-test"
    type     = "ADMIN"
    host     = "host.docker.internal"
    port     = "3307"
    username = "root"
    password = "testpwd1"
  }
}

# MySQL production instance
resource "bytebase_instance" "prod" {
  depends_on  = [bytebase_setting.environments]
  resource_id = "mysql-prod"
  environment = "environments/${local.environment_id_prod}"
  title       = "MySQL prod"
  engine      = "MYSQL"
  activation  = false

  data_sources {
    id       = "admin data source mysql-prod"
    type     = "ADMIN"
    host     = "host.docker.internal"
    port     = "3308"
    username = "root"
    password = "testpwd1"
  }
}
