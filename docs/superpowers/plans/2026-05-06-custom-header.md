# Custom Header Provider Configuration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add AWS CloudFront-style `custom_header { name, value }` provider configuration that sends custom headers on Bytebase login and authenticated API calls.

**Architecture:** The provider schema owns Terraform-facing configuration and converts repeated `custom_header` blocks into a header map. The client owns HTTP behavior through an option-based constructor and one Connect interceptor that applies custom headers plus the existing bearer token. Documentation is generated from provider schema.

**Tech Stack:** Go 1.24, Terraform Plugin SDK v2, Connect RPC, `httptest`, `tfplugindocs`.

---

## File Structure

- Modify: `provider/provider_test.go` - provider schema regression test for `custom_header`.
- Modify: `provider/provider.go` - provider constants, nested schema, config parsing, client option wiring.
- Create: `client/client_test.go` - client-level request test with real Connect handlers.
- Modify: `client/client.go` - client option type, custom header storage, login request header application.
- Modify: `client/auth.go` - interceptor support for custom headers on unary and streaming requests.
- Modify: `docs/index.md` - generated provider documentation.
- Optional modify: `README.md` - only if generated docs do not clearly show usage.

## Current Worktree Notes

The branch already contains uncommitted red-test work:

- `provider/provider_test.go` includes `TestProviderCustomHeaderSchema`.
- `client/client_test.go` includes `TestNewClientSendsCustomHeadersToLoginAndAuthenticatedRequests`.

Keep and adapt those edits. Do not revert them.

### Task 1: Confirm Red Tests

**Files:**
- Modify: `provider/provider_test.go`
- Create: `client/client_test.go`

- [ ] **Step 1: Ensure provider schema test exists**

`provider/provider_test.go` should include:

```go
func TestProviderCustomHeaderSchema(t *testing.T) {
	provider := NewProvider()
	customHeaderSchema, ok := provider.Schema[settingKeyForCustomHeader]
	if !ok {
		t.Fatal("custom_header schema is missing")
	}
	if customHeaderSchema.Type != schema.TypeList {
		t.Fatalf("custom_header schema type = %v, want %v", customHeaderSchema.Type, schema.TypeList)
	}
	if !customHeaderSchema.Optional {
		t.Fatal("custom_header should be optional")
	}

	resource, ok := customHeaderSchema.Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("custom_header Elem = %T, want *schema.Resource", customHeaderSchema.Elem)
	}
	if resource.Schema[settingKeyForCustomHeaderName].Type != schema.TypeString {
		t.Fatalf("custom_header.name schema type = %v, want %v", resource.Schema[settingKeyForCustomHeaderName].Type, schema.TypeString)
	}
	if !resource.Schema[settingKeyForCustomHeaderName].Required {
		t.Fatal("custom_header.name should be required")
	}
	if resource.Schema[settingKeyForCustomHeaderValue].Type != schema.TypeString {
		t.Fatalf("custom_header.value schema type = %v, want %v", resource.Schema[settingKeyForCustomHeaderValue].Type, schema.TypeString)
	}
	if !resource.Schema[settingKeyForCustomHeaderValue].Required {
		t.Fatal("custom_header.value should be required")
	}
	if !resource.Schema[settingKeyForCustomHeaderValue].Sensitive {
		t.Fatal("custom_header.value should be sensitive")
	}
}
```

- [ ] **Step 2: Ensure client request test exists**

`client/client_test.go` should contain:

