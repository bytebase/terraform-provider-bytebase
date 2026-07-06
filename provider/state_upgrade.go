package provider

import (
	"context"
	"strconv"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func resourceEnvironmentV0Type() cty.Type {
	resourceSchema := resourceEnvironmentSchema()
	resourceSchema["color"] = legacyColorStringSchema("The environment color.")
	return schemaMapType(resourceSchema)
}

func resourceProjectV0Type() cty.Type {
	resourceSchema := resourceProjectSchema()
	issueLabels := resourceSchema["issue_labels"].Elem.(*schema.Resource)
	issueLabels.Schema["color"] = legacyColorStringSchema("The color code for the label (e.g., hex color).")
	return schemaMapType(resourceSchema)
}

func resourceSettingV0Type() cty.Type {
	resourceSchema := resourceSettingSchema()
	resourceSchema["workspace_profile"] = getWorkspaceProfileSettingV0(false)
	resourceSchema["environment_setting"] = getEnvironmentSettingV0(false)
	return schemaMapType(resourceSchema)
}

func schemaMapType(resourceSchema map[string]*schema.Schema) cty.Type {
	return (&schema.Resource{Schema: resourceSchema}).CoreConfigSchema().ImpliedType()
}

func legacyColorStringSchema(description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: description,
	}
}

func getWorkspaceProfileSettingV0(computed bool) *schema.Schema {
	result := getWorkspaceProfileSetting(computed)
	workspaceProfile := result.Elem.(*schema.Resource)
	delete(workspaceProfile.Schema, "maximum_request_expiration_in_seconds")
	workspaceProfile.Schema["maximum_role_expiration_in_seconds"] = &schema.Schema{
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
		Description: "The max duration in seconds for role expired. If the value is less than or equal to 0, we will remove the setting. AKA no limit.",
	}

	announcement := workspaceProfile.Schema["announcement"].Elem.(*schema.Resource)
	delete(announcement.Schema, "theme")
	announcement.Schema["level"] = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The alert level of announcement",
	}
	return result
}

func getEnvironmentSettingV0(computed bool) *schema.Schema {
	result := getEnvironmentSetting(computed)
	environmentSetting := result.Elem.(*schema.Resource)
	environment := environmentSetting.Schema["environment"].Elem.(*schema.Resource)
	environment.Schema["color"] = legacyColorStringSchema("The environment color.")
	return result
}

func upgradeEnvironmentColorState(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	if rawState == nil {
		return rawState, nil
	}
	if err := upgradeColorAttribute(rawState, "color"); err != nil {
		return nil, err
	}
	return rawState, nil
}

func upgradeProjectColorState(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	if rawState == nil {
		return rawState, nil
	}
	labels, ok := rawState["issue_labels"].([]interface{})
	if !ok {
		return rawState, nil
	}
	for _, rawLabel := range labels {
		label, ok := rawLabel.(map[string]interface{})
		if !ok {
			continue
		}
		if err := upgradeColorAttribute(label, "color"); err != nil {
			return nil, err
		}
	}
	return rawState, nil
}

func upgradeSettingColorState(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	if rawState == nil {
		return rawState, nil
	}
	if err := upgradeEnvironmentSettingColorState(rawState); err != nil {
		return nil, err
	}
	upgradeWorkspaceProfileState(rawState)
	return rawState, nil
}

func upgradeEnvironmentSettingColorState(rawState map[string]interface{}) error {
	settings, ok := rawState["environment_setting"].([]interface{})
	if !ok {
		return nil
	}
	for _, rawSetting := range settings {
		setting, ok := rawSetting.(map[string]interface{})
		if !ok {
			continue
		}
		environments, ok := setting["environment"].([]interface{})
		if !ok {
			continue
		}
		for _, rawEnvironment := range environments {
			environment, ok := rawEnvironment.(map[string]interface{})
			if !ok {
				continue
			}
			if err := upgradeColorAttribute(environment, "color"); err != nil {
				return err
			}
		}
	}
	return nil
}

func upgradeWorkspaceProfileState(rawState map[string]interface{}) {
	profiles, ok := rawState["workspace_profile"].([]interface{})
	if !ok {
		return
	}
	for _, rawProfile := range profiles {
		profile, ok := rawProfile.(map[string]interface{})
		if !ok {
			continue
		}
		if oldValue, ok := profile["maximum_role_expiration_in_seconds"]; ok {
			if _, exists := profile["maximum_request_expiration_in_seconds"]; !exists {
				profile["maximum_request_expiration_in_seconds"] = oldValue
			}
			delete(profile, "maximum_role_expiration_in_seconds")
		}
		announcements, ok := profile["announcement"].([]interface{})
		if !ok {
			continue
		}
		for _, rawAnnouncement := range announcements {
			announcement, ok := rawAnnouncement.(map[string]interface{})
			if !ok {
				continue
			}
			delete(announcement, "level")
		}
	}
}

func upgradeColorAttribute(raw map[string]interface{}, key string) error {
	value, exists := raw[key]
	if !exists {
		return nil
	}
	switch color := value.(type) {
	case nil:
		delete(raw, key)
	case string:
		block, err := upgradeLegacyColorString(color)
		if err != nil {
			return err
		}
		raw[key] = block
	default:
		return nil
	}
	return nil
}

func upgradeLegacyColorString(color string) ([]interface{}, error) {
	color = strings.TrimSpace(color)
	if color == "" {
		return []interface{}{}, nil
	}
	color = strings.TrimPrefix(color, "#")
	if len(color) != 6 {
		return nil, errors.Errorf("invalid legacy color %q, want #RRGGBB", color)
	}
	red, err := parseLegacyColorChannel(color[0:2])
	if err != nil {
		return nil, err
	}
	green, err := parseLegacyColorChannel(color[2:4])
	if err != nil {
		return nil, err
	}
	blue, err := parseLegacyColorChannel(color[4:6])
	if err != nil {
		return nil, err
	}
	return []interface{}{
		map[string]interface{}{
			"red":   red,
			"green": green,
			"blue":  blue,
		},
	}, nil
}

func parseLegacyColorChannel(raw string) (float64, error) {
	value, err := strconv.ParseUint(raw, 16, 8)
	if err != nil {
		return 0, errors.Errorf("invalid legacy color channel %q", raw)
	}
	return normalizeColorChannel(float32(value) / 255), nil
}
