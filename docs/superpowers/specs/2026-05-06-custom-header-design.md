# Custom Header Provider Configuration Design

## Context

Linear issue BYT-7141 asks the Terraform provider to support custom HTTP headers, primarily so users can pass a zero-trust token when their Bytebase instance sits behind an access gateway.

The existing provider exposes only `url`, `service_account`, and `service_key` at provider configuration time. `providerConfigure` creates a single Bytebase client, and that client handles both login and later Connect RPC API calls. The current auth interceptor is already the central point for attaching the Bytebase bearer token.

## User-Facing Shape

Add an optional repeated provider block named `custom_header`, following the AWS CloudFront-style shape:

```hcl
provider "bytebase" {
  url             = "https://bytebase.example.com"
  service_account = "service@example.com"
  service_key     = var.bytebase_service_key

  custom_header {
    name  = "zero_trust_token"
    value = var.zero_trust_token
  }
}
```

Each `custom_header` block has:

- `name`: required string.
- `value`: required sensitive string.

Multiple `custom_header` blocks are allowed.

## Behavior

The provider sends configured custom headers on:

- The initial login request.
- All authenticated Connect RPC requests after login.

This is required because a zero-trust gateway can protect the login endpoint as well as later API endpoints.

The provider-owned `Authorization: Bearer ...` header remains controlled by the Bytebase client after login. Custom headers are for additional gateway or proxy headers and are not intended to replace the Bytebase authorization token.

## Implementation Design

Provider schema:

- Add provider-level constants for `custom_header`, `name`, and `value`.
- Define `custom_header` as an optional list of nested resources.
- Mark `custom_header.value` as sensitive.
- Parse the configured blocks in `providerConfigure` into a `map[string]string`.

Client construction:

- Extend `client.NewClient` with option support, including `WithCustomHeaders(map[string]string)`.
- Store a defensive copy of the custom headers in client configuration.
- Apply custom headers to the login request before calling `AuthService.Login`.
- Extend the Connect interceptor so it sets custom headers on unary and streaming client requests.
- Continue setting `Authorization` after custom headers so the provider's bearer token remains authoritative.

Duplicate header names are resolved by normal Terraform list order: later blocks overwrite earlier blocks when parsed into the client header map.

## Documentation

Regenerate provider docs so `docs/index.md` includes the new `custom_header` nested block. If the generated docs do not clearly communicate the zero-trust use case, add a short provider example in the README or examples directory.

## Testing

Add focused tests before implementation:

- Provider schema test:
  - `custom_header` exists.
  - It is optional and repeatable.
  - `name` and `value` are required strings.
  - `value` is sensitive.

- Client request test:
  - Use real Connect handlers with `httptest`.
  - Assert custom headers reach the login handler.
  - Assert custom headers reach an authenticated follow-up handler.
  - Assert the follow-up request still includes the Bytebase bearer token.

Run `go test ./provider ./client`, then `go test ./...` after implementation.

## Out Of Scope

- Environment variables for arbitrary custom headers.
- A special `zero_trust_token` provider field.
- Per-resource custom headers.
- Supporting duplicate header names with multiple values.