```go
package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"buf.build/gen/go/bytebase/bytebase/connectrpc/go/v1/bytebasev1connect"
	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestNewClientSendsCustomHeadersToLoginAndAuthenticatedRequests(t *testing.T) {
	authHandler := &recordingAuthHandler{}
	actuatorHandler := &recordingActuatorHandler{}

	mux := http.NewServeMux()
	authPath, authHTTPHandler := bytebasev1connect.NewAuthServiceHandler(authHandler)
	actuatorPath, actuatorHTTPHandler := bytebasev1connect.NewActuatorServiceHandler(actuatorHandler)
	mux.Handle(authPath, authHTTPHandler)
	mux.Handle(actuatorPath, actuatorHTTPHandler)

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	headers := map[string]string{
		"zero_trust_token": "test-zero-trust-token",
		"X-Bytebase-Test":  "test-value",
	}
	if _, err := NewClient(server.URL, "service@example.com", "secret", WithCustomHeaders(headers)); err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	for name, value := range headers {
		if got := authHandler.headers.Get(name); got != value {
			t.Fatalf("login header %q = %q, want %q", name, got, value)
		}
		if got := actuatorHandler.headers.Get(name); got != value {
			t.Fatalf("actuator header %q = %q, want %q", name, got, value)
		}
	}
	if got := actuatorHandler.headers.Get("Authorization"); got != "Bearer test-token" {
		t.Fatalf("actuator Authorization header = %q, want %q", got, "Bearer test-token")
	}
}

type recordingAuthHandler struct {
	bytebasev1connect.UnimplementedAuthServiceHandler
	headers http.Header
}

func (h *recordingAuthHandler) Login(_ context.Context, req *connect.Request[v1pb.LoginRequest]) (*connect.Response[v1pb.LoginResponse], error) {
	h.headers = req.Header()
	return connect.NewResponse(&v1pb.LoginResponse{Token: "test-token"}), nil
}

type recordingActuatorHandler struct {
	bytebasev1connect.UnimplementedActuatorServiceHandler
	headers http.Header
}

func (h *recordingActuatorHandler) GetActuatorInfo(_ context.Context, req *connect.Request[v1pb.GetActuatorInfoRequest]) (*connect.Response[v1pb.ActuatorInfo], error) {
	h.headers = req.Header()
	return connect.NewResponse(&v1pb.ActuatorInfo{Workspace: "workspaces/test"}), nil
}

func (*recordingActuatorHandler) SetupSample(context.Context, *connect.Request[v1pb.SetupSampleRequest]) (*connect.Response[emptypb.Empty], error) {
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (*recordingActuatorHandler) DeleteCache(context.Context, *connect.Request[v1pb.DeleteCacheRequest]) (*connect.Response[emptypb.Empty], error) {
	return connect.NewResponse(&emptypb.Empty{}), nil
}
```

- [ ] **Step 3: Run focused tests and verify they fail for missing feature**

Run:

```bash
go test ./provider ./client
```

Expected:

```text
undefined: settingKeyForCustomHeader
undefined: WithCustomHeaders
too many arguments in call to NewClient
```

- [ ] **Step 4: Commit red tests**

```bash
git add provider/provider_test.go client/client_test.go
git commit -m "test: cover provider custom headers"
```

### Task 2: Add Provider Schema and Parsing

**Files:**
- Modify: `provider/provider.go`

- [ ] **Step 1: Add provider constants**

In `provider/provider.go`, extend the existing const block:

```go
const (
	envKeyForBytebaseURL    = "BYTEBASE_URL"
	envKeyForServiceAccount = "BYTEBASE_SERVICE_ACCOUNT"
	envKeyForServiceKey     = "BYTEBASE_SERVICE_KEY"

	settingKeyForURL               = "url"
	settingKeyForServiceAccount    = "service_account"
	settingKeyForServiceKey        = "service_key"
	settingKeyForCustomHeader      = "custom_header"
	settingKeyForCustomHeaderName  = "name"
	settingKeyForCustomHeaderValue = "value"
)
```

- [ ] **Step 2: Add the nested provider schema**

In `NewProvider().Schema`, after `service_key`, add:

```go
settingKeyForCustomHeader: {
	Type:        schema.TypeList,
	Optional:    true,
	Description: "Custom HTTP headers to include in Bytebase API requests, for example headers required by a zero-trust gateway.",
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			settingKeyForCustomHeaderName: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The custom HTTP header name.",
			},
			settingKeyForCustomHeaderValue: {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The custom HTTP header value.",
			},
		},
	},
},
```

- [ ] **Step 3: Add parser helper**

Add this helper near `providerConfigure`:

```go
func getCustomHeaders(d *schema.ResourceData) map[string]string {
	headers := map[string]string{}
	for _, item := range d.Get(settingKeyForCustomHeader).([]interface{}) {
		header := item.(map[string]interface{})
		name := header[settingKeyForCustomHeaderName].(string)
		value := header[settingKeyForCustomHeaderValue].(string)
		headers[name] = value
	}
	return headers
}
```

- [ ] **Step 4: Wire parsed headers into client creation**

Change:

```go
c, err := client.NewClient(bytebaseURL, email, key)
```

to:

```go
c, err := client.NewClient(bytebaseURL, email, key, client.WithCustomHeaders(getCustomHeaders(d)))
```

- [ ] **Step 5: Run provider tests**

Run:

```bash
go test ./provider
```

Expected: provider package passes.

- [ ] **Step 6: Commit provider schema**

```bash
git add provider/provider.go
git commit -m "feat: add custom header provider schema"
```

### Task 3: Add Client Header Options and Interceptor Behavior

**Files:**
- Modify: `client/client.go`
- Modify: `client/auth.go`

- [ ] **Step 1: Add client option types**

In `client/client.go`, before `NewClient`, add:

