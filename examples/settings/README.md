# Workspace settings example

This example configures AI and MCP and reads the workspace's approval, profile,
classification, and semantic-type settings. Run the [setup example](../setup/)
first to populate the settings being queried.

Use Bytebase 3.24.0 and provider 3.24.0. The example uses the Terraform Registry
provider once version 3.24.0 is published. For local development, follow the
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

## MCP

The example sets `mcp.capability` to `READ_ONLY`. The other supported values are
`DISABLED` and `READ_WRITE`. MCP sessions remain bounded by the connected user's
permissions. Destroying the MCP resource disables MCP for the entire workspace.

## Inactivity Timeout

The [setup profile](../setup/main.tf) configures
`inactive_session_timeout_in_seconds = 1800`. Zero disables the timeout. This
requires the token duration control license feature. The settings example reads
the configured timeout without managing the profile a second time.
