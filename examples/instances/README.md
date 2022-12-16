# Example for instance data source

This is an example of using the Bytebase Terraform provider to query the instance.

You should replace the provider initial variables with your own and exec the [setup](../setup/) before running this example.

## List instance

```terraform
data "bytebase_instance_list" "all" {}
```

## Find instance by name

```terraform
data "bytebase_instance" "find_instance" {
  name = "<target instance name>"
}
```
