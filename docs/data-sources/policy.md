---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bytebase_policy Data Source - terraform-provider-bytebase"
subcategory: ""
description: |-
  The policy data source.
---

# bytebase_policy (Data Source)

The policy data source. You can get a single policy through `bytebase_policy` data source.

## Example Usage

```terraform
# Find SQL review policy in test environment.
data "bytebase_policy" "sql_review" {
  environment = "test"
  type        = "SQL_REVIEW"
}

output "sql_review_policy" {
  value = data.bytebase_policy.sql_review
}
```

You can check [examples](https://github.com/bytebase/terraform-provider-bytebase/blob/main/examples/policies) for more usage examples.

<!-- schema generated by tfplugindocs -->

## Schema

### Required

#### The policy type

- `type` (String) The policy type. Should be one of:
  - `DEPLOYMENT_APPROVAL`
  - `BACKUP_PLAN`
  - `SENSITIVE_DATA`
  - `ACCESS_CONTROL`
  - `SQL_REVIEW`

### Optional

#### Locate the policy resource

See [Locate the policy resource](https://registry.terraform.io/providers/bytebase/bytebase/latest/docs/resources/policy#optional) for details.

### Read-Only

#### The policy payload

See [The policy payload](https://registry.terraform.io/providers/bytebase/bytebase/latest/docs/resources/policy#the-policy-payload) for details.