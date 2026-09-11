package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestWorkloadIdentityConfigSupportsGenericOIDCWithJWKS(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceWorkloadIdentity().Schema, map[string]interface{}{
		"workload_identity_config": []interface{}{map[string]interface{}{
			"provider_type":     "OIDC",
			"issuer_url":        "https://issuer.example.com",
			"allowed_audiences": []interface{}{"bytebase"},
			"subject_pattern":   "sub:ci:deploy",
			"jwks_url":          "https://issuer.example.com/.well-known/jwks.json",
		}},
	})

	config := expandWorkloadIdentityConfig(d)
	if got, want := config.ProviderType.String(), "OIDC"; got != want {
		t.Fatalf("ProviderType = %q, want %q", got, want)
	}

	flattened := flattenWorkloadIdentityConfig(config)
	if got, want := flattened[0]["jwks_url"], "https://issuer.example.com/.well-known/jwks.json"; got != want {
		t.Errorf("jwks_url = %q, want %q", got, want)
	}
}
