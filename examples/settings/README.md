# Workspace settings example

This example configures AI and MCP and reads the workspace's approval, profile,
classification, and semantic-type settings. Run the [setup example](../setup/)
first to populate the settings being queried.

Use Bytebase 3.25.0 and provider 3.25.0. The example uses the Terraform Registry
provider once version 3.25.0 is published. For local development, follow the
[local development instructions](../README.md) and change the provider source to
`terraform.local/bytebase/bytebase`.

## AI

Provide `ai_api_key` and `ai_model` through Terraform variables, for example
`TF_VAR_ai_api_key` and `TF_VAR_ai_model`. Choose a model that supports OpenAI's
Chat Completions endpoint. Bytebase uses the full inference URL, not just a base
URL. The example configures a self-hosted Bytebase server. AI provider settings on
Bytebase SaaS are server-managed: omit the AI resource and AI input variables.
The AI data source can still read these settings; set its `name` to `"settings/AI"`
instead of referencing the omitted resource.

The API key is sensitive but is stored in Terraform state; protect the state
backend. The server redacts the key on reads, so the example outputs only
non-secret AI fields. `version` is read-only in Bytebase 3.24.0.

Removing the AI resource preserves the server's configuration. To disable AI,
set `enabled = false` and apply before removal.

AI SQL review has a separate `ai.sql_review_enabled` switch. See the dedicated
[AI review example](../ai_review/) for the licensed feature, including a SaaS
toggle-only configuration. Enabling general AI here does not enable AI SQL
review. Add the switch to the existing resource if it already manages
`settings/AI`; do not manage the same setting in two resources.

## SharePoint Export

The dedicated [SharePoint example](../sharepoint/) configures `settings/EXPORT`
and a project's export destination. It requires a self-hosted server with the
SharePoint export license feature. This settings example does not enable that
feature or change shared export credentials.

## MCP

The example sets `mcp.capability` to `READ_ONLY`. The other supported values are
`DISABLED` and `READ_WRITE`. MCP sessions remain bounded by the connected user's
permissions. Destroying the MCP resource disables MCP for the entire workspace.

## Inactivity Timeout

The [setup profile](../setup/main.tf) configures
`inactive_session_timeout_in_seconds = 1800`. Zero disables the timeout. This
requires the token duration control license feature. The settings example reads
the configured timeout without managing the profile a second time.
