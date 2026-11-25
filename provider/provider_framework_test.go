package provider

import (
	"context"
	"reflect"
	"strings"
	"testing"

	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestProtocol6ServerSchema(t *testing.T) {
	ctx := context.Background()
	legacy := NewProvider()
	server, err := NewProtocol6Server(ctx, legacy)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	assertNoProtocol6Errors(t, resp.Diagnostics)
	if len(resp.ResourceSchemas) != len(legacy.ResourcesMap) || len(resp.DataSourceSchemas) != len(legacy.DataSourcesMap) {
		t.Fatalf("mux lost resources or data sources: %d resources, %d data sources", len(resp.ResourceSchemas), len(resp.DataSourceSchemas))
	}
	for name := range legacy.ResourcesMap {
		if resp.ResourceSchemas[name] == nil {
			t.Errorf("missing resource %s", name)
		}
	}
	for name := range legacy.DataSourcesMap {
		if resp.DataSourceSchemas[name] == nil {
			t.Errorf("missing data source %s", name)
		}
	}
	iam := resp.ResourceSchemas["bytebase_iam_policy"]
	var bindings *tfprotov6.SchemaAttribute
	for _, attribute := range iam.Block.Attributes {
		if attribute.Name == "bindings" {
			bindings = attribute
		}
	}
	if bindings == nil || bindings.NestedType == nil || bindings.NestedType.Nesting != tfprotov6.SchemaObjectNestingModeMap {
		t.Fatal("IAM bindings must be a nested map attribute")
	}
	if iam.Version != 0 {
		t.Fatalf("IAM schema version = %d, want unchanged version zero", iam.Version)
	}
	if legacy.ResourcesMap["bytebase_iam_policy"] == nil {
		t.Fatal("constructing the mux mutated the legacy provider")
	}
}

func TestFrameworkProviderConfigurationSchemaParity(t *testing.T) {
	ctx := context.Background()
	legacy := NewProvider()
	legacyServer, err := tf5to6server.UpgradeServer(ctx, legacy.GRPCProvider)
	if err != nil {
		t.Fatal(err)
	}
	legacyResp, err := legacyServer.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	frameworkServer := providerserver.NewProtocol6(&iamFrameworkProvider{legacy: legacy})()
	frameworkResp, err := frameworkServer.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	assertNoProtocol6Errors(t, frameworkResp.Diagnostics)
	if !reflect.DeepEqual(frameworkResp.Provider, legacyResp.Provider) {
		t.Fatalf("Framework provider configuration schema differs from SDK:\nFramework: %#v\nSDK: %#v", frameworkResp.Provider.Block, legacyResp.Provider.Block)
	}
}

func TestProtocol6ServerConfigureUsesSDKDefaultsAndHeaders(t *testing.T) {
	clearAuthenticationEnvironment(t)
	t.Setenv(envKeyForBytebaseURL, "https://bytebase.example.com")
	t.Setenv(envKeyForServiceAccount, "terraform@example.com")
	t.Setenv(envKeyForServiceKey, "service-key")
	legacy := NewProvider()
	calls := 0
	legacy.ConfigureContextFunc = func(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		calls++
		if got := d.Get(settingKeyForURL); got != "https://bytebase.example.com" {
			t.Errorf("URL = %v, want environment default", got)
		}
		if got := d.Get(settingKeyForServiceAccount); got != "terraform@example.com" {
			t.Errorf("service account = %v, want environment default", got)
		}
		if got := d.Get(settingKeyForServiceKey); got != "service-key" {
			t.Errorf("service key = %v, want environment default", got)
		}
		if got := getCustomHeaders(d); !reflect.DeepEqual(got, map[string]string{"X-Gateway-Token": "secret"}) {
			t.Errorf("headers = %v", got)
		}
		return &struct{}{}, nil
	}
	ctx := context.Background()
	server, err := NewProtocol6Server(ctx, legacy)
	if err != nil {
		t.Fatal(err)
	}
	schemaResp, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	assertNoProtocol6Errors(t, schemaResp.Diagnostics)
	headerType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String, "value": tftypes.String}}
	config := protocol6ProviderConfiguration(t, schemaResp.Provider, map[string]tftypes.Value{
		"custom_header": tftypes.NewValue(tftypes.List{ElementType: headerType}, []tftypes.Value{
			tftypes.NewValue(headerType, map[string]tftypes.Value{"name": tftypes.NewValue(tftypes.String, "x-gateway-token"), "value": tftypes.NewValue(tftypes.String, "secret")}),
		}),
	})
	resp, err := server.ConfigureProvider(ctx, &tfprotov6.ConfigureProviderRequest{Config: config, TerraformVersion: "1.0.0"})
	if err != nil {
		t.Fatal(err)
	}
	assertNoProtocol6Errors(t, resp.Diagnostics)
	if calls != 1 {
		t.Fatalf("SDK configuration ran %d times, want exactly once for both provider halves", calls)
	}
}

func TestFrameworkProviderSharesConfiguredSDKClient(t *testing.T) {
	legacy := NewProvider()
	client := &struct{ marker string }{marker: "shared-client"}
	legacy.ConfigureContextFunc = func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return client, nil
	}
	if diags := legacy.Configure(context.Background(), terraform.NewResourceConfigRaw(map[string]interface{}{})); diags.HasError() {
		t.Fatal(diags)
	}
	var resp frameworkprovider.ConfigureResponse
	(&iamFrameworkProvider{legacy: legacy}).Configure(context.Background(), frameworkprovider.ConfigureRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatal(resp.Diagnostics)
	}
	if resp.ResourceData != client {
		t.Fatal("Framework resources did not receive the SDK client's identity")
	}
}

func TestProtocol6ServerPreservesAuthenticationErrors(t *testing.T) {
	clearAuthenticationEnvironment(t)
	t.Setenv(envKeyForBytebaseURL, "https://bytebase.example.com")
	ctx := context.Background()
	server, err := NewProtocol6Server(ctx, NewProvider())
	if err != nil {
		t.Fatal(err)
	}
	schemaResp, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := server.ConfigureProvider(ctx, &tfprotov6.ConfigureProviderRequest{Config: protocol6ProviderConfiguration(t, schemaResp.Provider, nil)})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, diagnostic := range resp.Diagnostics {
		if diagnostic.Severity == tfprotov6.DiagnosticSeverityError && strings.Contains(diagnostic.Detail, "one authentication mode must be configured") {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing original authentication diagnostic: %v", resp.Diagnostics)
	}
}

func protocol6ProviderConfiguration(t *testing.T, providerSchema *tfprotov6.Schema, configured map[string]tftypes.Value) *tfprotov6.DynamicValue {
	t.Helper()
	configType := providerSchema.ValueType().(tftypes.Object)
	values := make(map[string]tftypes.Value, len(configType.AttributeTypes))
	for name, valueType := range configType.AttributeTypes {
		values[name] = tftypes.NewValue(valueType, nil)
	}
	for name, value := range configured {
		values[name] = value
	}
	value, err := tfprotov6.NewDynamicValue(configType, tftypes.NewValue(configType, values))
	if err != nil {
		t.Fatal(err)
	}
	return &value
}

func assertNoProtocol6Errors(t *testing.T, diagnostics []*tfprotov6.Diagnostic) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if diagnostic != nil && diagnostic.Severity == tfprotov6.DiagnosticSeverityError {
			t.Fatalf("%s: %s", diagnostic.Summary, diagnostic.Detail)
		}
	}
}
