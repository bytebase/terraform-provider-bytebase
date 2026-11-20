package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
)

func TestExportSettingRefreshKeepsLocalState(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(fmt.Sprint(existing), func(t *testing.T) {
			client := &exportSettingTestClient{}
			fields := map[string]interface{}{}
			wantVersion := "0"
			if existing {
				client.tenant, client.clientID, client.secret = "tenant-a", "app-a", "old-secret"
				fields["client_secret"], fields["client_secret_version"] = "refresh-secret", 1
				wantVersion = "1"
			}
			state := applySettingTestConfig(t, exportSettingConfig(fields), nil, client)
			client.tenant, client.clientID, client.secret = "", "", ""
			client.patch = nil
			r := resourceSetting()
			state.RawConfig = cty.NullVal(r.CoreConfigSchema().ImpliedType())
			refreshed, diags := r.RefreshWithoutUpgrade(context.Background(), state, client)
			if diags.HasError() {
				t.Fatal(diags)
			}
			if refreshed.Attributes["export.0.sharepoint.#"] != "1" || refreshed.Attributes[exportVersionPath] != wantVersion {
				t.Fatalf("refresh lost the local block or rotation marker: %v", refreshed.Attributes)
			}
			if client.patch != nil {
				t.Fatal("refresh must not write credentials")
			}
		})
	}
}

func TestExportSettingUnstoredSecretFails(t *testing.T) {
	client := &exportSettingTestClient{}
	raw := exportSettingConfig(map[string]interface{}{"client_secret": "new-secret", "client_secret_version": 1})
	state, diags := resourceSetting().Apply(context.Background(), nil, settingTestDiff(t, raw, nil), client)
	if !diags.HasError() || !strings.Contains(fmt.Sprint(diags), "tenant_id") || !strings.Contains(fmt.Sprint(diags), "client_id") {
		t.Fatalf("discarded secret needs actionable ID validation: %v", diags)
	}
	if state != nil && state.ID != "" {
		t.Fatal("discarded secret must not be reported as a successful resource")
	}
	if strings.Contains(fmt.Sprint(diags), "new-secret") {
		t.Fatal("credential appeared in diagnostics")
	}
}

func TestExportSettingFailedRotationRetainsMarker(t *testing.T) {
	client := &exportSettingTestClient{tenant: "tenant-a", clientID: "app-a", secret: "old-secret"}
	fields := map[string]interface{}{"client_secret": "first-secret", "client_secret_version": 1}
	state := applySettingTestConfig(t, exportSettingConfig(fields), nil, client)
	client.writeErr = errors.New("credential update denied")
	fields["client_secret"], fields["client_secret_version"] = "next-secret", 2
	updated, diags := resourceSetting().Apply(context.Background(), state, settingTestDiff(t, exportSettingConfig(fields), state), client)
	if !diags.HasError() {
		t.Fatal("failed credential update was accepted")
	}
	if updated == nil || updated.Attributes[exportVersionPath] != "1" {
		t.Fatalf("failed rotation advanced the local marker: %v", updated)
	}
	assertExportSecretNotInState(t, updated, "next-secret")
}

func TestAccExportSettingReadPreservesShape(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(fmt.Sprint(existing), func(t *testing.T) {
			client := &exportSettingTestClient{}
			config := `resource "bytebase_setting" "export" {
  name = "settings/EXPORT"
  export {
    sharepoint {}
  }
}`
			version := "0"
			if existing {
				client.tenant, client.clientID, client.secret = "tenant-a", "app-a", "old-secret"
				version = "1"
				config = `resource "bytebase_setting" "export" {
  name = "settings/EXPORT"
  export {
    sharepoint {
      client_secret = "refresh-secret"
      client_secret_version = 1
    }
  }
}`
			}
			p := NewProvider()
			p.ConfigureContextFunc = func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) { return client, nil }
			resource.Test(t, resource.TestCase{
				Providers: map[string]*schema.Provider{"bytebase": p},
				Steps: []resource.TestStep{
					{Config: config},
					{
						PreConfig: func() {
							client.mu.Lock()
							defer client.mu.Unlock()
							client.tenant, client.clientID, client.secret = "", "", ""
							client.patch = nil
						},
						Config: config,
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("bytebase_setting.export", "export.0.sharepoint.#", "1"),
							resource.TestCheckResourceAttr("bytebase_setting.export", exportVersionPath, version),
							func(*terraform.State) error {
								client.mu.Lock()
								defer client.mu.Unlock()
								if client.patch != nil {
									return errors.New("unchanged configuration replayed a credential after remote clear")
								}
								return nil
							},
						),
					},
				},
			})
		})
	}
}
