# Example for instance data source

This is an example of using the Bytebase Terraform provider to manage the instance.

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

## Create the instance

```terraform
resource "bytebase_instance" "dev" {
  name        = "the_instance_name"
  engine      = "POSTGRES"
  host        = "127.0.0.1"
  port        = 5432
  environment = "env_name"

  # You need to specific the data source
  data_source_list {
    name     = "admin data source"
    type     = "ADMIN"
    username = "<The connection user name>"
    password = "<The connection user password>"
  }

  # And you can add another data_source_list with RO type
  data_source_list {
    name     = "read-only data source"
    type     = "RO"
    username = "<The connection user name>"
    password = "<The connection user password>"
  }
}
```
