# List all environments using settings
data "bytebase_setting" "environments" {
   name = "settings/ENVIRONMENT"
}
output "all_environments" {
   value = data.bytebase_setting.environments
}