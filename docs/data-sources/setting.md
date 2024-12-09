---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bytebase_setting Data Source - terraform-provider-bytebase"
subcategory: ""
description: |-
  The setting data source.
---

# bytebase_setting (Data Source)

The setting data source.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String)

### Read-Only

- `approval_flow` (Block List) (see [below for nested schema](#nestedblock--approval_flow))
- `external_approval_nodes` (Block List) (see [below for nested schema](#nestedblock--external_approval_nodes))
- `id` (String) The ID of this resource.

<a id="nestedblock--approval_flow"></a>
### Nested Schema for `approval_flow`

Read-Only:

- `rules` (List of Object) (see [below for nested schema](#nestedatt--approval_flow--rules))

<a id="nestedatt--approval_flow--rules"></a>
### Nested Schema for `approval_flow.rules`

Read-Only:

- `conditions` (List of Object) (see [below for nested schema](#nestedobjatt--approval_flow--rules--conditions))
- `flow` (List of Object) (see [below for nested schema](#nestedobjatt--approval_flow--rules--flow))

<a id="nestedobjatt--approval_flow--rules--conditions"></a>
### Nested Schema for `approval_flow.rules.conditions`

Read-Only:

- `level` (String)
- `source` (String)


<a id="nestedobjatt--approval_flow--rules--flow"></a>
### Nested Schema for `approval_flow.rules.flow`

Read-Only:

- `creator` (String)
- `description` (String)
- `steps` (List of Object) (see [below for nested schema](#nestedobjatt--approval_flow--rules--flow--steps))
- `title` (String)

<a id="nestedobjatt--approval_flow--rules--flow--steps"></a>
### Nested Schema for `approval_flow.rules.flow.title`

Read-Only:

- `node` (String)
- `type` (String)





<a id="nestedblock--external_approval_nodes"></a>
### Nested Schema for `external_approval_nodes`

Read-Only:

- `nodes` (List of Object) (see [below for nested schema](#nestedatt--external_approval_nodes--nodes))

<a id="nestedatt--external_approval_nodes--nodes"></a>
### Nested Schema for `external_approval_nodes.nodes`

Read-Only:

- `endpoint` (String)
- `id` (String)
- `title` (String)

