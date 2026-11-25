# IAM Policy Example

`main.tf` reads workspace and project IAM policies without changing them. Run
the [setup example](../setup/) first to create `projects/project-sample`.
The IAM data source continues to return the legacy `iam_policy.binding` shape.

## Keyed Resource Bindings

Provider 3.25.1 introduces the `bindings` map for `bytebase_iam_policy` resources.
Use stable, descriptive keys so membership edits appear under the same binding
in a Terraform plan. The [setup IAM configuration](../setup/iam.tf) is a complete
resource example. For an existing configuration, follow the
[migration guide](../../migration/3.25.1.md).

For example, a project resource can declare separate owner and SQL access grants:

```hcl
resource "bytebase_iam_policy" "project_iam" {
  parent = "projects/project-sample"

  bindings = {
    project_owners = {
      role    = "roles/projectOwner"
      members = ["group:owners@example.com"]
    }
    production_sql_users = {
      role    = "roles/sqlEditorUser"
      members = ["group:developers@example.com"]
      condition = {
        database        = "instances/prod-sample-instance/databases/hr_prod"
        environment_ids = ["prod"]
      }
    }
  }
}
```

This is an illustrative snippet, not an additional resource to apply alongside
the setup. Each resource owns the entire policy for its parent: include every
existing grant you intend to keep, and do not manage the same parent in multiple
resources. Workspace policies must retain the Terraform-running identity's
administrative access. Review the full plan before applying.
