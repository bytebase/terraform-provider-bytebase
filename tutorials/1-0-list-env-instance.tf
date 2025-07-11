# List all environments using settings
data "bytebase_setting" "environments" {
   name = "settings/ENVIRONMENT"
}
output "all_environments" {
   value = data.bytebase_setting.environments
}

# List all instances
data "bytebase_instance_list" "all" {}
output "all_instances" {
   value = data.bytebase_instance_list.all
}