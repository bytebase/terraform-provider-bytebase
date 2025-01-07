---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bytebase_database_catalog Data Source - terraform-provider-bytebase"
subcategory: ""
description: |-
  The database catalog data source.
---

# bytebase_database_catalog (Data Source)

The database catalog data source.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `database` (String) The database full name in instances/{instance}/databases/{database} format

### Read-Only

- `id` (String) The ID of this resource.
- `schemas` (List of Object) (see [below for nested schema](#nestedatt--schemas))

<a id="nestedatt--schemas"></a>
### Nested Schema for `schemas`

Read-Only:

- `name` (String)
- `tables` (List of Object) (see [below for nested schema](#nestedobjatt--schemas--tables))

<a id="nestedobjatt--schemas--tables"></a>
### Nested Schema for `schemas.tables`

Read-Only:

- `classification` (String)
- `columns` (List of Object) (see [below for nested schema](#nestedobjatt--schemas--tables--columns))
- `name` (String)

<a id="nestedobjatt--schemas--tables--columns"></a>
### Nested Schema for `schemas.tables.columns`

Read-Only:

- `classification` (String)
- `labels` (Map of String)
- `name` (String)
- `semantic_type` (String)

