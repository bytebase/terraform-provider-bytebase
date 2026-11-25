# Policy example

This example reads masking policies, manages project standard-review policies,
and lists the project's supported policies. Run the [setup example](../setup/)
first to create `projects/project-sample`.

Use Bytebase 3.25.0 and provider 3.25.1. The example uses the Terraform Registry
provider once version 3.25.1 is published. For local development, follow the
[local development instructions](../README.md) and change the provider source to
`terraform.local/bytebase/bytebase`.

## Standard Review

`review_rule_policy.rules` is the complete enabled set. A project policy replaces
the workspace selection; it does not add to it. Include `SYNTAX` to run standard
review rules. An empty list, or a list without `SYNTAX`, disables all of them.
Removing the project policy restores the workspace selection, or all rules when
there is no workspace policy.

Standard-review policies support workspace and project parents only. Omit
`parent` to manage a workspace policy instead.

## AI Review

`REVIEW_AI` is a separate, optional natural-language policy, not a standard-rule
selection. See the [AI review example](../ai_review/) for the licensed AI SQL
review toggle and optional workspace or project instructions. This example does
not create an AI review policy or require that feature.

The policy-list data source now includes `REVIEW_AI` when configured on the
server. Consumers of policy payloads should filter by `type` rather than assume
every result is a masking or standard-review policy.
