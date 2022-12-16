# Example for environment data source

This is an example of using the Terraform Bytebase Provider to query the environment.

You should replace the provider initial variables with your own and exec the [setup](../setup/) before running this example.

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
