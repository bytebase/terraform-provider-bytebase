package provider

import (
	"strings"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
)

func getExportSetting(computed bool) *schema.Schema {
	fields := map[string]*schema.Schema{
		"tenant_id": {
			Type: schema.TypeString, Optional: !computed, Computed: true,
			Description: "The Microsoft Entra tenant ID. Omit to preserve the current value.",
		},
		"client_id": {
			Type: schema.TypeString, Optional: !computed, Computed: true,
			Description: "The Microsoft Entra application ID. Omit to preserve the current value.",
		},
		"client_secret_set": {
			Type: schema.TypeBool, Computed: true,
			Description: "Whether the server stores a SharePoint application secret. The secret is never returned.",
		},
	}
	if !computed {
		for _, name := range []string{"tenant_id", "client_id"} {
			fields[name].StateFunc = func(v interface{}) string { return strings.TrimSpace(v.(string)) }
		}
		fields["client_secret"] = &schema.Schema{
			Type: schema.TypeString, Optional: true, Sensitive: true, WriteOnly: true,
			Description: "The application secret. Write-only: never stored in Terraform state. Set client_secret_version to a positive integer and increment it whenever this value changes. Omit both to preserve the stored secret.",
		}
		fields["client_secret_version"] = &schema.Schema{
			Type: schema.TypeInt, Optional: true, Default: 0,
			ValidateFunc: validation.IntAtLeast(0),
			Description:  "Local rotation marker, not a server version. Set to at least 1 with client_secret and increment to rotate it. Unrelated updates never resend an unchanged secret.",
		}
	}
	description := "The workspace SharePoint export application. Requires the SharePoint export license feature and a self-hosted server."
	if !computed {
		description += " Omitted fields preserve server values. Configure an empty export {} block to clear the application and its secret. Destroying the resource preserves these shared credentials."
	}
	return singletonBlock(computed, description, map[string]*schema.Schema{
		"sharepoint": singletonBlock(computed, "The Microsoft Entra application used to upload exports to SharePoint.", fields),
	})
}

func exportSharePointConfig(config cty.Value) (cty.Value, error) {
	if !config.IsKnown() {
		return cty.DynamicVal, nil
	}
	if config.IsNull() {
		return cty.NilVal, errors.New("export block is required when name is settings/EXPORT")
	}
	for _, name := range []string{"export", "sharepoint"} {
		block := config.GetAttr(name)
		if !block.IsKnown() {
			return cty.DynamicVal, nil
		}
		if block.IsNull() || block.LengthInt() == 0 {
			if name == "sharepoint" {
				return cty.NilVal, nil
			}
			return cty.NilVal, errors.New("export block is required when name is settings/EXPORT")
		}
		config = block.AsValueSlice()[0]
		if !config.IsKnown() {
			return cty.DynamicVal, nil
		}
	}
	return config, nil
}

func validateExportSettingConfig(config cty.Value) error {
	block, err := exportSharePointConfig(config)
	if err != nil || !block.IsKnown() || block.IsNull() {
		return err
	}
	for _, name := range []string{"tenant_id", "client_id"} {
		value := block.GetAttr(name)
		if value.IsKnown() && !value.IsNull() && strings.TrimSpace(value.AsString()) == "" {
			return errors.New("SharePoint IDs must not be empty; omit a field to preserve it or configure export {} to clear credentials")
		}
	}
	secret := block.GetAttr("client_secret")
	version := block.GetAttr("client_secret_version")
	if !secret.IsKnown() || !version.IsKnown() {
		return nil
	}
	var marker int64
	if !version.IsNull() {
		marker, _ = version.AsBigFloat().Int64()
	}
	if marker < 0 {
		return errors.New("client_secret_version must not be negative")
	}
	if !secret.IsNull() {
		if secret.AsString() == "" {
			return errors.New("client_secret must not be empty; omit it to preserve the stored secret")
		}
		if marker < 1 {
			return errors.New("client_secret_version must be at least 1 when client_secret is configured")
		}
	} else if marker > 0 {
		return errors.New("client_secret is required when client_secret_version is configured")
	}
	return nil
}

func convertToV1ExportSetting(d *schema.ResourceData) (*v1pb.ExportSetting, []string, error) {
	config := d.GetRawConfig()
	if err := validateExportSettingConfig(config); err != nil {
		return nil, nil, err
	}
	block, err := exportSharePointConfig(config)
	if err != nil {
		return nil, nil, err
	}
	if !block.IsKnown() {
		return nil, nil, errors.New("export.sharepoint must be known before applying")
	}
	if block.IsNull() {
		return &v1pb.ExportSetting{}, []string{"value.export.sharepoint.tenant_id", "value.export.sharepoint.client_id"}, nil
	}
	sharepoint := &v1pb.SharePointSetting{}
	masks := []string{}
	for _, name := range []string{"tenant_id", "client_id"} {
		value := block.GetAttr(name)
		if value.IsNull() || !value.IsKnown() {
			continue
		}
		switch name {
		case "tenant_id":
			sharepoint.TenantId = strings.TrimSpace(value.AsString())
		case "client_id":
			sharepoint.ClientId = strings.TrimSpace(value.AsString())
		default:
			return nil, nil, errors.Errorf("unsupported SharePoint field %q", name)
		}
		masks = append(masks, "value.export.sharepoint."+name)
	}
	secret := block.GetAttr("client_secret")
	if !secret.IsNull() && secret.IsKnown() && (d.Id() == "" || d.HasChange("export.0.sharepoint.0.client_secret_version")) {
		sharepoint.ClientSecret = secret.AsString()
		masks = append(masks, "value.export.sharepoint.client_secret")
	}
	return &v1pb.ExportSetting{Sharepoint: sharepoint}, masks, nil
}

func flattenExportSetting(setting *v1pb.ExportSetting, computed bool, version int, keepBlock bool) []interface{} {
	sharepoint := setting.GetSharepoint()
	blocks := []interface{}{}
	if sharepoint != nil || !computed && keepBlock {
		fields := map[string]interface{}{
			"tenant_id": sharepoint.GetTenantId(), "client_id": sharepoint.GetClientId(),
			"client_secret_set": sharepoint.GetClientSecretSet(),
		}
		if !computed {
			fields["client_secret_version"] = version
		}
		blocks = append(blocks, fields)
	}
	return []interface{}{map[string]interface{}{"sharepoint": blocks}}
}
