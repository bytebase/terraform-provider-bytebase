package provider

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
)

// getAppIMSetting returns the schema for the bb.app.im workspace setting.
//
// Every IM-message field is WriteOnly on the resource side because the server's
// store→v1 converter (bytebase/backend/api/v1/setting_service_converter.go
// convertToAppIMSetting) strips all payload fields on GET, returning only the
// list of configured IM types with empty payload structs. Modeling fields as
// regular Sensitive strings would produce perpetual refresh diffs.
//
// On the data source side (computed=true), WriteOnly is false and fields are
// computed, but their values will always be empty strings — the data source
// surfaces only which IM types are configured.
func getAppIMSetting(computed bool) *schema.Schema {
	return singletonBlock(computed,
		"The APP_IM workspace setting. Configures Slack/Feishu/Wecom/Lark/DingTalk/Teams integrations. All credential and identifier fields are write-only — the server returns empty payloads on GET, so values cannot round-trip through state.",
		map[string]*schema.Schema{
			"slack": singletonBlock(computed, "Slack integration.", map[string]*schema.Schema{
				"token": appIMField(computed, "The Slack bot token."),
			}),
			"feishu": singletonBlock(computed, "Feishu integration.", map[string]*schema.Schema{
				"app_id":     appIMField(computed, "The Feishu app id."),
				"app_secret": appIMField(computed, "The Feishu app secret."),
			}),
			"wecom": singletonBlock(computed, "WeCom integration.", map[string]*schema.Schema{
				"corp_id":  appIMField(computed, "The WeCom corp id."),
				"agent_id": appIMField(computed, "The WeCom agent id."),
				"secret":   appIMField(computed, "The WeCom secret."),
			}),
			"lark": singletonBlock(computed, "Lark integration.", map[string]*schema.Schema{
				"app_id":     appIMField(computed, "The Lark app id."),
				"app_secret": appIMField(computed, "The Lark app secret."),
			}),
			"dingtalk": singletonBlock(computed, "DingTalk integration.", map[string]*schema.Schema{
				"client_id":     appIMField(computed, "The DingTalk client id."),
				"client_secret": appIMField(computed, "The DingTalk client secret."),
				"robot_code":    appIMField(computed, "The DingTalk robot code."),
			}),
			"teams": singletonBlock(computed, "Microsoft Teams integration.", map[string]*schema.Schema{
				"tenant_id":     appIMField(computed, "The Teams tenant id."),
				"client_id":     appIMField(computed, "The Teams client id."),
				"client_secret": appIMField(computed, "The Teams client secret."),
			}),
		},
	)
}

// singletonBlock builds a TypeList schema that represents a singleton block.
// When computed, it's purely Computed (no MaxItems, no Optional). Otherwise
// it's an optional singleton (MaxItems: 1).
func singletonBlock(computed bool, description string, attrs map[string]*schema.Schema) *schema.Schema {
	s := &schema.Schema{
		Type:        schema.TypeList,
		Computed:    computed,
		Optional:    !computed,
		Description: description,
		Elem:        &schema.Resource{Schema: attrs},
	}
	if !computed {
		s.MaxItems = 1
	}
	return s
}

func appIMField(computed bool, description string) *schema.Schema {
	suffix := " This value is write-only and will not be stored in Terraform state."
	if computed {
		suffix = " The server returns empty payloads on GET, so this is always empty when read from the data source."
	}
	s := &schema.Schema{
		Type:        schema.TypeString,
		Required:    !computed,
		Computed:    computed,
		WriteOnly:   !computed,
		Description: description + suffix,
	}
	if !computed {
		s.ValidateFunc = validation.StringIsNotEmpty
	}
	return s
}

