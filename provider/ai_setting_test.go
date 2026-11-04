package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"google.golang.org/protobuf/proto"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func TestAISettingExcludesDarkLaunchedReview(t *testing.T) {
	for name, resource := range map[string]*schema.Resource{"resource": resourceSetting(), "data_source": dataSourceSetting()} {
		t.Run(name, func(t *testing.T) {
			fields := resource.Schema["ai"].Elem.(*schema.Resource).Schema
			if fields["sql_review_enabled"] != nil {
				t.Fatal("dark-launched AI SQL review must not be exposed in the Terraform schema")
			}
		})
	}
	t.Run("flatten", func(t *testing.T) {
		setting := aiSettingResponse().GetValue().GetAi()
		raw := flattenAISetting(setting, "")[0].(map[string]interface{})
		if _, ok := raw["sql_review_enabled"]; ok {
			t.Fatal("dark-launched AI SQL review must not be written to Terraform state")
		}
		if !raw["enabled"].(bool) || raw["model"] != "model-a" {
			t.Fatalf("ordinary AI fields were lost: %v", raw)
		}
	})
}

func TestAISettingSchema(t *testing.T) {
	for _, resource := range []*schema.Resource{resourceSetting(), dataSourceSetting()} {
		if resource.Schema["ai"] == nil {
			t.Fatal("ai setting schema is missing")
		}
		fields := resource.Schema["ai"].Elem.(*schema.Resource).Schema
		for _, field := range []string{"enabled", "provider", "endpoint", "api_key", "model", "version"} {
			if fields[field] == nil {
				t.Fatalf("ai.%s schema is missing", field)
			}
		}
		if !fields["api_key"].Sensitive {
			t.Fatal("AI API key must be sensitive")
		}
		if !fields["version"].Computed || fields["version"].Optional {
			t.Fatal("version must be read-only because Bytebase 3.24 rejects its update mask")
		}
		if diags := resource.Schema["name"].ValidateDiagFunc("settings/AI", nil); diags.HasError() {
			t.Fatalf("settings/AI name rejected: %v", diags)
		}
	}
	provider := resourceSetting().Schema["ai"].Elem.(*schema.Resource).Schema["provider"]
	for _, value := range []string{"PROVIDER_UNSPECIFIED", "OPEN_AI", "CLAUDE", "GEMINI", "AZURE_OPENAI"} {
		if _, errs := provider.ValidateFunc(value, "ai.0.provider"); len(errs) != 0 {
			t.Fatalf("provider %q rejected: %v", value, errs)
		}
	}
	if _, errs := provider.ValidateFunc("INVALID", "ai.0.provider"); len(errs) == 0 {
		t.Fatal("unknown AI provider accepted")
	}
}

func TestAISettingRequiresBlock(t *testing.T) {
	r := resourceSetting()
	raw := map[string]interface{}{"name": "settings/AI"}
	config := terraform.NewResourceConfigRaw(raw)
	value, err := schema.JSONMapToStateValue(raw, r.CoreConfigSchema())
	if err != nil {
		t.Fatal(err)
	}
	config.CtyValue = value
	if _, err := r.Diff(context.Background(), &terraform.InstanceState{RawConfig: value}, config, nil); err == nil {
		t.Fatal("settings/AI accepted without ai block")
	}
}

func TestAISettingUpsertPreservesRedactedKey(t *testing.T) {
	client := &settingTestClient{response: aiSettingResponse()}
	raw := map[string]interface{}{
		"name": "settings/AI",
		"ai": []interface{}{map[string]interface{}{
			"enabled": true, "provider": "CLAUDE", "endpoint": "https://ai.example.com",
			"api_key": "secret-key", "model": "model-a",
		}},
	}
	state := applySettingTestConfig(t, raw, nil, client)
	want := &v1pb.AISetting{
		Enabled: true, Provider: v1pb.AISetting_CLAUDE, Endpoint: "https://ai.example.com",
		ApiKey: "secret-key", Model: "model-a",
	}
	if !proto.Equal(client.patch.GetValue().GetAi(), want) {
		t.Fatalf("AI patch = %v, want %v", client.patch.GetValue().GetAi(), want)
	}
	wantMasks := []string{"value.ai.enabled", "value.ai.provider", "value.ai.endpoint", "value.ai.api_key", "value.ai.model"}
	if !reflect.DeepEqual(client.masks, wantMasks) {
		t.Fatalf("AI masks = %v, want %v", client.masks, wantMasks)
	}
	if got := state.Attributes["ai.0.api_key"]; got != "secret-key" {
		t.Fatalf("refreshed API key = %q, want configured key despite server redaction", got)
	}
	if got := state.Attributes["ai.0.version"]; got != "2024-06-01" {
		t.Fatalf("read version = %q", got)
	}
	if diff := settingTestDiff(t, raw, state); diff != nil && !diff.Empty() {
		t.Fatalf("redacted API key causes perpetual diff: %#v", diff.Attributes)
	}

	raw["ai"].([]interface{})[0].(map[string]interface{})["model"] = "model-b"
	client.response.GetValue().GetAi().Model = "model-b"
	state = applySettingTestConfig(t, raw, state, client)
	for _, mask := range client.masks {
		if mask == "value.ai.api_key" {
			t.Fatal("unrelated update must not restore an unchanged, potentially stale API key")
		}
	}
	if client.patch.GetValue().GetAi().GetApiKey() != "" {
		t.Fatal("unrelated update should not send an unchanged API key")
	}

	raw["ai"].([]interface{})[0].(map[string]interface{})["api_key"] = "rotated-key"
	rotated := applySettingTestConfig(t, raw, state, client)
	if got := client.patch.GetValue().GetAi().GetApiKey(); got != "rotated-key" {
		t.Fatalf("rotated API key patch = %q", got)
	}
	if got := rotated.Attributes["ai.0.api_key"]; got != "rotated-key" {
		t.Fatalf("rotated API key state = %q", got)
	}
}

