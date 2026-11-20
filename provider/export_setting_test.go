package provider

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

const exportSecretPath = "export.0.sharepoint.0.client_secret"
const exportVersionPath = "export.0.sharepoint.0.client_secret_version"

func TestExportSettingSchema(t *testing.T) {
	r := resourceSetting()
	block := r.Schema["export"]
	if block == nil {
		t.Fatal("EXPORT setting schema is missing")
	}
	fields := block.Elem.(*schema.Resource).Schema["sharepoint"].Elem.(*schema.Resource).Schema
	if !fields["client_secret"].WriteOnly || !fields["client_secret"].Sensitive || fields["client_secret"].Computed {
		t.Fatal("SharePoint secret must be sensitive and write-only, not computed")
	}
	if block.Computed || block.Elem.(*schema.Resource).Schema["sharepoint"].Computed {
		t.Fatal("write-only secret ancestors must not be computed")
	}
	ds := dataSourceSetting()
	dsFields := ds.Schema["export"].Elem.(*schema.Resource).Schema["sharepoint"].Elem.(*schema.Resource).Schema
	if dsFields["client_secret"] != nil || dsFields["client_secret_version"] != nil {
		t.Fatal("data source must not expose secrets or the local rotation marker")
	}
	for _, resource := range []*schema.Resource{r, ds} {
		if diags := resource.Schema["name"].ValidateDiagFunc("settings/EXPORT", nil); diags.HasError() {
			t.Fatal(diags)
		}
	}
}

func TestExportSettingCredentialLifecycle(t *testing.T) {
	client := &exportSettingTestClient{}
	raw := exportSettingConfig(map[string]interface{}{
		"tenant_id": "tenant-a", "client_id": "app-a", "client_secret": "initial-secret", "client_secret_version": 1,
	})
	state := applySettingTestConfig(t, raw, nil, client)
	if client.patch.GetValue().GetExport().GetSharepoint().GetClientSecret() != "initial-secret" {
		t.Fatal("initial credential was not sent")
	}
	if !slices.Equal(client.masks, []string{"value.export.sharepoint.tenant_id", "value.export.sharepoint.client_id", "value.export.sharepoint.client_secret"}) {
		t.Fatalf("credential masks = %v", client.masks)
	}
	assertExportSecretNotInState(t, state, "initial-secret")
	if state.Attributes[exportVersionPath] != "1" || state.Attributes["export.0.sharepoint.0.client_secret_set"] != "true" {
		t.Fatalf("credential status/version was not refreshed: %v", state.Attributes)
	}
	assertExportNoManagedDiff(t, raw, state)

	client.secret = "externally-rotated-secret"
	fields := raw["export"].([]interface{})[0].(map[string]interface{})["sharepoint"].([]interface{})[0].(map[string]interface{})
	fields["client_id"] = "app-b"
	state = applySettingTestConfig(t, raw, state, client)
	if slices.Contains(client.masks, "value.export.sharepoint.client_secret") || client.patch.GetValue().GetExport().GetSharepoint().GetClientSecret() != "" {
		t.Fatal("unrelated update replayed an unchanged secret")
	}
	if client.secret != "externally-rotated-secret" {
		t.Fatal("unrelated update replaced the externally rotated secret")
	}
	fields["client_secret"] = "next-secret"
	fields["client_secret_version"] = 2
	state = applySettingTestConfig(t, raw, state, client)
	if !slices.Contains(client.masks, "value.export.sharepoint.client_secret") || client.secret != "next-secret" {
		t.Fatal("rotation marker did not send the new credential")
	}
	assertExportSecretNotInState(t, state, "next-secret")
	if state.Attributes[exportVersionPath] != "2" {
		t.Fatalf("rotation marker = %q", state.Attributes[exportVersionPath])
	}

	raw = map[string]interface{}{"name": "settings/EXPORT", "export": []interface{}{map[string]interface{}{}}}
	state = applySettingTestConfig(t, raw, state, client)
	if client.secret != "" || state.Attributes["export.0.sharepoint.#"] != "0" {
		t.Fatalf("explicit clear left credentials configured: %v", state.Attributes)
	}
	assertExportNoManagedDiff(t, raw, state)
}

func TestExportSettingOmittedLeavesPreserveCredentials(t *testing.T) {
	client := &exportSettingTestClient{tenant: "existing-tenant", clientID: "existing-app", secret: "existing-secret"}
	raw := exportSettingConfig(map[string]interface{}{"client_id": "updated-app"})
	state := applySettingTestConfig(t, raw, nil, client)
	if !slices.Equal(client.masks, []string{"value.export.sharepoint.client_id"}) || client.tenant != "existing-tenant" || client.secret != "existing-secret" {
		t.Fatalf("omitted credentials changed: masks=%v, tenant=%s", client.masks, client.tenant)
	}
	if state.Attributes["export.0.sharepoint.0.tenant_id"] != "existing-tenant" {
		t.Fatal("existing tenant was not read into state")
	}
	client.patch = nil
	if diff := settingTestDiff(t, exportSettingConfig(map[string]interface{}{}), state); diff != nil && !diff.Empty() {
		t.Fatalf("omitting managed credential fields should not create a change: %v", diff.Attributes)
	}
	state = applySettingTestConfig(t, exportSettingConfig(map[string]interface{}{}), nil, client)
	if client.patch != nil || client.secret != "existing-secret" {
		t.Fatal("empty credential configuration wrote existing credentials")
	}
	assertExportSecretNotInState(t, state, "existing-secret")
}

