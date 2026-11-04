package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
)

func TestMCPSettingCapabilityValidation(t *testing.T) {
	r := resourceSetting()
	if r.Schema["mcp"] == nil {
		t.Fatal("mcp setting schema is missing")
	}
	field := r.Schema["mcp"].Elem.(*schema.Resource).Schema["capability"]
	if !field.Required {
		t.Fatal("mcp.capability must be explicitly configured")
	}
	for _, value := range []string{"DISABLED", "READ_ONLY", "READ_WRITE"} {
		if _, errs := field.ValidateFunc(value, "mcp.0.capability"); len(errs) != 0 {
			t.Fatalf("capability %q rejected: %v", value, errs)
		}
	}
	for _, value := range []string{"", "CAPABILITY_UNSPECIFIED", "METADATA_ONLY", "INVALID"} {
		if _, errs := field.ValidateFunc(value, "mcp.0.capability"); len(errs) == 0 {
			t.Fatalf("unsupported capability %q accepted", value)
		}
	}
	for _, r := range []*schema.Resource{r, dataSourceSetting()} {
		if diags := r.Schema["name"].ValidateDiagFunc("settings/MCP", nil); diags.HasError() {
			t.Fatalf("settings/MCP name rejected: %v", diags)
		}
	}
}

func TestMCPSettingRequiresBlock(t *testing.T) {
	r := resourceSetting()
	raw := map[string]interface{}{"name": "settings/MCP"}
	config := terraform.NewResourceConfigRaw(raw)
	value, err := schema.JSONMapToStateValue(raw, r.CoreConfigSchema())
	if err != nil {
		t.Fatal(err)
	}
	config.CtyValue = value
	if _, err := r.Diff(context.Background(), &terraform.InstanceState{RawConfig: value}, config, nil); err == nil {
		t.Fatal("settings/MCP accepted without mcp block")
	}
}

func TestMCPSettingLifecycle(t *testing.T) {
	var state *terraform.InstanceState
	for _, tc := range []struct {
		capability string
		want       v1pb.MCPSetting_Capability
	}{
		{capability: "READ_WRITE", want: v1pb.MCPSetting_READ_WRITE},
		{capability: "READ_ONLY", want: v1pb.MCPSetting_READ_ONLY},
		{capability: "DISABLED", want: v1pb.MCPSetting_DISABLED},
	} {
		client := &settingTestClient{response: mcpSettingResponse(tc.want)}
		raw := map[string]interface{}{
			"name": "settings/MCP",
			"mcp":  []interface{}{map[string]interface{}{"capability": tc.capability}},
		}
		state = applySettingTestConfig(t, raw, state, client)
		if got := client.patch.GetValue().GetMcp(); got == nil || got.GetCapability() != tc.want {
			t.Fatalf("MCP patch = %v, want %s", got, tc.capability)
		}
		if !reflect.DeepEqual(client.masks, []string{"value.mcp.capability"}) {
			t.Fatalf("MCP masks = %v, want capability-only mask", client.masks)
		}
		if got := state.Attributes["mcp.0.capability"]; got != tc.capability {
			t.Fatalf("MCP capability state = %q, want %q", got, tc.capability)
		}
		if diff := settingTestDiff(t, raw, state); diff != nil && !diff.Empty() {
			t.Fatalf("MCP read causes perpetual diff: %#v", diff.Attributes)
		}
	}
}

func TestMCPSettingRead(t *testing.T) {
	for _, r := range []*schema.Resource{resourceSetting(), dataSourceSetting()} {
		d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"name": "settings/MCP"})
		d.SetId("settings/MCP")
		client := &settingTestClient{response: mcpSettingResponse(v1pb.MCPSetting_READ_ONLY)}
		if diags := r.ReadContext(context.Background(), d, client); diags.HasError() {
			t.Fatalf("read MCP setting: %v", diags)
		}
		if got := d.Get("mcp.0.capability"); got != "READ_ONLY" {
			t.Fatalf("MCP capability = %v, want READ_ONLY", got)
		}
	}
}

func TestMCPSettingDestroyDisables(t *testing.T) {
	state := &terraform.InstanceState{ID: "settings/MCP", Attributes: map[string]string{
		"name": "settings/MCP", "mcp.#": "1", "mcp.0.capability": "READ_WRITE",
	}}
	client := &settingTestClient{response: mcpSettingResponse(v1pb.MCPSetting_READ_WRITE)}
	updated, diags := resourceSetting().Apply(context.Background(), state, &terraform.InstanceDiff{Destroy: true}, client)
	if diags.HasError() {
		t.Fatalf("destroy MCP setting: %v", diags)
	}
	if updated != nil {
		t.Fatalf("destroy retained Terraform state: %v", updated)
	}
	if got := client.patch.GetValue().GetMcp(); got == nil || got.GetCapability() != v1pb.MCPSetting_DISABLED {
		t.Fatalf("destroy MCP patch = %v, want DISABLED", got)
	}
	if !reflect.DeepEqual(client.masks, []string{"value.mcp.capability"}) {
		t.Fatalf("destroy MCP masks = %v, want capability-only mask", client.masks)
	}
}

func mcpSettingResponse(capability v1pb.MCPSetting_Capability) *v1pb.Setting {
	return &v1pb.Setting{
		Name:  "settings/MCP",
		Value: &v1pb.SettingValue{Value: &v1pb.SettingValue_Mcp{Mcp: &v1pb.MCPSetting{Capability: capability}}},
	}
}
