package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
)

func getAISetting(computed bool) *schema.Schema {
	fields := map[string]*schema.Schema{
		"enabled": {
			Type:        schema.TypeBool,
			Description: "Whether AI features are enabled. Cannot be changed in SaaS mode.",
		},
		"provider": {
			Type: schema.TypeString,
			ValidateFunc: validation.StringInSlice([]string{
				v1pb.AISetting_PROVIDER_UNSPECIFIED.String(),
				v1pb.AISetting_OPEN_AI.String(),
				v1pb.AISetting_CLAUDE.String(),
				v1pb.AISetting_GEMINI.String(),
				v1pb.AISetting_AZURE_OPENAI.String(),
			}, false),
			Description: "AI provider. Cannot be changed in SaaS mode.",
		},
		"endpoint": {
			Type:        schema.TypeString,
			Description: "The AI API endpoint. Required when AI is enabled. Cannot be changed in SaaS mode.",
		},
		"api_key": {
			Type:        schema.TypeString,
			Sensitive:   true,
			Description: "The AI API key. The server redacts it on reads, so the configured value is retained in Terraform state. Omit to preserve the existing key. Cannot be changed in SaaS mode.",
		},
		"model": {
			Type:        schema.TypeString,
			Description: "The AI model. Required when AI is enabled. Cannot be changed in SaaS mode.",
		},
	}
	for _, field := range fields {
		field.Optional = !computed
		field.Computed = true
	}
	if computed {
		fields["api_key"].Description = "The server redacts the AI API key, so this is always empty on reads."
		fields["provider"].ValidateFunc = nil
	}
	fields["version"] = &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The API version reported by Bytebase. Read-only because Bytebase 3.24 does not support updating this field.",
	}
	description := "The AI workspace setting."
	if !computed {
		description += " Omitted fields are not changed. AI provider settings are managed by the server in SaaS mode. Removing the Terraform resource preserves the server's AI settings; explicitly set enabled to false and apply before removal to disable AI."
	}
	return singletonBlock(computed, description, fields)
}

func convertToV1AISetting(d *schema.ResourceData) (*v1pb.AISetting, []string, error) {
	rawConfig := d.GetRawConfig()
	if rawConfig.IsNull() || !rawConfig.IsKnown() {
		return nil, nil, errors.New("ai block is required when name is settings/AI")
	}
	ai, ok := singleBlock(rawConfig, "ai")
	if !ok {
		return nil, nil, errors.New("ai block is required when name is settings/AI")
	}
	setting := &v1pb.AISetting{}
	masks := []string{}
	for _, field := range []string{"enabled", "provider", "endpoint", "api_key", "model"} {
		value := ai.GetAttr(field)
		if value.IsNull() || !value.IsKnown() {
			continue
		}
		// A redacted key cannot detect remote rotations. Do not restore an old
		// state value while updating another field.
		if field == "api_key" && d.Id() != "" && !d.HasChange("ai.0.api_key") {
			continue
		}
		switch field {
		case "enabled":
			setting.Enabled = value.True()
		case "provider":
			setting.Provider = v1pb.AISetting_Provider(v1pb.AISetting_Provider_value[value.AsString()])
		case "endpoint":
			setting.Endpoint = value.AsString()
		case "api_key":
			setting.ApiKey = value.AsString()
		case "model":
			setting.Model = value.AsString()
		default:
			return nil, nil, errors.Errorf("unsupported AI setting field %q", field)
		}
		masks = append(masks, "value.ai."+field)
	}
	return setting, masks, nil
}

func flattenAISetting(setting *v1pb.AISetting, previousKey string) []interface{} {
	apiKey := setting.GetApiKey()
	if apiKey == "" {
		apiKey = previousKey
	}
	return []interface{}{map[string]interface{}{
		"enabled":  setting.GetEnabled(),
		"provider": setting.GetProvider().String(),
		"endpoint": setting.GetEndpoint(),
		"api_key":  apiKey,
		"model":    setting.GetModel(),
		"version":  setting.GetVersion(),
	}}
}
