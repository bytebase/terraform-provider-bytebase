# Example for instance data source

This is an example of using the Terraform Bytebase Provider to query the instance.

You should replace the provider initial variables with your own and exec the [setup](../setup/) before running this example.

## List instance

```terraform
data "bytebase_instance_list" "all" {}
```

## Find instance by id and environment

```terraform
data "bytebase_instance" "find_instance" {
  resource_id = "<target instance resource id>"
  environment = "<the instance environment resource id>"
}
```
