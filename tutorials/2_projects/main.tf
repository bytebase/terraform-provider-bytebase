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

# (Optional) Reference to instances and environments from previous setup
# You should already have bytebase_instance.test and bytebase_instance.prod defined

# (Optional) List existing databases for introspection:
data "bytebase_database_list" "all" {
  parent = "workspaces/-"
}

output "current_databases" {
  value = data.bytebase_database_list.all.databases
}

# Create a new project and assign specific databases
resource "bytebase_project" "another" {
  depends_on = [
    bytebase_instance.test
  ]
  resource_id = "another-project"
  title       = "Another project"

  databases = [
    "instances/mysql-test/databases/demo"
  ]
}

# (Optional) Add more databases using multiple instances
# resource "bytebase_project" "another_with_more" {
#   depends_on = [
#     bytebase_instance.test,
#     bytebase_instance.prod
#   ]
#   resource_id = "another-with-more"
#   title       = "Another project with more"
# 
#   databases = [
#     "instances/mysql-test/databases/demo",
#     "instances/mysql-prod/databases/app_prod"
#   ]
# }

# (Optional)  Use all databases from an instance dynamically
# resource "bytebase_project" "test_apps" {
#   depends_on = [
#     bytebase_instance.test
#   ]
#   resource_id = "test-applications"
#   title       = "Test Applications"
# 
#   databases = bytebase_instance.test.databases
# }