func TestAISettingUpsertPreservesUnmanagedFields(t *testing.T) {
	client := &settingTestClient{response: aiSettingResponse()}
	client.response.GetValue().GetAi().Enabled = false
	raw := map[string]interface{}{
		"name": "settings/AI",
		"ai":   []interface{}{map[string]interface{}{"enabled": false}},
	}
	state := applySettingTestConfig(t, raw, nil, client)
	if !reflect.DeepEqual(client.masks, []string{"value.ai.enabled"}) {
		t.Fatalf("AI masks = %v, want only explicitly configured AI enabled flag", client.masks)
	}
	if client.patch.GetValue().GetAi().GetApiKey() != "" {
		t.Fatal("omitted API key should not be sent")
	}
	if got := state.Attributes["ai.0.endpoint"]; got != "https://ai.example.com" {
		t.Fatalf("unmanaged endpoint read = %q", got)
	}
}

func TestAISettingDataSourceReadsRedactedKey(t *testing.T) {
	d := schema.TestResourceDataRaw(t, dataSourceSetting().Schema, map[string]interface{}{"name": "settings/AI"})
	if diags := dataSourceSettingRead(context.Background(), d, &settingTestClient{response: aiSettingResponse()}); diags.HasError() {
		t.Fatalf("data source read: %v", diags)
	}
	if got := d.Get("ai.0.provider"); got != "CLAUDE" {
		t.Fatalf("AI provider = %v", got)
	}
	if got := d.Get("ai.0.api_key"); got != "" {
		t.Fatalf("redacted data source API key = %v", got)
	}
}

func TestAISettingDestroyPreservesServerConfiguration(t *testing.T) {
	state := &terraform.InstanceState{ID: "settings/AI", Attributes: map[string]string{
		"name": "settings/AI", "ai.#": "1", "ai.0.enabled": "true",
	}}
	client := &settingTestClient{response: aiSettingResponse()}
	updated, diags := resourceSetting().Apply(context.Background(), state, &terraform.InstanceDiff{Destroy: true}, client)
	if diags.HasError() {
		t.Fatalf("destroy AI setting: %v", diags)
	}
	if updated != nil {
		t.Fatalf("destroy retained Terraform state: %v", updated)
	}
	if client.patch != nil {
		t.Fatalf("destroy must preserve server AI settings, sent patch %v", client.patch)
	}
	if len(diags) != 1 || diags[0].Severity != diag.Warning {
		t.Fatalf("destroy must warn that AI settings are retained: %v", diags)
	}
}

func aiSettingResponse() *v1pb.Setting {
	return &v1pb.Setting{
		Name: "settings/AI",
		Value: &v1pb.SettingValue{Value: &v1pb.SettingValue_Ai{Ai: &v1pb.AISetting{
			Enabled: true, Provider: v1pb.AISetting_CLAUDE, Endpoint: "https://ai.example.com",
			Model: "model-a", Version: "2024-06-01", SqlReviewEnabled: true,
		}}},
	}
}

type settingTestClient struct {
	api.Client
	response *v1pb.Setting
	patch    *v1pb.Setting
	masks    []string
}

func (c *settingTestClient) GetSetting(context.Context, string) (*v1pb.Setting, error) {
	return c.response, nil
}

func (c *settingTestClient) UpsertSetting(_ context.Context, setting *v1pb.Setting, masks []string) (*v1pb.Setting, error) {
	c.patch = proto.Clone(setting).(*v1pb.Setting)
	c.masks = append([]string(nil), masks...)
	return c.response, nil
}

func settingTestDiff(t *testing.T, raw map[string]interface{}, state *terraform.InstanceState) *terraform.InstanceDiff {
	t.Helper()
	r := resourceSetting()
	config := terraform.NewResourceConfigRaw(raw)
	value, err := schema.JSONMapToStateValue(raw, r.CoreConfigSchema())
	if err != nil {
		t.Fatalf("coerce setting config: %v", err)
	}
	config.CtyValue = value
	planningState := &terraform.InstanceState{RawConfig: value}
	if state != nil {
		planningState = state.DeepCopy()
		planningState.RawConfig = value
	}
	diff, err := r.Diff(context.Background(), planningState, config, nil)
	if err != nil {
		t.Fatalf("setting diff: %v", err)
	}
	if diff != nil {
		diff.RawConfig = value
	}
	return diff
}

func applySettingTestConfig(t *testing.T, raw map[string]interface{}, state *terraform.InstanceState, client api.Client) *terraform.InstanceState {
	t.Helper()
	diff := settingTestDiff(t, raw, state)
	updated, diags := resourceSetting().Apply(context.Background(), state, diff, client)
	if diags.HasError() {
		t.Fatalf("apply setting config: %v", diags)
	}
	return updated
}
