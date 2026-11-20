# SharePoint Export Example

Configure the workspace's shared Microsoft Entra application and create a
project with a SharePoint export destination. Use Bytebase and provider 3.25.0,
Terraform 1.11 or later, and a self-hosted server with the SharePoint export
license feature. SharePoint export is not available in Bytebase Cloud.

The application must have access to the destination library and permission to
write exports. Saving a project destination only verifies that the application
can resolve and read it; it does not prove write access.

## Run

The empty provider block reads authentication from the environment. Set
`BYTEBASE_URL`, `BYTEBASE_SERVICE_ACCOUNT`, and `BYTEBASE_SERVICE_KEY`. Workload
identity authentication is also supported; configure its environment variables
instead of service-account credentials as described in the [root README](../../README.md).

Provide these Terraform variables through your secret manager or environment:

- `TF_VAR_sharepoint_tenant_id`: Microsoft Entra tenant ID.
- `TF_VAR_sharepoint_client_id`: Microsoft Entra application ID.
- `TF_VAR_sharepoint_client_secret`: application secret.
- `TF_VAR_sharepoint_destination_url`: a document library or folder URL, for
  example `https://contoso.sharepoint.com/sites/Finance/Shared%20Documents/Exports`.

Then run `terraform init`, `terraform plan`, and `terraform apply`. The Registry
provider must be published first; for a local build, follow the
[development instructions](../README.md) and change the provider source.

This example creates `projects/sharepoint-example`. To configure an existing
Terraform-managed project, add `export_setting` and the credential dependency to
its existing resource instead of creating a second resource for the project.
The explicit `depends_on` ensures credentials are configured before the server
resolves the destination URL. Bytebase returns a canonical destination URL.

## Secret Rotation and Adoption

`client_secret` is sensitive and write-only. The example variable is also
ephemeral, so the secret is not stored in Terraform state or saved plans.
Supply it again when applying a saved plan. Outputs contain only application
IDs, secret-presence status, and the destination URL.

`client_secret_version` is a local rotation marker, not a Microsoft or Bytebase
version. It defaults to `1`. Change the secret and increment
`TF_VAR_sharepoint_client_secret_version` to rotate it. Changing the secret alone
does not schedule an update, and unrelated updates do not resend an unchanged
secret. A configured secret requires a positive version; a positive version
requires the secret to be supplied.

To adopt existing credentials without replacing them, change the resource to:

```hcl
resource "bytebase_setting" "export" {
  name = "settings/EXPORT"

  export {
    sharepoint {}
  }
}
```

Remove the now-unused required application input variables when adapting the
full example to this adoption-only configuration.

Omitted IDs and an omitted secret/version preserve their server values. Import
and refresh cannot recover the secret; import starts the local rotation marker
at `0`. The data source exposes IDs and `client_secret_set`, never the secret or
the local marker.

## Clearing and Removal

Set `export_setting {}` on the project to clear its destination. Omitting that
whole block preserves the destination instead.

Set `export {}` on `bytebase_setting.export` and apply to clear the workspace
application and its secret. This affects every project's SharePoint exports.
Empty ID strings are rejected, and removing the required `export` block is not
a clearing operation.

Destroying the setting resource preserves shared credentials and emits a
warning. `terraform destroy` still deletes the project created by this example;
review that plan before applying it.
