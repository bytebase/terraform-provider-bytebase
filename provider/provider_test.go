package provider

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"buf.build/gen/go/bytebase/bytebase/connectrpc/go/v1/bytebasev1connect"
	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = NewProvider()
	testAccProvider.ConfigureContextFunc = internal.MockProviderConfigure
	testAccProviders = map[string]*schema.Provider{
		"bytebase": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := NewProvider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(_ *testing.T) {
	var _ = NewProvider()
}

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

	nameSchema, ok := resource.Schema[settingKeyForCustomHeaderName]
	if !ok {
		t.Fatal("custom_header.name schema is missing")
	}
	if nameSchema.Type != schema.TypeString {
		t.Fatalf("custom_header.name schema type = %v, want %v", nameSchema.Type, schema.TypeString)
	}
	if !nameSchema.Required {
		t.Fatal("custom_header.name should be required")
	}

	valueSchema, ok := resource.Schema[settingKeyForCustomHeaderValue]
	if !ok {
		t.Fatal("custom_header.value schema is missing")
	}
	if valueSchema.Type != schema.TypeString {
		t.Fatalf("custom_header.value schema type = %v, want %v", valueSchema.Type, schema.TypeString)
	}
	if !valueSchema.Required {
		t.Fatal("custom_header.value should be required")
	}
	if !valueSchema.Sensitive {
		t.Fatal("custom_header.value should be sensitive")
	}
}

func TestProviderWorkloadIdentityAuthenticationSchema(t *testing.T) {
	provider := NewProvider()
	tests := []struct {
		name      string
		sensitive bool
	}{
		{name: settingKeyForWorkloadIdentityEmail},
		{name: settingKeyForWorkloadIdentityToken, sensitive: true},
		{name: settingKeyForWorkloadIdentityTokenFile},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			field, ok := provider.Schema[test.name]
			if !ok {
				t.Fatalf("provider schema is missing %q", test.name)
			}
			if field.Type != schema.TypeString || !field.Optional {
				t.Fatalf("provider schema %q must be an optional string", test.name)
			}
			if field.Sensitive != test.sensitive {
				t.Fatalf("provider schema %q Sensitive = %v, want %v", test.name, field.Sensitive, test.sensitive)
			}
			if field.DefaultFunc == nil {
				t.Fatalf("provider schema %q must have an environment default", test.name)
			}
		})
	}
}