func TestExportSettingReadImportAndDestroy(t *testing.T) {
	client := &exportSettingTestClient{tenant: "existing-tenant", clientID: "existing-app", secret: "existing-secret"}
	r := resourceSetting()
	d := schema.TestResourceDataRaw(t, r.Schema, nil)
	d.SetId("settings/EXPORT")
	if _, err := r.Importer.StateContext(context.Background(), d, client); err != nil {
		t.Fatal(err)
	}
	if diags := r.ReadContext(context.Background(), d, client); diags.HasError() {
		t.Fatal(diags)
	}
	secretSet, _ := d.Get("export.0.sharepoint.0.client_secret_set").(bool)
	if d.Get("name") != "settings/EXPORT" || !secretSet {
		t.Fatal("import did not read setting identity and credential status")
	}
	assertExportSecretNotInState(t, d.State(), "existing-secret")
	ds := dataSourceSetting()
	read := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{"name": "settings/EXPORT"})
	if diags := ds.ReadContext(context.Background(), read, client); diags.HasError() {
		t.Fatal(diags)
	}
	if got := read.Get("export.0.sharepoint.0.tenant_id"); got != "existing-tenant" {
		t.Fatalf("data source tenant = %v", got)
	}
	if diags := r.DeleteContext(context.Background(), d, client); diags.HasError() || len(diags) == 0 {
		t.Fatalf("destroy must preserve credentials with a warning: %v", diags)
	}
	if d.Id() != "" || client.patch != nil || client.secret != "existing-secret" {
		t.Fatal("destroy changed shared workspace credentials")
	}
}

func TestExportSettingValidation(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  map[string]interface{}
	}{
		{"missing export", map[string]interface{}{"name": "settings/EXPORT"}},
		{"empty IDs", exportSettingConfig(map[string]interface{}{"tenant_id": "", "client_id": ""})},
		{"secret without version", exportSettingConfig(map[string]interface{}{"client_secret": "new-secret"})},
		{"version without secret", exportSettingConfig(map[string]interface{}{"client_secret_version": 1})},
		{"empty secret", exportSettingConfig(map[string]interface{}{"client_secret": "", "client_secret_version": 1})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := resourceSetting()
			value, err := schema.JSONMapToStateValue(tc.raw, r.CoreConfigSchema())
			if err != nil {
				t.Fatal(err)
			}
			config := terraform.NewResourceConfigRaw(tc.raw)
			config.CtyValue = value
			if _, err := r.Diff(context.Background(), &terraform.InstanceState{RawConfig: value}, config, nil); err == nil {
				t.Fatal("invalid export configuration was accepted")
			}
		})
	}
}

func TestExportSettingUnknownSecretPlan(t *testing.T) {
	r := resourceSetting()
	value, err := schema.JSONMapToStateValue(exportSettingConfig(map[string]interface{}{"client_secret": "new-secret", "client_secret_version": 1}), r.CoreConfigSchema())
	if err != nil {
		t.Fatal(err)
	}
	attrs := value.AsValueMap()
	export := attrs["export"].AsValueSlice()[0].AsValueMap()
	sharepoint := export["sharepoint"].AsValueSlice()[0].AsValueMap()
	sharepoint["client_secret"] = cty.UnknownVal(cty.String)
	sharepoint["client_secret_version"] = cty.UnknownVal(cty.Number)
	export["sharepoint"] = cty.ListVal([]cty.Value{cty.ObjectVal(sharepoint)})
	attrs["export"] = cty.ListVal([]cty.Value{cty.ObjectVal(export)})
	value = cty.ObjectVal(attrs)
	config := terraform.NewResourceConfigShimmed(value, r.CoreConfigSchema())
	if _, err := r.Diff(context.Background(), &terraform.InstanceState{RawConfig: value}, config, nil); err != nil {
		t.Fatalf("unknown credentials must be deferred until apply: %v", err)
	}
}

