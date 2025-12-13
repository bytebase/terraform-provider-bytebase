package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/type/expr"
	"google.golang.org/protobuf/types/known/durationpb"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"

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
				Type:        schema.TypeString,
				Required:    true,
				Description: `The setting name in settings/{name} format. The name support "WORKSPACE_APPROVAL", "WORKSPACE_PROFILE", "DATA_CLASSIFICATION", "SEMANTIC_TYPES", "ENVIRONMENT". Check the proto https://github.com/bytebase/bytebase/blob/main/proto/v1/v1/setting_service.proto#L109 for details`,
				ValidateDiagFunc: internal.ResourceNameValidation(
					fmt.Sprintf("^%s%s$", internal.SettingNamePrefix, v1pb.Setting_WORKSPACE_APPROVAL.String()),
					fmt.Sprintf("^%s%s$", internal.SettingNamePrefix, v1pb.Setting_WORKSPACE_PROFILE.String()),
					fmt.Sprintf("^%s%s$", internal.SettingNamePrefix, v1pb.Setting_DATA_CLASSIFICATION.String()),
					fmt.Sprintf("^%s%s$", internal.SettingNamePrefix, v1pb.Setting_SEMANTIC_TYPES.String()),
					fmt.Sprintf("^%s%s$", internal.SettingNamePrefix, v1pb.Setting_ENVIRONMENT.String()),
				),
			},
			"approval_flow":       getWorkspaceApprovalSetting(false),
			"workspace_profile":   getWorkspaceProfileSetting(false),
			"classification":      getClassificationSetting(false),
			"semantic_types":      getSemanticTypesSetting(false),
			"environment_setting": getEnvironmentSetting(false),
		},
	}
}

func resourceSettingUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	var diags diag.Diagnostics

	settingName := d.Get("name").(string)
	name, err := internal.GetSettingName(settingName)
	if err != nil {
		return diag.FromErr(err)
	}

	setting := &v1pb.Setting{
		Name: settingName,
	}
	updateMasks := []string{}

	switch name {
	case v1pb.Setting_WORKSPACE_APPROVAL:
		workspaceApproval, err := convertToV1ApprovalSetting(d)
		if err != nil {
			return diag.FromErr(err)
		}
		setting.Value = &v1pb.SettingValue{
			Value: &v1pb.SettingValue_WorkspaceApproval{
				WorkspaceApproval: workspaceApproval,
			},
		}
	case v1pb.Setting_WORKSPACE_PROFILE:
		workspaceProfile, updatePathes, err := convertToV1WorkspaceProfileSetting(d)
		if err != nil {
			return diag.FromErr(err)
		}
		setting.Value = &v1pb.SettingValue{
			Value: &v1pb.SettingValue_WorkspaceProfile{
				WorkspaceProfile: workspaceProfile,
			},
		}
		updateMasks = updatePathes
	case v1pb.Setting_DATA_CLASSIFICATION:
		classificationSetting, err := convertToV1ClassificationSetting(d)
		if err != nil {
			return diag.FromErr(err)
		}
		setting.Value = &v1pb.SettingValue{
			Value: &v1pb.SettingValue_DataClassification{
				DataClassification: classificationSetting,
			},
		}
	case v1pb.Setting_SEMANTIC_TYPES:
		semanticTypeSetting, err := convertToV1SemanticTypeSetting(d)
		if err != nil {
			return diag.FromErr(err)
		}
		setting.Value = &v1pb.SettingValue{
			Value: &v1pb.SettingValue_SemanticType{
				SemanticType: semanticTypeSetting,
			},
		}
	case v1pb.Setting_ENVIRONMENT:
		environmentSetting, err := convertToV1EnvironmentSetting(d)
		if err != nil {
			return diag.FromErr(err)
		}
		setting.Value = &v1pb.SettingValue{
			Value: &v1pb.SettingValue_Environment{
				Environment: environmentSetting,
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

	rawConfig := d.GetRawConfig().GetAttr("workspace_profile")
	if rawConfig.IsNull() {
		return nil, nil, errors.Errorf("invalid workspace_profile")
	}
	if rawConfig.IsKnown() && rawConfig.LengthInt() == 0 {
		return nil, nil, errors.Errorf("invalid workspace_profile")
	}
	workspaceRawConfig := rawConfig.AsValueSlice()[0]

	updateMasks := []string{}
	raw := rawList[0].(map[string]interface{})

	workspacePrfile := &v1pb.WorkspaceProfileSetting{}

	if config := workspaceRawConfig.GetAttr("external_url"); !config.IsNull() {
		workspacePrfile.ExternalUrl = raw["external_url"].(string)
		updateMasks = append(updateMasks, "value.workspace_profile.external_url")
	}
	if config := workspaceRawConfig.GetAttr("disallow_signup"); !config.IsNull() {
		workspacePrfile.DisallowSignup = raw["disallow_signup"].(bool)
		updateMasks = append(updateMasks, "value.workspace_profile.disallow_signup")
	}
	if config := workspaceRawConfig.GetAttr("disallow_password_signin"); !config.IsNull() {
		workspacePrfile.DisallowPasswordSignin = raw["disallow_password_signin"].(bool)
		updateMasks = append(updateMasks, "value.workspace_profile.disallow_password_signin")
	}
	if config := workspaceRawConfig.GetAttr("domains"); !config.IsNull() {
		if domains, ok := raw["domains"]; ok {
			if enforceIdentityDomain, ok := raw["enforce_identity_domain"]; ok {
				workspacePrfile.EnforceIdentityDomain = enforceIdentityDomain.(bool)
				updateMasks = append(updateMasks, "value.workspace_profile.enforce_identity_domain")
			}
			for _, domain := range domains.([]interface{}) {
				workspacePrfile.Domains = append(workspacePrfile.Domains, domain.(string))
			}
			updateMasks = append(updateMasks, "value.workspace_profile.domains")
		} else if _, ok := raw["enforce_identity_domain"]; ok {
			return nil, nil, errors.Errorf("enforce_identity_domain must works with domains")
		}
	}
	if config := workspaceRawConfig.GetAttr("database_change_mode"); !config.IsNull() {
		workspacePrfile.DatabaseChangeMode = v1pb.DatabaseChangeMode(v1pb.DatabaseChangeMode_value[raw["database_change_mode"].(string)])
		updateMasks = append(updateMasks, "value.workspace_profile.database_change_mode")
	}
	if config := workspaceRawConfig.GetAttr("token_duration_in_seconds"); !config.IsNull() {
		workspacePrfile.TokenDuration = &durationpb.Duration{
			Seconds: int64(raw["token_duration_in_seconds"].(int)),
		}
		updateMasks = append(updateMasks, "value.workspace_profile.token_duration")
	}
	if config := workspaceRawConfig.GetAttr("maximum_role_expiration_in_seconds"); !config.IsNull() {
		workspacePrfile.MaximumRoleExpiration = &durationpb.Duration{
			Seconds: int64(raw["maximum_role_expiration_in_seconds"].(int)),
		}
		updateMasks = append(updateMasks, "value.workspace_profile.maximum_role_expiration")
	}
	if config := workspaceRawConfig.GetAttr("announcement"); !config.IsNull() {
		rawList := raw["announcement"].([]interface{})
		if len(rawList) == 0 {
			workspacePrfile.Announcement = &v1pb.Announcement{}
		} else {
			raw := rawList[0].(map[string]interface{})
			workspacePrfile.Announcement = &v1pb.Announcement{
				Text:  raw["text"].(string),
				Link:  raw["link"].(string),
				Level: v1pb.Announcement_AlertLevel(v1pb.Announcement_AlertLevel_value[raw["level"].(string)]),
			}
		}
		updateMasks = append(updateMasks, "value.workspace_profile.announcement")
	}
	if config := workspaceRawConfig.GetAttr("enable_audit_log_stdout"); !config.IsNull() {
		workspacePrfile.EnableAuditLogStdout = raw["enable_audit_log_stdout"].(bool)
		updateMasks = append(updateMasks, "value.workspace_profile.enable_audit_log_stdout")
	}
	if config := workspaceRawConfig.GetAttr("watermark"); !config.IsNull() {
		workspacePrfile.Watermark = raw["watermark"].(bool)
		updateMasks = append(updateMasks, "value.workspace_profile.watermark")
	}
	if config := workspaceRawConfig.GetAttr("branding_logo"); !config.IsNull() {
		workspacePrfile.BrandingLogo = raw["branding_logo"].(string)
		updateMasks = append(updateMasks, "value.workspace_profile.branding_logo")
	}
	if config := workspaceRawConfig.GetAttr("password_restriction"); !config.IsNull() {
		if pwRestrictionList, ok := raw["password_restriction"].([]interface{}); ok && len(pwRestrictionList) > 0 {
			pwRaw := pwRestrictionList[0].(map[string]interface{})
			pwRestriction := &v1pb.WorkspaceProfileSetting_PasswordRestriction{}
			pwRawConfig := workspaceRawConfig.GetAttr("password_restriction").AsValueSlice()[0]

			if pwConfig := pwRawConfig.GetAttr("min_length"); !pwConfig.IsNull() {
				pwRestriction.MinLength = int32(pwRaw["min_length"].(int))
			}
			if pwConfig := pwRawConfig.GetAttr("require_number"); !pwConfig.IsNull() {
				pwRestriction.RequireNumber = pwRaw["require_number"].(bool)
			}
			if pwConfig := pwRawConfig.GetAttr("require_letter"); !pwConfig.IsNull() {
				pwRestriction.RequireLetter = pwRaw["require_letter"].(bool)
			}
			if pwConfig := pwRawConfig.GetAttr("require_uppercase_letter"); !pwConfig.IsNull() {
				pwRestriction.RequireUppercaseLetter = pwRaw["require_uppercase_letter"].(bool)
			}
			if pwConfig := pwRawConfig.GetAttr("require_special_character"); !pwConfig.IsNull() {
				pwRestriction.RequireSpecialCharacter = pwRaw["require_special_character"].(bool)
			}
			if pwConfig := pwRawConfig.GetAttr("require_reset_password_for_first_login"); !pwConfig.IsNull() {
				pwRestriction.RequireResetPasswordForFirstLogin = pwRaw["require_reset_password_for_first_login"].(bool)
			}
			if pwConfig := pwRawConfig.GetAttr("password_rotation_in_seconds"); !pwConfig.IsNull() {
				pwRestriction.PasswordRotation = &durationpb.Duration{
					Seconds: int64(pwRaw["password_rotation_in_seconds"].(int)),
				}
			}
			workspacePrfile.PasswordRestriction = pwRestriction
			updateMasks = append(updateMasks, "value.workspace_profile.password_restriction")
		}
	}

	return workspacePrfile, updateMasks, nil
}

func convertToV1Level(rawSchema interface{}) *v1pb.DataClassificationSetting_DataClassificationConfig_Level {
	rawLevel := rawSchema.(map[string]interface{})
	return &v1pb.DataClassificationSetting_DataClassificationConfig_Level{
		Id:          rawLevel["id"].(string),
		Title:       rawLevel["title"].(string),
		Description: rawLevel["description"].(string),
	}
}

func convertToV1Classification(rawSchema interface{}) *v1pb.DataClassificationSetting_DataClassificationConfig_DataClassification {
	rawClassification := rawSchema.(map[string]interface{})
	classificationData := &v1pb.DataClassificationSetting_DataClassificationConfig_DataClassification{
		Id:          rawClassification["id"].(string),
		Title:       rawClassification["title"].(string),
		Description: rawClassification["description"].(string),
	}
	levelID, ok := rawClassification["level"].(string)
	if ok {
		classificationData.LevelId = &levelID
	}
	return classificationData
}

func convertToV1ClassificationSetting(d *schema.ResourceData) (*v1pb.DataClassificationSetting, error) {
	rawList, ok := d.Get("classification").([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.Errorf("invalid classification")
	}

	raw := rawList[0].(map[string]interface{})

	dataClassificationConfig := &v1pb.DataClassificationSetting_DataClassificationConfig{
		Id:             raw["id"].(string),
		Title:          raw["title"].(string),
		Levels:         []*v1pb.DataClassificationSetting_DataClassificationConfig_Level{},
		Classification: map[string]*v1pb.DataClassificationSetting_DataClassificationConfig_DataClassification{},
	}
	if dataClassificationConfig.Id == "" {
		return nil, errors.Errorf("id is required for classification config")
	}

	rawLevels := raw["levels"].(*schema.Set)
	if !ok {
		return nil, errors.Errorf("levels is required for classification config")
	}
	for _, level := range rawLevels.List() {
		classificationLevel := convertToV1Level(level)
		if classificationLevel.Id == "" {
			return nil, errors.Errorf("classification level id is required")
		}
		if classificationLevel.Title == "" {
			return nil, errors.Errorf("classification level title is required")
		}
		dataClassificationConfig.Levels = append(dataClassificationConfig.Levels, classificationLevel)
	}

	rawClassificationss := raw["classifications"].(*schema.Set)
	if !ok {
		return nil, errors.Errorf("classifications is required for classification config")
	}
	for _, classification := range rawClassificationss.List() {
		classificationData := convertToV1Classification(classification)
		if classificationData.Id == "" {
			return nil, errors.Errorf("classification id is required")
		}
		if classificationData.Title == "" {
			return nil, errors.Errorf("classification title is required")
		}
		dataClassificationConfig.Classification[classificationData.Id] = classificationData
	}

	return &v1pb.DataClassificationSetting{
		Configs: []*v1pb.DataClassificationSetting_DataClassificationConfig{
			dataClassificationConfig,
		},
	}, nil
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
		flowList, ok := rawRule["flow"].([]interface{})
		if !ok || len(flowList) != 1 {
			return nil, errors.Errorf("invalid flow")
		}
		rawFlow := flowList[0].(map[string]interface{})
		approvalRule := &v1pb.WorkspaceApprovalSetting_Rule{
			Template: &v1pb.ApprovalTemplate{
				Title:       rawFlow["title"].(string),
				Description: rawFlow["description"].(string),
				Flow:        &v1pb.ApprovalFlow{},
			},
			Source: v1pb.WorkspaceApprovalSetting_Rule_Source(v1pb.WorkspaceApprovalSetting_Rule_Source_value[rawRule["source"].(string)]),
			Condition: &expr.Expr{
				Expression: rawRule["condition"].(string),
			},
		}

		roleList, ok := rawFlow["roles"].([]interface{})
		if !ok {
			return nil, errors.Errorf("invalid roles")
		}
		for _, raw := range roleList {
			role := raw.(string)
			if !strings.HasPrefix(role, internal.RoleNamePrefix) {
				return nil, errors.Errorf("invalid role format, role must in roles/{id} format")
			}
			approvalRule.Template.Flow.Roles = append(approvalRule.Template.Flow.Roles, role)
		}

		workspaceApprovalSetting.Rules = append(workspaceApprovalSetting.Rules, approvalRule)
	}

	return workspaceApprovalSetting, nil
}

func convertToV1EnvironmentSetting(d *schema.ResourceData) (*v1pb.EnvironmentSetting, error) {
	rawList, ok := d.Get("environment_setting").([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.Errorf("invalid environment_setting")
	}

	raw := rawList[0].(map[string]interface{})
	environments := raw["environment"].([]interface{})

	environmentSetting := &v1pb.EnvironmentSetting{}
	for _, environment := range environments {
		rawEnv := environment.(map[string]interface{})
		id := rawEnv["id"].(string)
		protected := rawEnv["protected"].(bool)
		if !internal.ResourceIDRegex.MatchString(id) {
			return nil, errors.Errorf("invalid environment id")
		}
		v1Env := &v1pb.EnvironmentSetting_Environment{
			Id:    id,
			Title: rawEnv["title"].(string),
			Color: rawEnv["color"].(string),
			Tags: map[string]string{
				"protected": "protected",
			},
		}
		if !protected {
			v1Env.Tags["protected"] = "unprotected"
		}
		environmentSetting.Environments = append(environmentSetting.Environments, v1Env)
	}
	return environmentSetting, nil
}

func getNumberValueFromSchema(rawSchema map[string]interface{}, field string) int32 {
	raw, ok := rawSchema[field]
	if !ok {
		return 0
	}
	switch v := raw.(type) {
	case int:
		return int32(v)
	case int32:
		return v
	default:
		return 0
	}
}

func convertToV1SemanticType(rawSchema interface{}) (*v1pb.SemanticTypeSetting_SemanticType, error) {
	rawSemanticType := rawSchema.(map[string]interface{})
	semanticType := &v1pb.SemanticTypeSetting_SemanticType{
		Id:          rawSemanticType["id"].(string),
		Title:       rawSemanticType["title"].(string),
		Description: rawSemanticType["description"].(string),
	}
	if semanticType.Id == "" || semanticType.Title == "" {
		return nil, errors.Errorf("semantic type id and title is required")
	}

	algorithmList, ok := rawSemanticType["algorithm"].([]interface{})
	if !ok || len(algorithmList) != 1 {
		return semanticType, nil
	}

	rawAlgorithm := algorithmList[0].(map[string]interface{})
	if fullMasks, ok := rawAlgorithm["full_mask"].([]interface{}); ok && len(fullMasks) == 1 {
		fullMask := fullMasks[0].(map[string]interface{})
		semanticType.Algorithm = &v1pb.Algorithm{
			Mask: &v1pb.Algorithm_FullMask_{
				FullMask: &v1pb.Algorithm_FullMask{
					Substitution: fullMask["substitution"].(string),
				},
			},
		}
	} else if rangeMasks, ok := rawAlgorithm["range_mask"].([]interface{}); ok && len(rangeMasks) == 1 {
		rawRangeMask := rangeMasks[0].(map[string]interface{})
		rangeMask := &v1pb.Algorithm_RangeMask{}
		for _, raw := range rawRangeMask["slices"].([]interface{}) {
			rawSlice := raw.(map[string]interface{})
			rangeMask.Slices = append(rangeMask.Slices, &v1pb.Algorithm_RangeMask_Slice{
				Start:        getNumberValueFromSchema(rawSlice, "start"),
				End:          getNumberValueFromSchema(rawSlice, "end"),
				Substitution: rawSlice["substitution"].(string),
			})
		}
		semanticType.Algorithm = &v1pb.Algorithm{
			Mask: &v1pb.Algorithm_RangeMask_{
				RangeMask: rangeMask,
			},
		}
	} else if md5Masks, ok := rawAlgorithm["md5_mask"].([]interface{}); ok && len(md5Masks) == 1 {
		md5Mask := md5Masks[0].(map[string]interface{})
		semanticType.Algorithm = &v1pb.Algorithm{
			Mask: &v1pb.Algorithm_Md5Mask{
				Md5Mask: &v1pb.Algorithm_MD5Mask{
					Salt: md5Mask["salt"].(string),
				},
			},
		}
	} else if innerOuterMasks, ok := rawAlgorithm["inner_outer_mask"].([]interface{}); ok && len(innerOuterMasks) == 1 {
		innerOuterMask := innerOuterMasks[0].(map[string]interface{})
		t := v1pb.Algorithm_InnerOuterMask_MaskType(
			v1pb.Algorithm_InnerOuterMask_MaskType_value[innerOuterMask["type"].(string)],
		)
		if t == v1pb.Algorithm_InnerOuterMask_MASK_TYPE_UNSPECIFIED {
			return nil, errors.Errorf("invalid inner_outer_mask type: %s", innerOuterMask["type"].(string))
		}
		semanticType.Algorithm = &v1pb.Algorithm{
			Mask: &v1pb.Algorithm_InnerOuterMask_{
				InnerOuterMask: &v1pb.Algorithm_InnerOuterMask{
					PrefixLen:    getNumberValueFromSchema(innerOuterMask, "prefix_len"),
					SuffixLen:    getNumberValueFromSchema(innerOuterMask, "suffix_len"),
					Substitution: innerOuterMask["substitution"].(string),
					Type: v1pb.Algorithm_InnerOuterMask_MaskType(
						v1pb.Algorithm_InnerOuterMask_MaskType_value[innerOuterMask["type"].(string)],
					),
				},
			},
		}
	}

	return semanticType, nil
}

func convertToV1SemanticTypeSetting(d *schema.ResourceData) (*v1pb.SemanticTypeSetting, error) {
	set, ok := d.Get("semantic_types").(*schema.Set)
	if !ok {
		return nil, errors.Errorf("invalid semantic_types")
	}

	setting := &v1pb.SemanticTypeSetting{}
	for _, raw := range set.List() {
		semanticType, err := convertToV1SemanticType(raw)
		if err != nil {
			return nil, err
		}
		setting.Types = append(setting.Types, semanticType)
	}

	return setting, nil
}

func resourceSettingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	settingName := d.Id()
	setting, err := c.GetSetting(ctx, settingName)
	if err != nil {
		// Check if the resource was deleted outside of Terraform
		if internal.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Resource %s not found, removing from state", settingName))
			// Remove from state to trigger recreation on next apply
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return setSettingMessage(d, setting)
}

func resourceSettingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(api.Client)
	settingName := d.Id()

	name, err := internal.GetSettingName(settingName)
	if err != nil {
		return diag.FromErr(err)
	}

	setting := &v1pb.Setting{
		Name: settingName,
	}
	updateMasks := []string{}

	switch name {
	case v1pb.Setting_WORKSPACE_APPROVAL:
		setting.Value = &v1pb.SettingValue{
			Value: &v1pb.SettingValue_WorkspaceApproval{
				WorkspaceApproval: &v1pb.WorkspaceApprovalSetting{},
			},
		}
	case v1pb.Setting_WORKSPACE_PROFILE:
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Unsupport delete workspace profile setting",
			Detail:   "We will reset the workspace profile with default value",
		})
		setting.Value = &v1pb.SettingValue{
			Value: &v1pb.SettingValue_WorkspaceProfile{
				WorkspaceProfile: &v1pb.WorkspaceProfileSetting{
					Announcement: &v1pb.Announcement{},
				},
			},
		}
		updateMasks = []string{
			"value.workspace_profile.disallow_signup",
			"value.workspace_profile.external_url",
			"value.workspace_profile.disallow_password_signin",
			"value.workspace_profile.enforce_identity_domain",
			"value.workspace_profile.domains",
			"value.workspace_profile.database_change_mode",
			"value.workspace_profile.token_duration",
			"value.workspace_profile.maximum_role_expiration",
			"value.workspace_profile.announcement",
			"value.workspace_profile.enable_audit_log_stdout",
		}
	case v1pb.Setting_DATA_CLASSIFICATION:
		setting.Value = &v1pb.SettingValue{
			Value: &v1pb.SettingValue_DataClassification{
				DataClassification: &v1pb.DataClassificationSetting{
					Configs: []*v1pb.DataClassificationSetting_DataClassificationConfig{},
				},
			},
		}
	case v1pb.Setting_SEMANTIC_TYPES:
		setting.Value = &v1pb.SettingValue{
			Value: &v1pb.SettingValue_SemanticType{
				SemanticType: &v1pb.SemanticTypeSetting{},
			},
		}
	case v1pb.Setting_ENVIRONMENT:
		setting.Value = &v1pb.SettingValue{
			Value: &v1pb.SettingValue_Environment{
				Environment: &v1pb.EnvironmentSetting{},
			},
		}
	default:
		return diag.FromErr(errors.Errorf("Unsupport setting: %v", name))
	}

	if _, err := c.UpsertSetting(ctx, setting, updateMasks); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
