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
