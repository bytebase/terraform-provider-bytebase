# Example for policy data source

This is an example of using the Terraform Bytebase Provider to query the policy.

You should replace the provider initial variables with your own and exec the [setup](../setup/) before running this example.

## List policies

```terraform
# List all policies in a specific environment.
data "bytebase_policy_list" "env_policies" {
  environment = "<environment resource id>"
}
```

## Find policy

```terraform
data "bytebase_policy" "find_policy" {
  type = "<the policy type>"
  environment = "<environment resource id>"
}
```
