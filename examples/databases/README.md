# Example for database data source

This is an example of using the Terraform Bytebase Provider to query the database.

You should replace the provider initial variables with your own and exec the [setup](../setup/) before running this example.

## List database

```terraform
# list all databases in all instance
data "bytebase_database_list" "all" {}

# list databases in a specific instance
data "bytebase_database_list" "all" {
  instance = "<target instance resource id>"
}

# list databases with project filter
data "bytebase_database_list" "all" {
  project = "<target project resource id>"
}
```

## Find database by name and instance

```terraform
data "bytebase_database" "find_db" {
  name = "<target database name>"
  instance = "<target instance resource id>"
}
```
