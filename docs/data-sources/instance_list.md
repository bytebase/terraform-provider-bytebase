---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bytebase_instance_list Data Source - terraform-provider-bytebase"
subcategory: ""
description: |-
  The instance data source list.
---

# bytebase_instance_list (Data Source)

The instance data source list. You can list instances through `bytebase_instance_list` data source.

## Example Usage

```terraform
# List all instances in all environments
data "bytebase_instance_list" "all" {}

output "all_instances" {
  value = data.bytebase_instance_list.all
}

# List all instances in the "dev" environment
data "bytebase_instance_list" "dev_instances" {
  environment = "dev"
}

output "dev_instances" {
  value = data.bytebase_instance_list.dev_instances
}
```

You can check [examples](https://github.com/bytebase/terraform-provider-bytebase/blob/main/examples/instances) for more usage examples.

<!-- schema generated by tfplugindocs -->

## Schema

### Optional

- `environment` (String) The environment **unique resource id**.
- `show_deleted` (Boolean) If also show the deleted instances.

### Read-Only

- `instances` (List of Object) (see [below for nested schema](#nestedatt--instances))

<a id="nestedatt--instances"></a>

### Nested Schema for `instances`

See [Instance Schema](https://registry.terraform.io/providers/bytebase/bytebase/latest/docs/data-sources/instance#schema).

### Nested Schema for `data_sources`

See [Data Source Schema](https://registry.terraform.io/providers/bytebase/bytebase/latest/docs/data-sources/instance#nested-schema-for-data_sources).