```go
type options struct {
	customHeaders map[string]string
}

// Option configures the Bytebase API client.
type Option func(*options)

// WithCustomHeaders configures HTTP headers to include in Bytebase API requests.
func WithCustomHeaders(headers map[string]string) Option {
	return func(o *options) {
		o.customHeaders = copyHeaders(headers)
	}
}

func copyHeaders(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return nil
	}
	copied := make(map[string]string, len(headers))
	for name, value := range headers {
		copied[name] = value
	}
	return copied
}
```

- [ ] **Step 2: Extend NewClient signature and initialize options**

Change:

```go
func NewClient(url, email, password string) (api.Client, error) {
	c := client{
		url: strings.TrimSuffix(url, "/"),
	}
```

to:

```go
func NewClient(url, email, password string, opts ...Option) (api.Client, error) {
	clientOptions := &options{}
	for _, opt := range opts {
		opt(clientOptions)
	}

	c := client{
		url: strings.TrimSuffix(url, "/"),
	}
```

- [ ] **Step 3: Pass custom headers into the auth interceptor**

Change:

```go
authInt := &authInterceptor{}
```

to:

```go
authInt := &authInterceptor{customHeaders: clientOptions.customHeaders}
```

- [ ] **Step 4: Apply custom headers to login**

After creating `loginReq`, add:

```go
setHeaders(loginReq.Header(), clientOptions.customHeaders)
```

The login block should become:

```go
loginReq := connect.NewRequest(&v1pb.LoginRequest{
	Email:    email,
	Password: password,
})
setHeaders(loginReq.Header(), clientOptions.customHeaders)

loginResp, err := c.authClient.Login(context.Background(), loginReq)
```

- [ ] **Step 5: Add header helper and interceptor field**

In `client/auth.go`, change `authInterceptor` to:

```go
type authInterceptor struct {
	token         string
	customHeaders map[string]string
}
```

Add this helper:

```go
func setHeaders(dst http.Header, headers map[string]string) {
	for name, value := range headers {
		dst.Set(name, value)
	}
}
```

Update imports in `client/auth.go`:

```go
import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
)
```

- [ ] **Step 6: Set custom headers on unary requests before Authorization**

Change `WrapUnary` to:

```go
func (a *authInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if req.Spec().IsClient {
			setHeaders(req.Header(), a.customHeaders)
			if a.token != "" {
				req.Header().Set("Authorization", fmt.Sprintf("Bearer %s", a.token))
			}
		}
		return next(ctx, req)
	})
}
```

- [ ] **Step 7: Set custom headers on streaming client requests before Authorization**

Change `WrapStreamingClient` to:

```go
func (a *authInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)
		setHeaders(conn.RequestHeader(), a.customHeaders)
		if a.token != "" {
			conn.RequestHeader().Set("Authorization", fmt.Sprintf("Bearer %s", a.token))
		}
		return conn
	})
}
```

- [ ] **Step 8: Run focused client and provider tests**

Run:

```bash
go test ./provider ./client
```

Expected: both packages pass.

- [ ] **Step 9: Commit client behavior**

```bash
git add client/client.go client/auth.go
git commit -m "feat: send provider custom headers"
```

### Task 4: Documentation and Full Verification

**Files:**
- Modify: `docs/index.md`
- Optional modify: `README.md`

- [ ] **Step 1: Regenerate Terraform provider docs**

Run:

```bash
GOOS=darwin GOARCH=amd64 go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs --provider-name=terraform-provider-bytebase
```

Expected: `docs/index.md` includes a `custom_header` block with `name` and sensitive `value`.

- [ ] **Step 2: Inspect generated provider docs**

Run:

```bash
sed -n '1,120p' docs/index.md
```

Expected content includes:

```text
custom_header
name
value
```

- [ ] **Step 3: Add README example only if needed**

If `docs/index.md` does not show an example, add this to `README.md` under Usage:

````markdown
Provider configuration can include custom HTTP headers for access gateways:

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
````

- [ ] **Step 4: Run focused tests**

Run:

```bash
go test ./provider ./client
```

Expected: pass.

- [ ] **Step 5: Run full test suite**

Run:

```bash
go test ./...
```

Expected: pass.

- [ ] **Step 6: Check final diff**

Run:

```bash
git diff --stat
git diff -- provider/provider.go client/client.go client/auth.go docs/index.md README.md
```

Expected: only custom header implementation, generated provider docs, and optional README usage example changed.

- [ ] **Step 7: Commit docs and verification updates**

```bash
git add docs/index.md README.md
git commit -m "docs: document provider custom headers"
```

If `README.md` was not changed, run:

```bash
git add docs/index.md
git commit -m "docs: document provider custom headers"
```
