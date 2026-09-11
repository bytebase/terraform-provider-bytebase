# Workload identity authentication

An administrator or service account must first create the Bytebase workload
identity, configure its OIDC issuer, audience, and subject restrictions, and
grant its IAM roles. The workload identity cannot grant its own initial access.

After bootstrap, configure Terraform with only the workload identity email and
the external OIDC token file:

```bash
unset BYTEBASE_SERVICE_ACCOUNT BYTEBASE_SERVICE_KEY
export TF_VAR_bytebase_url="https://bytebase.example.com"
export TF_VAR_workload_identity_email="atlantis@workload.bytebase.com"
export TF_VAR_workload_identity_token_file="/secrets/nomad_bytebase.jwt"
terraform plan
```

The provider reads the external token during initial authentication. If a
Bytebase API request reports that its Bytebase access token is no longer
authenticated, the provider rereads the external token file, exchanges it,
and retries the request once. Keep the token file out of Terraform state,
configuration, and logs.
