package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
)

func getMCPSetting(computed bool) *schema.Schema {
	capability := &schema.Schema{
		Type:        schema.TypeString,
		Required:    !computed,
		Computed:    computed,
		Description: "The workspace MCP capability ceiling: DISABLED, READ_ONLY, or READ_WRITE. MCP sessions remain bounded by the connected user's permissions.",
	}
	if !computed {
		capability.ValidateFunc = validation.StringInSlice([]string{
			v1pb.MCPSetting_DISABLED.String(),
			v1pb.MCPSetting_READ_ONLY.String(),
			v1pb.MCPSetting_READ_WRITE.String(),
		}, false)
	}
	description := "The MCP workspace setting."
	if !computed {
		description += " Removing the Terraform resource disables MCP for the workspace."
	}
	return singletonBlock(computed, description, map[string]*schema.Schema{"capability": capability})
}

func convertToV1MCPSetting(d *schema.ResourceData) (*v1pb.MCPSetting, error) {
	rawConfig := d.GetRawConfig()
	if rawConfig.IsNull() || !rawConfig.IsKnown() {
		return nil, errors.New("mcp block is required when name is settings/MCP")
	}
	mcp, ok := singleBlock(rawConfig, "mcp")
	if !ok {
		return nil, errors.New("mcp block is required when name is settings/MCP")
	}
	capability := v1pb.MCPSetting_Capability(v1pb.MCPSetting_Capability_value[ctyString(mcp, "capability")])
	switch capability {
	case v1pb.MCPSetting_DISABLED, v1pb.MCPSetting_READ_ONLY, v1pb.MCPSetting_READ_WRITE:
		return &v1pb.MCPSetting{Capability: capability}, nil
	default:
		return nil, errors.New("mcp.capability must be DISABLED, READ_ONLY, or READ_WRITE")
	}
}

func flattenMCPSetting(setting *v1pb.MCPSetting) []interface{} {
	return []interface{}{map[string]interface{}{"capability": setting.GetCapability().String()}}
}
