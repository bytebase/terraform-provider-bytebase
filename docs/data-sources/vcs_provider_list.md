---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bytebase_vcs_provider_list Data Source - terraform-provider-bytebase"
subcategory: ""
description: |-
  The vcs provider data source list.
---

# bytebase_vcs_provider_list (Data Source)

The vcs provider data source list.



<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `id` (String) The ID of this resource.
- `vcs_providers` (List of Object) (see [below for nested schema](#nestedatt--vcs_providers))

<a id="nestedatt--vcs_providers"></a>
### Nested Schema for `vcs_providers`

Read-Only:

- `name` (String)
- `resource_id` (String)
- `title` (String)
- `type` (String)
- `url` (String)

