package provider

import (
	"os"
	"testing"

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
	var _ *schema.Provider = NewProvider()
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
