# Example for project data source

This is an example of using the Terraform Bytebase Provider to query the project.

You should replace the provider initial variables with your own and exec the [setup](../setup/) before running this example.

## List project

```terraform
data "bytebase_project_list" "all" {}
```

## Find project by id

```terraform
data "bytebase_project" "find_project" {
  resource_id = "<target project resource id>"
}
```
