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

# Create a new user.
resource "bytebase_user" "workspace_dba" {
  title = "DBA"
  email = "dba@bytebase.com"

  # Grant workspace level roles.
  roles = ["roles/workspaceDBA"]
}

# Create a new user.
resource "bytebase_user" "project_developer" {
  title = "Developer"
  email = "developer@bytebase.com"

  # Grant workspace level roles, will grant projectViewer for this user in all
  roles = ["roles/projectViewer"]
}

# Create a new project
resource "bytebase_project" "sample_project" {
  depends_on = [
    bytebase_user.workspace_dba,
    bytebase_user.project_developer
  ]

  resource_id = local.project_id
  title       = "Sample project"
  key         = "SAMM"

  members {
    member = format("user:%s", bytebase_user.workspace_dba.email)
    role   = "roles/projectOwner"
  }

  members {
    member = format("user:%s", bytebase_user.project_developer.email)
    role   = "roles/projectExporter"
    condition {
      database         = "instances/test-sample-instance/databases/employee"
      tables           = ["dept_emp", "dept_manager"]
      row_limit        = 10000
      expire_timestamp = "2027-03-09T16:17:49.637Z"
    }
  }
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

resource "bytebase_vcs_provider" "github" {
  resource_id  = "vcs-github"
  title        = "GitHub GitOps"
  type         = "GITHUB"
  access_token = "<github personal token>"
}

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
