# Example for environment data source

This is an example of using the Bytebase Terraform provider to manage the environment.

## List environment

```terraform
data "bytebase_environment_list" "all" {}
```

## Find environment by name

```terraform
data "bytebase_environment" "find_env" {
  name = "<target environment name>"
}
```

## Create the environment

```terraform
resource "bytebase_environment" "new_env" {
  name                     = "env_name"
  order                    = 0
  environment_tier_policy  = "UNPROTECTED"
  pipeline_approval_policy = "MANUAL_APPROVAL_NEVER"
  backup_plan_policy       = "UNSET"
}
```
