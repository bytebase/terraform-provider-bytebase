---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bytebase_policy Resource - terraform-provider-bytebase"
subcategory: ""
description: |-
  The policy resource.
---

# bytebase_policy (Resource)

The policy resource.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `parent` (String) The policy parent name for the policy, support projects/{resource id}, environments/{resource id}, instances/{resource id}, or instances/{resource id}/databases/{database name}
- `type` (String) The policy type.

### Optional

- `access_control_policy` (Block List) (see [below for nested schema](#nestedblock--access_control_policy))
- `backup_plan_policy` (Block List) (see [below for nested schema](#nestedblock--backup_plan_policy))
- `deployment_approval_policy` (Block List) (see [below for nested schema](#nestedblock--deployment_approval_policy))
- `enforce` (Boolean) Decide if the policy is enforced.
- `inherit_from_parent` (Boolean) Decide if the policy should inherit from the parent.
- `sensitive_data_policy` (Block List) (see [below for nested schema](#nestedblock--sensitive_data_policy))

### Read-Only

- `id` (String) The ID of this resource.
- `name` (String) The policy full name

<a id="nestedblock--access_control_policy"></a>
### Nested Schema for `access_control_policy`

Optional:

- `disallow_rules` (Block List) (see [below for nested schema](#nestedblock--access_control_policy--disallow_rules))

<a id="nestedblock--access_control_policy--disallow_rules"></a>
### Nested Schema for `access_control_policy.disallow_rules`

Optional:

- `all_databases` (Boolean)



<a id="nestedblock--backup_plan_policy"></a>
### Nested Schema for `backup_plan_policy`

Optional:

- `retention_duration` (Number) The minimum allowed seconds that backup data is kept for databases in an environment.
- `schedule` (String)


<a id="nestedblock--deployment_approval_policy"></a>
### Nested Schema for `deployment_approval_policy`

Optional:

- `default_strategy` (String)
- `deployment_approval_strategies` (Block List) (see [below for nested schema](#nestedblock--deployment_approval_policy--deployment_approval_strategies))

<a id="nestedblock--deployment_approval_policy--deployment_approval_strategies"></a>
### Nested Schema for `deployment_approval_policy.deployment_approval_strategies`

Optional:

- `approval_group` (String)
- `approval_strategy` (String)
- `deployment_type` (String)



<a id="nestedblock--sensitive_data_policy"></a>
### Nested Schema for `sensitive_data_policy`

Optional:

- `sensitive_data` (Block List) (see [below for nested schema](#nestedblock--sensitive_data_policy--sensitive_data))

<a id="nestedblock--sensitive_data_policy--sensitive_data"></a>
### Nested Schema for `sensitive_data_policy.sensitive_data`

Optional:

- `column` (String)
- `mask_type` (String)
- `schema` (String)
- `table` (String)


