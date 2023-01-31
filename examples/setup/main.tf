terraform {
  required_providers {
    bytebase = {
      version = "0.0.7-alpha.7"
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

locals {
  environment_id_test = "test"
  environment_id_prod = "prod"
  instance_id_test    = "test-instance"
  instance_id_prod    = "prod-instance"
  role_name           = "role_test_terraform"
}

# Create a new environment named "Test"
resource "bytebase_environment" "test" {
  resource_id             = local.environment_id_test
  title                   = "Test"
  order                   = 0
  environment_tier_policy = "UNPROTECTED"
}

# Create another environment named "Prod"
resource "bytebase_environment" "prod" {
  resource_id             = local.environment_id_prod
  title                   = "Prod"
  order                   = 1
  environment_tier_policy = "PROTECTED"
}

# Create a new instance named "test instance"
# You can replace the parameters with your real instance
resource "bytebase_instance" "test" {
  resource_id = local.instance_id_test
  environment = bytebase_environment.test.resource_id
  title       = "test instance"
  engine      = "POSTGRES"

  # You need to specific the data source
  data_sources {
    title    = "admin data source"
    type     = "ADMIN"
    username = "<The admin role name>"
    password = "<The admin password name>"
    host     = "127.0.0.1"
    port     = "5432"
  }

  # And you can add another data_sources with RO type
  data_sources {
    title    = "read-only data source"
    type     = "READ_ONLY"
    username = "<The read-only user name>"
    password = "<The read-only user password>"
    host     = "192.168.0.1"
    port     = "1234"
  }
}

# Create a new instance named "prod instance"
resource "bytebase_instance" "prod" {
  resource_id = local.instance_id_prod
  environment = bytebase_environment.prod.resource_id
  title       = "prod instance"
  engine      = "POSTGRES"

  # You need to specific the data source
  data_sources {
    title    = "admin data source"
    type     = "ADMIN"
    username = "<The connection user name>"
    password = "<The connection user password>"
    host     = "127.0.0.1"
    port     = "5432"
  }
}

# Create a new role named "role_test_terraform" in the instance "test-instance"
resource "bytebase_instance_role" "test" {
  name        = local.role_name
  instance    = bytebase_instance.test.resource_id
  environment = bytebase_instance.test.environment

  password         = "123456"
  connection_limit = 10
  valid_until      = "2022-12-31T00:00:00+08:00"

  attribute {
    super_user  = true
    no_inherit  = true
    create_role = true
    create_db   = false
    can_login   = true
    replication = true
    bypass_rls  = true
  }
}

# Create deployment approval policy for test env.
resource "bytebase_policy" "deployment_approval" {
  environment = bytebase_instance.test.environment
  type        = "DEPLOYMENT_APPROVAL"

  deployment_approval_policy {
    default_strategy = "AUTOMATIC"

    deployment_approval_strategies {
      approval_group    = "APPROVAL_GROUP_DBA"
      approval_strategy = "AUTOMATIC"
      deployment_type   = "DATABASE_CREATE"
    }
    deployment_approval_strategies {
      approval_group    = "APPROVAL_GROUP_PROJECT_OWNER"
      approval_strategy = "AUTOMATIC"
      deployment_type   = "DATABASE_DDL"
    }
  }
}

# Create backup plan policy for test env.
resource "bytebase_policy" "backup_plan" {
  environment = bytebase_instance.test.environment
  type        = "BACKUP_PLAN"

  backup_plan_policy {
    schedule           = "WEEKLY"
    retention_duration = 86400
  }
}

# Create SQL review policy for test env.
resource "bytebase_policy" "sql_review" {
  environment = bytebase_instance.test.environment
  type        = "SQL_REVIEW"

  sql_review_policy {
    title = "SQL Review Policy for Test environment"
    rules {
      type   = "statement.select.no-select-all"
      level  = "ERROR"
      engine = "MYSQL"
    }
    rules {
      type   = "statement.where.no-leading-wildcard-like"
      level  = "DISABLED"
      engine = "MYSQL"
    }
    rules {
      type   = "column.comment"
      level  = "ERROR"
      engine = "MYSQL"
      payload {
        max_length = 99
        required   = true
      }
    }
    rules {
      type   = "table.comment"
      level  = "WARNING"
      engine = "MYSQL"
      payload {
        max_length = 30
        required   = false
      }
    }
    rules {
      type   = "naming.table"
      level  = "ERROR"
      engine = "MYSQL"
      payload {
        max_length = 99
        format     = "^[a-z]+$"
      }
    }
    rules {
      type   = "column.required"
      level  = "WARNING"
      engine = "MYSQL"
      payload {
        list = ["id", "created_ts", "updated_ts"]
      }
    }
    rules {
      type   = "column.auto-increment-initial-value"
      level  = "WARNING"
      engine = "MYSQL"
      payload {
        number = 1
      }
    }
  }
}
