---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bytebase_group_list Data Source - terraform-provider-bytebase"
subcategory: ""
description: |-
  The group data source list.
---

# bytebase_group_list (Data Source)

The group data source list.



<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `groups` (List of Object) (see [below for nested schema](#nestedatt--groups))
- `id` (String) The ID of this resource.

<a id="nestedatt--groups"></a>
### Nested Schema for `groups`

Read-Only:

- `description` (String)
- `members` (Set of Object) (see [below for nested schema](#nestedobjatt--groups--members))
- `name` (String)
- `source` (String)
- `title` (String)

<a id="nestedobjatt--groups--members"></a>
### Nested Schema for `groups.members`

Read-Only:

- `member` (String)
- `role` (String)


