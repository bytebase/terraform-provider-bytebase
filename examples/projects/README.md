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

The [setup project](../setup/project.tf) sets `allow_last_plan_editor_approval = false`
to prevent the last plan editor from approving an issue, independently of
`allow_self_approval`. This example also reads that flag from the project data source.