// convertToV1AppIMSetting builds an AppIMSetting from raw config. All fields
// are WriteOnly, so we read from raw config exclusively.
func convertToV1AppIMSetting(rawConfig cty.Value) (*v1pb.AppIMSetting, error) {
	if rawConfig.IsNull() {
		return nil, errors.New("app_im block is required when name is settings/APP_IM")
	}
	appIMList := rawConfig.GetAttr("app_im")
	if appIMList.IsNull() || !appIMList.IsKnown() || appIMList.LengthInt() == 0 {
		return nil, errors.New("app_im block is required when name is settings/APP_IM")
	}
	appIM := appIMList.AsValueSlice()[0]

	settings := []*v1pb.AppIMSetting_IMSetting{}

	if slack, ok := singleBlock(appIM, "slack"); ok {
		settings = append(settings, &v1pb.AppIMSetting_IMSetting{
			Type: v1pb.WebhookType_SLACK,
			Payload: &v1pb.AppIMSetting_IMSetting_Slack{
				Slack: &v1pb.AppIMSetting_Slack{
					Token: ctyString(slack, "token"),
				},
			},
		})
	}
	if feishu, ok := singleBlock(appIM, "feishu"); ok {
		settings = append(settings, &v1pb.AppIMSetting_IMSetting{
			Type: v1pb.WebhookType_FEISHU,
			Payload: &v1pb.AppIMSetting_IMSetting_Feishu{
				Feishu: &v1pb.AppIMSetting_Feishu{
					AppId:     ctyString(feishu, "app_id"),
					AppSecret: ctyString(feishu, "app_secret"),
				},
			},
		})
	}
	if wecom, ok := singleBlock(appIM, "wecom"); ok {
		settings = append(settings, &v1pb.AppIMSetting_IMSetting{
			Type: v1pb.WebhookType_WECOM,
			Payload: &v1pb.AppIMSetting_IMSetting_Wecom{
				Wecom: &v1pb.AppIMSetting_Wecom{
					CorpId:  ctyString(wecom, "corp_id"),
					AgentId: ctyString(wecom, "agent_id"),
					Secret:  ctyString(wecom, "secret"),
				},
			},
		})
	}
	if lark, ok := singleBlock(appIM, "lark"); ok {
		settings = append(settings, &v1pb.AppIMSetting_IMSetting{
			Type: v1pb.WebhookType_LARK,
			Payload: &v1pb.AppIMSetting_IMSetting_Lark{
				Lark: &v1pb.AppIMSetting_Lark{
					AppId:     ctyString(lark, "app_id"),
					AppSecret: ctyString(lark, "app_secret"),
				},
			},
		})
	}
	if dingtalk, ok := singleBlock(appIM, "dingtalk"); ok {
		settings = append(settings, &v1pb.AppIMSetting_IMSetting{
			Type: v1pb.WebhookType_DINGTALK,
			Payload: &v1pb.AppIMSetting_IMSetting_Dingtalk{
				Dingtalk: &v1pb.AppIMSetting_DingTalk{
					ClientId:     ctyString(dingtalk, "client_id"),
					ClientSecret: ctyString(dingtalk, "client_secret"),
					RobotCode:    ctyString(dingtalk, "robot_code"),
				},
			},
		})
	}
	if teams, ok := singleBlock(appIM, "teams"); ok {
		settings = append(settings, &v1pb.AppIMSetting_IMSetting{
			Type: v1pb.WebhookType_TEAMS,
			Payload: &v1pb.AppIMSetting_IMSetting_Teams{
				Teams: &v1pb.AppIMSetting_Teams{
					TenantId:     ctyString(teams, "tenant_id"),
					ClientId:     ctyString(teams, "client_id"),
					ClientSecret: ctyString(teams, "client_secret"),
				},
			},
		})
	}

	return &v1pb.AppIMSetting{Settings: settings}, nil
}

// singleBlock returns the first (and only) element of a singleton TypeList
// block. ok is false when the block is absent, null, unknown, or empty.
func singleBlock(parent cty.Value, name string) (cty.Value, bool) {
	v := parent.GetAttr(name)
	if v.IsNull() || !v.IsKnown() || v.LengthInt() == 0 {
		return cty.NilVal, false
	}
	return v.AsValueSlice()[0], true
}

// ctyString extracts a string attribute, returning "" for null/unknown.
func ctyString(parent cty.Value, name string) string {
	v := parent.GetAttr(name)
	if v.IsNull() || !v.IsKnown() {
		return ""
	}
	return v.AsString()
}

// flattenAppIMSetting builds the state representation of an AppIMSetting.
// The server returns empty payload structs for every IM type, so the field
// values here are always empty strings — block presence is the only signal.
// On the resource side, the SDK strips WriteOnly fields from state anyway.
func flattenAppIMSetting(setting *v1pb.AppIMSetting) []interface{} {
	if setting == nil {
		return nil
	}

	block := map[string]interface{}{
		"slack":    []interface{}{},
		"feishu":   []interface{}{},
		"wecom":    []interface{}{},
		"lark":     []interface{}{},
		"dingtalk": []interface{}{},
		"teams":    []interface{}{},
	}

	for _, im := range setting.GetSettings() {
		switch im.GetType() {
		case v1pb.WebhookType_SLACK:
			block["slack"] = []interface{}{map[string]interface{}{"token": ""}}
		case v1pb.WebhookType_FEISHU:
			block["feishu"] = []interface{}{map[string]interface{}{"app_id": "", "app_secret": ""}}
		case v1pb.WebhookType_WECOM:
			block["wecom"] = []interface{}{map[string]interface{}{"corp_id": "", "agent_id": "", "secret": ""}}
		case v1pb.WebhookType_LARK:
			block["lark"] = []interface{}{map[string]interface{}{"app_id": "", "app_secret": ""}}
		case v1pb.WebhookType_DINGTALK:
			block["dingtalk"] = []interface{}{map[string]interface{}{"client_id": "", "client_secret": "", "robot_code": ""}}
		case v1pb.WebhookType_TEAMS:
			block["teams"] = []interface{}{map[string]interface{}{"tenant_id": "", "client_id": "", "client_secret": ""}}
		}
	}

	return []interface{}{block}
}