func TestResolveAuthenticationConfig(t *testing.T) {
	clearAuthenticationEnvironment(t)
	tests := []struct {
		name         string
		raw          map[string]interface{}
		wantService  bool
		wantWorkload bool
		wantErr      bool
	}{
		{
			name:        "service account",
			raw:         map[string]interface{}{settingKeyForServiceAccount: "service@example.com", settingKeyForServiceKey: "service-secret"},
			wantService: true,
		},
		{
			name:         "inline workload token",
			raw:          map[string]interface{}{settingKeyForWorkloadIdentityEmail: "terraform@workload.bytebase.com", settingKeyForWorkloadIdentityToken: "external-token"},
			wantWorkload: true,
		},
		{
			name:         "file workload token",
			raw:          map[string]interface{}{settingKeyForWorkloadIdentityEmail: "terraform@workload.bytebase.com", settingKeyForWorkloadIdentityTokenFile: "/secrets/token.jwt"},
			wantWorkload: true,
		},
		{name: "no authentication", raw: map[string]interface{}{}, wantErr: true},
		{name: "service account without key", raw: map[string]interface{}{settingKeyForServiceAccount: "service@example.com"}, wantErr: true},
		{name: "service key without account", raw: map[string]interface{}{settingKeyForServiceKey: "service-secret"}, wantErr: true},
		{name: "workload email without token", raw: map[string]interface{}{settingKeyForWorkloadIdentityEmail: "terraform@workload.bytebase.com"}, wantErr: true},
		{name: "inline token without email", raw: map[string]interface{}{settingKeyForWorkloadIdentityToken: "external-token"}, wantErr: true},
		{name: "token file without email", raw: map[string]interface{}{settingKeyForWorkloadIdentityTokenFile: "/secrets/token.jwt"}, wantErr: true},
		{
			name: "inline and file workload tokens",
			raw: map[string]interface{}{
				settingKeyForWorkloadIdentityEmail:     "terraform@workload.bytebase.com",
				settingKeyForWorkloadIdentityToken:     "external-token",
				settingKeyForWorkloadIdentityTokenFile: "/secrets/token.jwt",
			},
			wantErr: true,
		},
		{
			name: "mixed authentication modes",
			raw: map[string]interface{}{
				settingKeyForServiceAccount:        "service@example.com",
				settingKeyForServiceKey:            "service-secret",
				settingKeyForWorkloadIdentityEmail: "terraform@workload.bytebase.com",
				settingKeyForWorkloadIdentityToken: "external-token",
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, NewProvider().Schema, test.raw)
			got, err := resolveAuthenticationConfig(data)
			if test.wantErr {
				if err == nil {
					t.Fatal("resolveAuthenticationConfig() error = nil, want error")
				}
				if strings.Contains(err.Error(), "service-secret") || strings.Contains(err.Error(), "external-token") {
					t.Fatalf("resolveAuthenticationConfig() error exposes a credential: %q", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveAuthenticationConfig() error = %v", err)
			}
			if (got.ServiceAccount != nil) != test.wantService {
				t.Fatalf("ServiceAccount configured = %v, want %v", got.ServiceAccount != nil, test.wantService)
			}
			if (got.WorkloadIdentity != nil) != test.wantWorkload {
				t.Fatalf("WorkloadIdentity configured = %v, want %v", got.WorkloadIdentity != nil, test.wantWorkload)
			}
		})
	}
}

func TestResolveAuthenticationConfigFromEnvironment(t *testing.T) {
	t.Setenv(envKeyForServiceAccount, "")
	t.Setenv(envKeyForServiceKey, "")
	t.Setenv(envKeyForWorkloadIdentityEmail, "terraform@workload.bytebase.com")
	t.Setenv(envKeyForWorkloadIdentityToken, "external-token")
	t.Setenv(envKeyForWorkloadIdentityTokenFile, "")

	data := schema.TestResourceDataRaw(t, NewProvider().Schema, map[string]interface{}{})
	got, err := resolveAuthenticationConfig(data)
	if err != nil {
		t.Fatalf("resolveAuthenticationConfig() error = %v", err)
	}
	if got.WorkloadIdentity == nil {
		t.Fatal("resolveAuthenticationConfig() workload identity = nil")
	}
	if got.WorkloadIdentity.Email != "terraform@workload.bytebase.com" || got.WorkloadIdentity.Token != "external-token" {
		t.Fatalf("resolveAuthenticationConfig() workload identity = %+v", got.WorkloadIdentity)
	}
}

func TestProviderConfigureWithWorkloadIdentity(t *testing.T) {
	clearAuthenticationEnvironment(t)
	authHandler := &providerAuthenticationHandler{accessToken: "bytebase-token"}
	server := newProviderAuthenticationServer(t, authHandler)
	data := schema.TestResourceDataRaw(t, NewProvider().Schema, map[string]interface{}{
		settingKeyForURL:                   server.URL,
		settingKeyForWorkloadIdentityEmail: "terraform@workload.bytebase.com",
		settingKeyForWorkloadIdentityToken: "external-token",
	})

	configured, diags := providerConfigure(context.Background(), data)
	if diags.HasError() {
		t.Fatalf("providerConfigure() diagnostics = %v", diags)
	}
	if configured == nil {
		t.Fatal("providerConfigure() client = nil")
	}
	if authHandler.email != "terraform@workload.bytebase.com" {
		t.Fatalf("exchange email = %q", authHandler.email)
	}
	if authHandler.token != "external-token" {
		t.Fatalf("exchange token = %q", authHandler.token)
	}
}

func TestProviderConfigureAuthenticationErrorDoesNotExposeToken(t *testing.T) {
	clearAuthenticationEnvironment(t)
	const externalToken = "external-token-must-not-leak"
	authHandler := &providerAuthenticationHandler{exchangeErr: connect.NewError(connect.CodeUnauthenticated, errors.New(externalToken))}
	server := newProviderAuthenticationServer(t, authHandler)
	data := schema.TestResourceDataRaw(t, NewProvider().Schema, map[string]interface{}{
		settingKeyForURL:                   server.URL,
		settingKeyForWorkloadIdentityEmail: "terraform@workload.bytebase.com",
		settingKeyForWorkloadIdentityToken: externalToken,
	})

	_, diags := providerConfigure(context.Background(), data)
	if !diags.HasError() {
		t.Fatal("providerConfigure() diagnostics have no error")
	}
	for _, diagnostic := range diags {
		if strings.Contains(diagnostic.Detail, externalToken) {
			t.Fatalf("providerConfigure() diagnostic exposes token: %q", diagnostic.Detail)
		}
	}
}

func newProviderAuthenticationServer(t *testing.T, authHandler bytebasev1connect.AuthServiceHandler) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	authPath, authHTTPHandler := bytebasev1connect.NewAuthServiceHandler(authHandler)
	actuatorPath, actuatorHTTPHandler := bytebasev1connect.NewActuatorServiceHandler(&providerActuatorHandler{})
	mux.Handle(authPath, authHTTPHandler)
	mux.Handle(actuatorPath, actuatorHTTPHandler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

type providerAuthenticationHandler struct {
	bytebasev1connect.UnimplementedAuthServiceHandler
	accessToken string
	exchangeErr error
	email       string
	token       string
}

func (h *providerAuthenticationHandler) ExchangeToken(_ context.Context, req *connect.Request[v1pb.ExchangeTokenRequest]) (*connect.Response[v1pb.ExchangeTokenResponse], error) {
	h.email = req.Msg.Email
	h.token = req.Msg.Token
	if h.exchangeErr != nil {
		return nil, h.exchangeErr
	}
	return connect.NewResponse(&v1pb.ExchangeTokenResponse{AccessToken: h.accessToken}), nil
}

type providerActuatorHandler struct {
	bytebasev1connect.UnimplementedActuatorServiceHandler
}

func (*providerActuatorHandler) GetActuatorInfo(context.Context, *connect.Request[v1pb.GetActuatorInfoRequest]) (*connect.Response[v1pb.ActuatorInfo], error) {
	return connect.NewResponse(&v1pb.ActuatorInfo{
		Workspace:      "workspaces/test",
		DefaultProject: "projects/default-test",
	}), nil
}

func clearAuthenticationEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv(envKeyForServiceAccount, "")
	t.Setenv(envKeyForServiceKey, "")
	t.Setenv(envKeyForWorkloadIdentityEmail, "")
	t.Setenv(envKeyForWorkloadIdentityToken, "")
	t.Setenv(envKeyForWorkloadIdentityTokenFile, "")
}

func testAccPreCheck(t *testing.T) {
	if err := os.Getenv(envKeyForServiceAccount); err == "" {
		t.Fatal("BYTEBASE_SERVICE_ACCOUNT must be set for acceptance tests")
	}
	if err := os.Getenv(envKeyForServiceKey); err == "" {
		t.Fatal("BYTEBASE_SERVICE_KEY must be set for acceptance tests")
	}
	if err := os.Getenv(envKeyForBytebaseURL); err == "" {
		t.Fatal("BYTEBASE_URL must be set for acceptance tests")
	}
}
