# Policy example

This example reads masking policies, manages project standard-review policies,
and lists the project's supported policies. Run the [setup example](../setup/)
first to create `projects/project-sample`.

Use Bytebase 3.24.0 and provider 3.24.0. The example uses the Terraform Registry
provider once version 3.24.0 is published. For local development, follow the
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
