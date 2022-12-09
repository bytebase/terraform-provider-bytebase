# Example for instance data source

## List instance

```terraform
data "bytebase_instance_list" "all" {}
```

## Find instance by name

```terraform
data "bytebase_instance" "find_instance" {
  name = "<target instance name>"
}
```
