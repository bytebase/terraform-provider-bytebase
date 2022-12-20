# Example of database role resource

This is an example of using the Terraform Bytebase Provider to query the database role in a specific instance.

You should replace the provider initial variables with your own and exec the [setup](../setup/) before running this example.

> The instance unique name is required to query the role.

## List roles

```terraform
data "bytebase_database_role_list" "all" {
  instance = "<the instance name>"
}
```

## Find role by name

```terraform
data "bytebase_database_role" "dev" {
  name     = "<the role name>"
  instance = "<the instance name>"
}
```
