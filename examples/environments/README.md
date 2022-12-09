# Example for environment data source

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