func TestAccExportSettingSecretLifecycle(t *testing.T) {
	const address = "bytebase_setting.export"
	client := &exportSettingTestClient{}
	p := NewProvider()
	p.ConfigureContextFunc = func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return client, nil
	}
	config := func(tenant, app, secret string, version int) string {
		if tenant == "" && app == "" {
			return `resource "bytebase_setting" "export" {
  name = "settings/EXPORT"
  export {}
}`
		}
		return fmt.Sprintf(`
variable "sharepoint_secret" {
  type = string
  sensitive = true
  ephemeral = true
  default = %q
}
resource "bytebase_setting" "export" {
  name = "settings/EXPORT"
  export {
    sharepoint {
      tenant_id = %q
      client_id = %q
      client_secret = var.sharepoint_secret
      client_secret_version = %d
    }
  }
}
data "bytebase_setting" "export" {
  name = "settings/EXPORT"
  depends_on = [bytebase_setting.export]
}
`, secret, tenant, app, version)
	}
	check := func(secret string, sent bool, version int) resource.TestCheckFunc {
		return func(state *terraform.State) error {
			client.mu.Lock()
			defer client.mu.Unlock()
			if client.secret != secret || slices.Contains(client.masks, "value.export.sharepoint.client_secret") != sent {
				return errors.Errorf("credential matches expected=%v; secret mask=%v (want %v); masks=%v", client.secret == secret, slices.Contains(client.masks, "value.export.sharepoint.client_secret"), sent, client.masks)
			}
			attrs := state.RootModule().Resources[address].Primary.Attributes
			if version >= 0 && attrs[exportVersionPath] != fmt.Sprint(version) {
				return errors.New("local rotation marker did not survive refresh")
			}
			for key, value := range attrs {
				if strings.Contains(value, "initial-secret") || strings.Contains(value, "next-secret") || key == exportSecretPath && value != "" {
					return errors.New("SharePoint secret was stored in Terraform state")
				}
			}
			return nil
		}
	}
	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"bytebase": p},
		Steps: []resource.TestStep{
			{Config: config("tenant-a", "app-a", "initial-secret", 1), Check: check("initial-secret", true, 1)},
			{
				PreConfig: func() {
					client.mu.Lock()
					defer client.mu.Unlock()
					client.secret = "externally-rotated-secret"
				},
				Config: config("tenant-a", "app-b", "initial-secret", 1),
				Check:  check("externally-rotated-secret", false, 1),
			},
			{Config: config("tenant-a", "app-b", "next-secret", 2), Check: check("next-secret", true, 2)},
			{ResourceName: address, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{exportVersionPath}},
			{
				Config: config("", "", "next-secret", 2),
				Check: resource.ComposeTestCheckFunc(
					check("", false, -1),
					resource.TestCheckResourceAttr(address, "export.0.sharepoint.#", "0"),
				),
			},
		},
	})
}

func assertExportNoManagedDiff(t *testing.T, raw map[string]interface{}, state *terraform.InstanceState) {
	t.Helper()
	if diff := settingTestDiff(t, raw, state); diff != nil {
		// The legacy unit API retains write-only inputs in diffs; the Terraform
		// CLI acceptance test checks that real plans converge after redaction.
		for field := range diff.Attributes {
			if field != exportSecretPath {
				t.Fatalf("credential refresh causes a repeated change to %s", field)
			}
		}
	}
}

func exportSettingConfig(fields map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name": "settings/EXPORT",
		"export": []interface{}{map[string]interface{}{
			"sharepoint": []interface{}{fields},
		}},
	}
}

func assertExportSecretNotInState(t *testing.T, state *terraform.InstanceState, secret string) {
	t.Helper()
	for key, value := range state.Attributes {
		if strings.Contains(value, secret) || key == exportSecretPath && value != "" {
			t.Fatalf("write-only credential leaked into state at %s", key)
		}
	}
}

type exportSettingTestClient struct {
	api.Client
	mu                       sync.Mutex
	tenant, clientID, secret string
	patch                    *v1pb.Setting
	masks                    []string
	writeErr                 error
}

func (c *exportSettingTestClient) GetSetting(context.Context, string) (*v1pb.Setting, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.response(), nil
}

func (c *exportSettingTestClient) response() *v1pb.Setting {
	setting := &v1pb.ExportSetting{}
	if c.tenant != "" || c.clientID != "" {
		setting.Sharepoint = &v1pb.SharePointSetting{TenantId: c.tenant, ClientId: c.clientID, ClientSecretSet: c.secret != ""}
	}
	return &v1pb.Setting{Name: "settings/EXPORT", Value: &v1pb.SettingValue{Value: &v1pb.SettingValue_Export{Export: setting}}}
}

func (c *exportSettingTestClient) UpsertSetting(_ context.Context, setting *v1pb.Setting, masks []string) (*v1pb.Setting, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.patch = proto.Clone(setting).(*v1pb.Setting)
	c.masks = slices.Clone(masks)
	if c.writeErr != nil {
		return nil, c.writeErr
	}
	value := setting.GetValue().GetExport().GetSharepoint()
	for _, mask := range masks {
		switch mask {
		case "value.export.sharepoint.tenant_id":
			c.tenant = strings.TrimSpace(value.GetTenantId())
		case "value.export.sharepoint.client_id":
			c.clientID = strings.TrimSpace(value.GetClientId())
		case "value.export.sharepoint.client_secret":
			if value.GetClientSecret() != "" {
				c.secret = value.GetClientSecret()
			}
		default:
			return nil, errors.Errorf("unexpected export mask %s", mask)
		}
	}
	if c.tenant == "" && c.clientID == "" {
		c.secret = ""
	} else if c.tenant == "" || c.clientID == "" || c.secret == "" {
		return nil, errors.New("tenant ID, client ID, and client secret are all required")
	}
	return c.response(), nil
}
