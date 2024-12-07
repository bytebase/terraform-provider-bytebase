terraform {
  required_providers {
    bytebase = {
      version = "1.0.4"
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

# Correspond to the sample data Bytebase generates during onboarding.
locals {
  environment_id_test = "test"
  environment_id_prod = "prod"
  instance_id_test    = "test-sample-instance"
  instance_id_prod    = "prod-sample-instance"
  project_id          = "project-sample"
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
  depends_on = [
    bytebase_environment.test
  ]
  resource_id = local.instance_id_test
  environment = bytebase_environment.test.name
  title       = "test instance"
  engine      = "MYSQL"

  # You need to specific the data source
  data_sources {
    id       = "admin data source"
    type     = "ADMIN"
    username = "<The connection user name>"
    password = "<The connection user password>"
    host     = "127.0.0.1"
    port     = "3366"
  }

  # And you can add another data_sources with RO type
  data_sources {
    id       = "read-only data source"
    type     = "READ_ONLY"
    username = "<The connection user name>"
    password = "<The connection user password>"
    host     = "127.0.0.1"
    port     = "3366"
  }
}

# Create a new instance named "prod instance"
resource "bytebase_instance" "prod" {
  depends_on = [
    bytebase_environment.prod
  ]

  resource_id = local.instance_id_prod
  environment = bytebase_environment.prod.name
  title       = "prod instance"
  engine      = "POSTGRES"

  # You need to specific the data source
  data_sources {
    id       = "admin data source"
    type     = "ADMIN"
    username = "<The connection user name>"
    password = "<The connection user password>"
    host     = "127.0.0.1"
    port     = "54321"
  }
}

# Create a new project
resource "bytebase_project" "sample_project" {
  resource_id = local.project_id
  title       = "Sample project"
  key         = "SAMM"
}

resource "bytebase_setting" "external_approval" {
  name = "bb.workspace.approval.external"

  external_approval_nodes {
    nodes {
      id       = "9e150339-f014-4835-83d7-123aeb1895ba"
      title    = "Example node"
      endpoint = "https://example.com"
    }

    nodes {
      id       = "49a976be-50de-4541-b2d3-f2e32f8e41ef"
      title    = "Example node 2"
      endpoint = "https://example.com"
    }
  }
}

resource "bytebase_setting" "approval_flow" {
  name = "bb.workspace.approval"
  approval_flow {
    rules {
      flow {
        title       = "DBA -> OWNER"
        description = "Need DBA and workspace owner approval"
        creator     = "users/support@bytebase.com"

        # Approval flow following the step order.
        steps {
          type = "GROUP"
          node = "WORKSPACE_DBA"
        }

        steps {
          type = "GROUP"
          node = "WORKSPACE_OWNER"
        }
      }

      # Match any condition will trigger this approval flow.
      conditions {
        source = "DML"
        level  = "MODERATE"
      }
      conditions {
        source = "DDL"
        level  = "HIGH"
      }
    }
  }
}

resource "bytebase_policy" "masking_policy" {
  depends_on = [
    bytebase_instance.test
  ]

  parent              = "instances/test-sample-instance/databases/employee"
  type                = "MASKING"
  enforce             = true
  inherit_from_parent = false

  masking_policy {
    mask_data {
      table         = "salary"
      column        = "amount"
      masking_level = "FULL"
    }
    mask_data {
      table         = "salary"
      column        = "emp_no"
      masking_level = "NONE"
    }
  }
}

resource "bytebase_policy" "masking_exception_policy" {
  depends_on = [
    bytebase_project.sample_project
  ]

  parent              = bytebase_project.sample_project.name
  type                = "MASKING_EXCEPTION"
  enforce             = true
  inherit_from_parent = false

  masking_exception_policy {
    exceptions {
      database      = "instances/test-sample-instance/databases/employee"
      table         = "salary"
      column        = "amount"
      masking_level = "NONE"
      member        = "user:ed@bytebase.com"
      action        = "EXPORT"
    }
    exceptions {
      database      = "instances/test-sample-instance/databases/employee"
      table         = "salary"
      column        = "amount"
      masking_level = "NONE"
      member        = "user:ed@bytebase.com"
      action        = "QUERY"
    }
  }
}
