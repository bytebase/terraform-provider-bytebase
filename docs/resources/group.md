---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bytebase_group Resource - terraform-provider-bytebase"
subcategory: ""
description: |-
  The group resource.
---

# bytebase_group (Resource)

The group resource.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `email` (String) The group email.
- `members` (Block Set, Min: 1) The members in the group. (see [below for nested schema](#nestedblock--members))
- `title` (String) The group title.

### Optional

- `description` (String) The group description.

### Read-Only

- `create_time` (String) The group create time in YYYY-MM-DDThh:mm:ss.000Z format
- `creator` (String) The group creator in users/{email} format.
- `id` (String) The ID of this resource.
- `name` (String) The group name in groups/{email} format.
- `source` (String) Source means where the group comes from. For now we support Entra ID SCIM sync, so the source could be Entra ID.

<a id="nestedblock--members"></a>
### Nested Schema for `members`

Required:

- `member` (String) The member in users/{email} format.
- `role` (String) The member's role in the group.

