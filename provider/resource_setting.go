package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/type/expr"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func resourceSetting() *schema.Resource {
	return &schema.Resource{
		Description:   "The setting resource.",
		CreateContext: resourceSettingUpsert,
		ReadContext:   resourceSettingRead,
		UpdateContext: resourceSettingUpsert,
		DeleteContext: resourceSettingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(api.SettingWorkspaceApproval),
					string(api.SettingWorkspaceExternalApproval),
					string(api.SettingWorkspaceProfile),
					string(api.SettingDataClassification),
				}, false),
			},
			"approval_flow":           getWorkspaceApprovalSetting(false),
			"external_approval_nodes": getExternalApprovalSetting(false),
			"workspace_profile":       getWorkspaceProfileSetting(false),
			"classification":          getClassificationSetting(false)},
	}
}

func resourceSettingUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	var diags diag.Diagnostics

	name := api.SettingName(d.Get("name").(string))
	settingName := fmt.Sprintf("%s%s", internal.SettingNamePrefix, string(name))

	setting := &v1pb.Setting{
		Name: settingName,
	}
	updateMasks := []string{}

	switch name {
	case api.SettingWorkspaceApproval:
		workspaceApproval, err := convertToV1ApprovalSetting(d)
		if err != nil {
			return diag.FromErr(err)
		}
		setting.Value = &v1pb.Value{
			Value: &v1pb.Value_WorkspaceApprovalSettingValue{
				WorkspaceApprovalSettingValue: workspaceApproval,
			},
		}
	case api.SettingWorkspaceExternalApproval:
		externalApproval, err := convertToV1ExternalNodesSetting(d)
		if err != nil {
			return diag.FromErr(err)
		}
		setting.Value = &v1pb.Value{
			Value: &v1pb.Value_ExternalApprovalSettingValue{
				ExternalApprovalSettingValue: externalApproval,
			},
		}
	case api.SettingWorkspaceProfile:
		workspaceProfile, updatePathes, err := convertToV1WorkspaceProfileSetting(d)
		if err != nil {
			return diag.FromErr(err)
		}
		setting.Value = &v1pb.Value{
			Value: &v1pb.Value_WorkspaceProfileSettingValue{
				WorkspaceProfileSettingValue: workspaceProfile,
			},
		}
		updateMasks = updatePathes
	case api.SettingDataClassification:
		classificationSetting, err := convertToV1ClassificationSetting(d)
		if err != nil {
			return diag.FromErr(err)
		}
		setting.Value = &v1pb.Value{
			Value: &v1pb.Value_DataClassificationSettingValue{
				DataClassificationSettingValue: classificationSetting,
			},
		}
	default:
		return diag.FromErr(errors.Errorf("Unsupport setting: %v", name))
	}

	updatedSetting, err := c.UpsertSetting(ctx, setting, updateMasks)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(updatedSetting.Name)

	diag := resourceSettingRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func convertToV1WorkspaceProfileSetting(d *schema.ResourceData) (*v1pb.WorkspaceProfileSetting, []string, error) {
	rawList, ok := d.Get("workspace_profile").([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, nil, errors.Errorf("invalid workspace_profile")
	}

	updateMasks := []string{}
	raw := rawList[0].(map[string]interface{})

	workspacePrfile := &v1pb.WorkspaceProfileSetting{}

	if externalURL, ok := raw["external_url"]; ok {
		workspacePrfile.ExternalUrl = externalURL.(string)
		updateMasks = append(updateMasks, "value.workspace_profile_setting_value.external_url")
	}
	if disallowSignup, ok := raw["disallow_signup"]; ok {
		workspacePrfile.DisallowSignup = disallowSignup.(bool)
		updateMasks = append(updateMasks, "value.workspace_profile_setting_value.disallow_signup")
	}
	if disallowPasswordSignin, ok := raw["disallow_password_signin"]; ok {
		workspacePrfile.DisallowPasswordSignin = disallowPasswordSignin.(bool)
		updateMasks = append(updateMasks, "value.workspace_profile_setting_value.disallow_password_signin")
	}
	if domains, ok := raw["domains"]; ok {
		if enforceIdentityDomain, ok := raw["enforce_identity_domain"]; ok {
			workspacePrfile.EnforceIdentityDomain = enforceIdentityDomain.(bool)
			updateMasks = append(updateMasks, "value.workspace_profile_setting_value.enforce_identity_domain")
		}
		for _, domain := range domains.([]interface{}) {
			workspacePrfile.Domains = append(workspacePrfile.Domains, domain.(string))
		}
		updateMasks = append(updateMasks, "value.workspace_profile_setting_value.domains")
	} else if _, ok := raw["enforce_identity_domain"]; ok {
		return nil, nil, errors.Errorf("enforce_identity_domain must works with domains")
	}

	return workspacePrfile, updateMasks, nil
}

func convertToV1ClassificationSetting(d *schema.ResourceData) (*v1pb.DataClassificationSetting, error) {
	rawList, ok := d.Get("classification").([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.Errorf("invalid classification")
	}

	raw := rawList[0].(map[string]interface{})

	dataClassificationConfig := &v1pb.DataClassificationSetting_DataClassificationConfig{
		Id:                       raw["id"].(string),
		Title:                    raw["title"].(string),
		ClassificationFromConfig: raw["classification_from_config"].(bool),
		Levels:                   []*v1pb.DataClassificationSetting_DataClassificationConfig_Level{},
		Classification:           map[string]*v1pb.DataClassificationSetting_DataClassificationConfig_DataClassification{},
	}
	if dataClassificationConfig.Id == "" {
		return nil, errors.Errorf("id is required for classification config")
	}

	rawLevels := raw["levels"].([]interface{})
	for _, level := range rawLevels {
		rawLevel := level.(map[string]interface{})
		classificationLevel := &v1pb.DataClassificationSetting_DataClassificationConfig_Level{
			Id:          rawLevel["id"].(string),
			Title:       rawLevel["title"].(string),
			Description: rawLevel["description"].(string),
		}
		if classificationLevel.Id == "" {
			return nil, errors.Errorf("classification level id is required")
		}
		if classificationLevel.Title == "" {
			return nil, errors.Errorf("classification level title is required")
		}
		dataClassificationConfig.Levels = append(dataClassificationConfig.Levels, classificationLevel)
	}

	rawClassificationss := raw["classifications"].([]interface{})
	for _, classification := range rawClassificationss {
		rawClassification := classification.(map[string]interface{})
		classificationData := &v1pb.DataClassificationSetting_DataClassificationConfig_DataClassification{
			Id:          rawClassification["id"].(string),
			Title:       rawClassification["title"].(string),
			Description: rawClassification["description"].(string),
		}
		if classificationData.Id == "" {
			return nil, errors.Errorf("classification id is required")
		}
		if classificationData.Title == "" {
			return nil, errors.Errorf("classification title is required")
		}
		levelID, ok := rawClassification["level"].(string)
		if ok {
			classificationData.LevelId = &levelID
		}
		dataClassificationConfig.Classification[classificationData.Id] = classificationData
	}

	return &v1pb.DataClassificationSetting{
		Configs: []*v1pb.DataClassificationSetting_DataClassificationConfig{
			dataClassificationConfig,
		},
	}, nil
}

func convertToV1ExternalNodesSetting(d *schema.ResourceData) (*v1pb.ExternalApprovalSetting, error) {
	rawList, ok := d.Get("external_approval_nodes").([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.Errorf("invalid external_approval_nodes")
	}

	raw := rawList[0].(map[string]interface{})
	nodes := raw["nodes"].([]interface{})
	externalApprovalSetting := &v1pb.ExternalApprovalSetting{}

	for _, node := range nodes {
		rawNode := node.(map[string]interface{})
		externalApprovalSetting.Nodes = append(externalApprovalSetting.Nodes, &v1pb.ExternalApprovalSetting_Node{
			Id:       rawNode["id"].(string),
			Title:    rawNode["title"].(string),
			Endpoint: rawNode["endpoint"].(string),
		})
	}
	return externalApprovalSetting, nil
}

func convertToV1ApprovalSetting(d *schema.ResourceData) (*v1pb.WorkspaceApprovalSetting, error) {
	rawList, ok := d.Get("approval_flow").([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.Errorf("invalid approval_flow")
	}

	raw := rawList[0].(map[string]interface{})
	rules := raw["rules"].([]interface{})
	workspaceApprovalSetting := &v1pb.WorkspaceApprovalSetting{}
	for _, rule := range rules {
		rawRule := rule.(map[string]interface{})

		// build condition expression.
		conditionList, ok := rawRule["conditions"].([]interface{})
		if !ok {
			return nil, errors.Errorf("invalid conditions")
		}
		buildCondition := []string{}
		for _, condition := range conditionList {
			rawCondition := condition.(map[string]interface{})
			rawLevel := rawCondition["level"].(string)
			buildCondition = append(buildCondition, fmt.Sprintf(`source == "%s" && level == %d`, rawCondition["source"].(string), api.RiskLevel(rawLevel).Int()))
		}
		expression := strings.Join(buildCondition, " || ")

		flowList, ok := rawRule["flow"].([]interface{})
		if !ok || len(flowList) != 1 {
			return nil, errors.Errorf("invalid flow")
		}
		rawFlow := flowList[0].(map[string]interface{})
		creator := rawFlow["creator"].(string)
		if !strings.HasPrefix(creator, "users/") {
			return nil, errors.Errorf("creator should in users/{email} format")
		}
		approvalRule := &v1pb.WorkspaceApprovalSetting_Rule{
			Template: &v1pb.ApprovalTemplate{
				Title:       rawFlow["title"].(string),
				Description: rawFlow["description"].(string),
				Creator:     rawFlow["creator"].(string),
				Flow:        &v1pb.ApprovalFlow{},
			},
			Condition: &expr.Expr{
				Expression: expression,
			},
		}

		stepList, ok := rawFlow["steps"].([]interface{})
		if !ok {
			return nil, errors.Errorf("invalid steps")
		}

		for _, step := range stepList {
			rawStep := step.(map[string]interface{})
			stepType := api.ApprovalNodeType(rawStep["type"].(string))
			node := rawStep["node"].(string)

			approvalNode := &v1pb.ApprovalNode{
				Type: v1pb.ApprovalNode_ANY_IN_GROUP,
			}
			switch stepType {
			case api.ApprovalNodeTypeRole:
				if !strings.HasPrefix(node, "roles/") {
					return nil, errors.Errorf("invalid role name: %v, role name should in roles/{role} format", node)
				}
				approvalNode.Payload = &v1pb.ApprovalNode_Role{
					Role: node,
				}
			case api.ApprovalNodeTypeGroup:
				group, ok := v1pb.ApprovalNode_GroupValue_value[node]
				if !ok {
					return nil, errors.Errorf(
						"invalid group: %v, group should be one of: %s, %s, %s, %s",
						node,
						v1pb.ApprovalNode_WORKSPACE_OWNER.String(),
						v1pb.ApprovalNode_WORKSPACE_DBA.String(),
						v1pb.ApprovalNode_PROJECT_OWNER.String(),
						v1pb.ApprovalNode_PROJECT_MEMBER.String(),
					)
				}
				approvalNode.Payload = &v1pb.ApprovalNode_GroupValue_{
					GroupValue: v1pb.ApprovalNode_GroupValue(group),
				}
			case api.ApprovalNodeTypeExternalNodeID:
				approvalNode.Payload = &v1pb.ApprovalNode_ExternalNodeId{
					ExternalNodeId: node,
				}
			}

			approvalStep := &v1pb.ApprovalStep{
				Type:  v1pb.ApprovalStep_ANY,
				Nodes: []*v1pb.ApprovalNode{approvalNode},
			}

			approvalRule.Template.Flow.Steps = append(approvalRule.Template.Flow.Steps, approvalStep)
		}

		workspaceApprovalSetting.Rules = append(workspaceApprovalSetting.Rules, approvalRule)
	}

	return workspaceApprovalSetting, nil
}

func resourceSettingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	settingName := d.Id()
	setting, err := c.GetSetting(ctx, settingName)
	if err != nil {
		return diag.FromErr(err)
	}

	return setSettingMessage(ctx, d, c, setting)
}

func resourceSettingDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}
