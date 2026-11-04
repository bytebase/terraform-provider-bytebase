package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"google.golang.org/protobuf/proto"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
)

func TestAppIMSettingUpsertMasksAllProviders(t *testing.T) {
	raw := appIMSettingTestConfig()
	client := &settingTestClient{response: appIMSettingTestResponse(
		v1pb.WebhookType_SLACK, v1pb.WebhookType_FEISHU, v1pb.WebhookType_WECOM,
		v1pb.WebhookType_LARK, v1pb.WebhookType_DINGTALK, v1pb.WebhookType_TEAMS,
	)}
	applySettingTestConfig(t, raw, nil, client)
	assertAppIMSettingTestMasks(t, client.masks)
	want := &v1pb.AppIMSetting{Settings: []*v1pb.AppIMSetting_IMSetting{
		{Type: v1pb.WebhookType_SLACK, Payload: &v1pb.AppIMSetting_IMSetting_Slack{Slack: &v1pb.AppIMSetting_Slack{Token: "slack-token"}}},
		{Type: v1pb.WebhookType_FEISHU, Payload: &v1pb.AppIMSetting_IMSetting_Feishu{Feishu: &v1pb.AppIMSetting_Feishu{AppId: "feishu-app", AppSecret: "feishu-secret"}}},
		{Type: v1pb.WebhookType_WECOM, Payload: &v1pb.AppIMSetting_IMSetting_Wecom{Wecom: &v1pb.AppIMSetting_Wecom{CorpId: "wecom-corp", AgentId: "wecom-agent", Secret: "wecom-secret"}}},
		{Type: v1pb.WebhookType_LARK, Payload: &v1pb.AppIMSetting_IMSetting_Lark{Lark: &v1pb.AppIMSetting_Lark{AppId: "lark-app", AppSecret: "lark-secret"}}},
		{Type: v1pb.WebhookType_DINGTALK, Payload: &v1pb.AppIMSetting_IMSetting_Dingtalk{Dingtalk: &v1pb.AppIMSetting_DingTalk{ClientId: "dingtalk-client", ClientSecret: "dingtalk-secret", RobotCode: "dingtalk-robot"}}},
		{Type: v1pb.WebhookType_TEAMS, Payload: &v1pb.AppIMSetting_IMSetting_Teams{Teams: &v1pb.AppIMSetting_Teams{TenantId: "teams-tenant", ClientId: "teams-client", ClientSecret: "teams-secret"}}},
	}}
	if got := client.patch.GetValue().GetAppIm(); !proto.Equal(got, want) {
		t.Fatalf("APP_IM patch = %v, want %v", got, want)
	}
}

func TestAppIMSettingUpdateUsesConfiguredCredentials(t *testing.T) {
	for _, tc := range []struct {
		provider   string
		field      string
		imType     v1pb.WebhookType
		credential func(*v1pb.AppIMSetting_IMSetting) string
	}{
		{"slack", "token", v1pb.WebhookType_SLACK, func(s *v1pb.AppIMSetting_IMSetting) string { return s.GetSlack().GetToken() }},
		{"feishu", "app_secret", v1pb.WebhookType_FEISHU, func(s *v1pb.AppIMSetting_IMSetting) string { return s.GetFeishu().GetAppSecret() }},
		{"wecom", "secret", v1pb.WebhookType_WECOM, func(s *v1pb.AppIMSetting_IMSetting) string { return s.GetWecom().GetSecret() }},
		{"lark", "app_secret", v1pb.WebhookType_LARK, func(s *v1pb.AppIMSetting_IMSetting) string { return s.GetLark().GetAppSecret() }},
		{"dingtalk", "client_secret", v1pb.WebhookType_DINGTALK, func(s *v1pb.AppIMSetting_IMSetting) string { return s.GetDingtalk().GetClientSecret() }},
		{"teams", "client_secret", v1pb.WebhookType_TEAMS, func(s *v1pb.AppIMSetting_IMSetting) string { return s.GetTeams().GetClientSecret() }},
	} {
		t.Run(tc.provider, func(t *testing.T) {
			block := appIMSettingTestConfig()["app_im"].([]interface{})[0].(map[string]interface{})[tc.provider]
			raw := map[string]interface{}{
				"name": "settings/APP_IM", "app_im": []interface{}{map[string]interface{}{tc.provider: block}},
			}
			client := &settingTestClient{response: appIMSettingTestResponse(tc.imType)}
			state := applySettingTestConfig(t, raw, nil, client)
			block.([]interface{})[0].(map[string]interface{})[tc.field] = "rotated-secret"
			applySettingTestConfig(t, raw, state, client)
			assertAppIMSettingTestMasks(t, client.masks)
			settings := client.patch.GetValue().GetAppIm().GetSettings()
			if len(settings) != 1 || settings[0].GetType() != tc.imType || tc.credential(settings[0]) != "rotated-secret" {
				t.Fatalf("APP_IM update did not send the configured %s credential: %v", tc.provider, settings)
			}
		})
	}
}

