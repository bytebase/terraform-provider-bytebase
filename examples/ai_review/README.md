# AI SQL Review Example

Enable the workspace's AI SQL review switch and optionally manage custom
natural-language review instructions. Use Bytebase and provider 3.25.0,
Terraform 1.11 or later, and the AI SQL review license feature. Applying this
configuration does not start a review.

AI SQL review has its own consent switch because reviews send change statements
to the AI provider and may send target schemas and table statistics. General AI
must also be enabled and usable. On a self-hosted server, configure the AI
provider, endpoint, model, and credentials separately first; the
[settings example](../settings/) demonstrates that setup. Add the review switch
to the same resource when Terraform already manages `settings/AI`; do not create
two resources for that setting.

In Bytebase SaaS, keep the toggle-only `ai` block in this example. Provider
credentials and general AI configuration are server-managed there.

## Run

The empty provider block reads authentication from the environment. Set
`BYTEBASE_URL`, `BYTEBASE_SERVICE_ACCOUNT`, and `BYTEBASE_SERVICE_KEY`, or use
workload identity authentication as described in the [root README](../../README.md).

Run `terraform init`, `terraform plan`, and `terraform apply`. The Registry
provider must be published first; for a local build, follow the
[development instructions](../README.md) and change the provider source.

By default, only `ai.sql_review_enabled = true` is managed. Other AI fields are
omitted and preserved. Set `TF_VAR_sql_review_enabled=false` and apply to disable
review. Removing the setting resource preserves the server configuration; it
does not disable review. Enabling requires the license feature; disabling the
review switch does not.

## Optional Instructions

Leave `review_policy_content` unset to create no custom policy. A policy is not
required to enable AI SQL review. To add one, supply
`TF_VAR_review_policy_content`, such as `Require reversible schema changes and
flag possible data loss.` Content must not be blank and is limited to 65,536
bytes, not characters.

The default parent is the workspace. Set `TF_VAR_review_policy_parent` to an
existing `projects/{id}` to manage that project's policy. Workspace and project
instructions both apply; project instructions take precedence where they
conflict. Unlike `REVIEW_RULE`, the project policy does not replace the entire
workspace policy. Removing a project AI policy leaves the workspace
instructions in effect.

Policies default to `enforce = true`. Creating a policy with `enforce = false`
requires two writes: the server initially creates it enforced, then the provider
disables it. This is not atomic. If the second write fails, inspect the actual
server policy and Terraform state before retrying; the policy may remain
enforced. Updating an existing policy's `enforce` flag is supported.

The policy-list data source includes `REVIEW_AI`. Filter results by `type` before
reading type-specific payloads. Example outputs contain only the AI switches
and optional policy name.