func TestAppIMSettingRemovesOmittedProvider(t *testing.T) {
	raw := appIMSettingTestConfig()
	client := &settingTestClient{response: appIMSettingTestResponse(
		v1pb.WebhookType_SLACK, v1pb.WebhookType_FEISHU, v1pb.WebhookType_WECOM,
		v1pb.WebhookType_LARK, v1pb.WebhookType_DINGTALK, v1pb.WebhookType_TEAMS,
	)}
	state := applySettingTestConfig(t, raw, nil, client)
	delete(raw["app_im"].([]interface{})[0].(map[string]interface{}), "teams")
	client.response = appIMSettingTestResponse(
		v1pb.WebhookType_SLACK, v1pb.WebhookType_FEISHU, v1pb.WebhookType_WECOM,
		v1pb.WebhookType_LARK, v1pb.WebhookType_DINGTALK,
	)
	state = applySettingTestConfig(t, raw, state, client)
	assertAppIMSettingTestMasks(t, client.masks)
	for _, integration := range client.patch.GetValue().GetAppIm().GetSettings() {
		if integration.GetType() == v1pb.WebhookType_TEAMS {
			t.Fatal("omitted Teams integration must have no payload so its mask removes it")
		}
	}
	if got := state.Attributes["app_im.0.teams.#"]; got != "0" {
		t.Fatalf("omitted Teams block count = %q, want 0", got)
	}
}

func TestAppIMSettingOmittedProvidersAreRemovedOnCreate(t *testing.T) {
	client := &settingTestClient{response: appIMSettingTestResponse()}
	applySettingTestConfig(t, map[string]interface{}{
		"name": "settings/APP_IM", "app_im": []interface{}{map[string]interface{}{}},
	}, nil, client)
	assertAppIMSettingTestMasks(t, client.masks)
	if got := client.patch.GetValue().GetAppIm(); got == nil || len(got.GetSettings()) != 0 {
		t.Fatalf("empty app_im block must remove every provider, got payload %v", got)
	}
}

func TestAppIMSettingDestroyMasksAllProviders(t *testing.T) {
	state := &terraform.InstanceState{ID: "settings/APP_IM", Attributes: map[string]string{
		"name": "settings/APP_IM", "app_im.#": "1", "app_im.0.slack.#": "1",
	}}
	client := &settingTestClient{response: appIMSettingTestResponse()}
	updated, diags := resourceSetting().Apply(context.Background(), state, &terraform.InstanceDiff{Destroy: true}, client)
	if diags.HasError() {
		t.Fatalf("destroy APP_IM setting: %v", diags)
	}
	if updated != nil {
		t.Fatalf("destroy retained Terraform state: %v", updated)
	}
	assertAppIMSettingTestMasks(t, client.masks)
	if got := client.patch.GetValue().GetAppIm(); got == nil || len(got.GetSettings()) != 0 {
		t.Fatalf("destroy must send no provider payloads, got %v", got)
	}
}

func TestAppIMSettingRequiresBlock(t *testing.T) {
	r := resourceSetting()
	raw := map[string]interface{}{"name": "settings/APP_IM"}
	config := terraform.NewResourceConfigRaw(raw)
	value, err := schema.JSONMapToStateValue(raw, r.CoreConfigSchema())
	if err != nil {
		t.Fatal(err)
	}
	config.CtyValue = value
	if _, err := r.Diff(context.Background(), &terraform.InstanceState{RawConfig: value}, config, nil); err == nil {
		t.Fatal("settings/APP_IM accepted without app_im block")
	}
}

func assertAppIMSettingTestMasks(t *testing.T, got []string) {
	t.Helper()
	want := []string{
		"value.app_im.slack", "value.app_im.feishu", "value.app_im.wecom",
		"value.app_im.lark", "value.app_im.dingtalk", "value.app_im_setting_value.teams",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("APP_IM masks = %v, want %v", got, want)
	}
}

func appIMSettingTestConfig() map[string]interface{} {
	return map[string]interface{}{
		"name": "settings/APP_IM",
		"app_im": []interface{}{map[string]interface{}{
			"slack":    []interface{}{map[string]interface{}{"token": "slack-token"}},
			"feishu":   []interface{}{map[string]interface{}{"app_id": "feishu-app", "app_secret": "feishu-secret"}},
			"wecom":    []interface{}{map[string]interface{}{"corp_id": "wecom-corp", "agent_id": "wecom-agent", "secret": "wecom-secret"}},
			"lark":     []interface{}{map[string]interface{}{"app_id": "lark-app", "app_secret": "lark-secret"}},
			"dingtalk": []interface{}{map[string]interface{}{"client_id": "dingtalk-client", "client_secret": "dingtalk-secret", "robot_code": "dingtalk-robot"}},
			"teams":    []interface{}{map[string]interface{}{"tenant_id": "teams-tenant", "client_id": "teams-client", "client_secret": "teams-secret"}},
		}},
	}
}

func appIMSettingTestResponse(types ...v1pb.WebhookType) *v1pb.Setting {
	integrations := make([]*v1pb.AppIMSetting_IMSetting, 0, len(types))
	for _, imType := range types {
		integrations = append(integrations, &v1pb.AppIMSetting_IMSetting{Type: imType})
	}
	return &v1pb.Setting{
		Name: "settings/APP_IM",
		Value: &v1pb.SettingValue{Value: &v1pb.SettingValue_AppIm{AppIm: &v1pb.AppIMSetting{
			Settings: integrations,
		}}},
	}
}
