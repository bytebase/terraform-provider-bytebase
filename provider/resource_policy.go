package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

// TODO(ed): add test and doc.
func resourcePolicy() *schema.Resource {
	return &schema.Resource{
		Description:   "The policy resource.",
		CreateContext: resourcePolicyCreate,
		ReadContext:   resourcePolicyRead,
		UpdateContext: resourcePolicyUpdate,
		DeleteContext: resourcePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: getPolicySchema(map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(api.PolicyTypeDeploymentApproval),
					string(api.PolicyTypeBackupPlan),
					string(api.PolicyTypeSQLReview),
					string(api.PolicyTypeSensitiveData),
					string(api.PolicyTypeAccessControl),
				}, false),
				Description: "The policy type.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy name",
			},
			"inherit_from_parent": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Decide if the policy should inherit from the parent.",
			},
			"deployment_approval_policy": getDeploymentApprovalPolicySchema(false),
			"backup_plan_policy":         getBackupPlanPolicySchema(false),
			"sensitive_data_policy":      getSensitiveDataPolicy(false),
			"access_control_policy":      getAccessControlPolicy(false),
			"sql_review_policy":          getSQLReviewPolicy(false),
		}),
	}
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	find, err := internal.GetPolicyFindMessageByName(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	policy, err := c.GetPolicy(ctx, find)
	if err != nil {
		return diag.FromErr(err)
	}

	return setPolicyMessage(d, policy)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	find, err := internal.GetPolicyFindMessageByName(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := c.DeletePolicy(ctx, find); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	find, err := getPolicyFind(d)
	if err != nil {
		return diag.FromErr(err)
	}

	inheritFromParent := d.Get("inherit_from_parent").(bool)

	patch := &api.PolicyPatchMessage{
		InheritFromParent: &inheritFromParent,
		Type:              *find.Type,
	}

	if _, ok := d.GetOk("deployment_approval_policy"); ok {
		policy, err := convertDeploymentApprovalPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.DeploymentApprovalPolicy = policy
	}
	if _, ok := d.GetOk("backup_plan_policy"); ok {
		policy, err := convertBackupPlanPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.BackupPlanPolicy = policy
	}
	if _, ok := d.GetOk("sensitive_data_policy"); ok {
		policy, err := convertSensitiveDataPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.SensitiveDataPolicy = policy
	}
	if _, ok := d.GetOk("access_control_policy"); ok {
		policy, err := convertAccessControlPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.AccessControlPolicy = policy
	}
	if _, ok := d.GetOk("sql_review_policy"); ok {
		policy, err := convertSQLReviewPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.SQLReviewPolicy = policy
	}

	if err := validatePolicy(find, patch); err != nil {
		return diag.FromErr(err)
	}

	p, err := c.UpsertPolicy(ctx, find, patch)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(p.Name)
	return resourcePolicyRead(ctx, d, m)
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	find, err := getPolicyFind(d)
	if err != nil {
		return diag.FromErr(err)
	}

	patch := &api.PolicyPatchMessage{
		Type: *find.Type,
	}

	if d.HasChange("inherit_from_parent") {
		v := d.Get("inherit_from_parent").(bool)
		patch.InheritFromParent = &v
	}

	if d.HasChange("deployment_approval_policy") {
		policy, err := convertDeploymentApprovalPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.DeploymentApprovalPolicy = policy
	}
	if d.HasChange("backup_plan_policy") {
		policy, err := convertBackupPlanPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.BackupPlanPolicy = policy
	}
	if d.HasChange("sensitive_data_policy") {
		policy, err := convertSensitiveDataPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.SensitiveDataPolicy = policy
	}
	if d.HasChange("access_control_policy") {
		policy, err := convertAccessControlPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.AccessControlPolicy = policy
	}
	if d.HasChange("sql_review_policy") {
		policy, err := convertSQLReviewPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.SQLReviewPolicy = policy
	}

	if err := validatePolicy(find, patch); err != nil {
		return diag.FromErr(err)
	}

	if _, err := c.UpsertPolicy(ctx, find, patch); err != nil {
		return diag.FromErr(err)
	}

	return resourcePolicyRead(ctx, d, m)
}

func convertDeploymentApprovalPolicy(d *schema.ResourceData) (*api.DeploymentApprovalPolicy, error) {
	rawList, ok := d.Get("deployment_approval_policy").([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.Errorf("invalid deployment_approval_policy")
	}

	raw := rawList[0].(map[string]interface{})
	strategies := raw["deployment_approval_strategies"].([]interface{})

	policy := &api.DeploymentApprovalPolicy{
		DefaultStrategy: api.ApprovalStrategy(raw["default_strategy"].(string)),
	}

	for _, strategy := range strategies {
		rawStrategy := strategy.(map[string]interface{})
		policy.DeploymentApprovalStrategies = append(policy.DeploymentApprovalStrategies, &api.DeploymentApprovalStrategy{
			ApprovalGroup:    api.ApprovalGroup(rawStrategy["approval_group"].(string)),
			ApprovalStrategy: api.ApprovalStrategy(rawStrategy["approval_strategy"].(string)),
			DeploymentType:   api.DeploymentType(rawStrategy["deployment_type"].(string)),
		})
	}
	return policy, nil
}

func convertBackupPlanPolicy(d *schema.ResourceData) (*api.BackupPlanPolicy, error) {
	rawList, ok := d.Get("backup_plan_policy").([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.Errorf("invalid backup_plan_policy")
	}

	raw := rawList[0].(map[string]interface{})
	return &api.BackupPlanPolicy{
		Schedule:          api.BackupPlanSchedule(raw["schedule"].(string)),
		RetentionDuration: fmt.Sprintf("%ds", raw["retention_duration"].(int)),
	}, nil
}

func convertSensitiveDataPolicy(d *schema.ResourceData) (*api.SensitiveDataPolicy, error) {
	rawList, ok := d.Get("sensitive_data_policy").([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.Errorf("invalid sensitive_data_policy")
	}

	raw := rawList[0].(map[string]interface{})
	dataList := raw["sensitive_data"].([]interface{})
	policy := &api.SensitiveDataPolicy{}

	for _, data := range dataList {
		rawData := data.(map[string]interface{})
		policy.SensitiveData = append(policy.SensitiveData, &api.SensitiveData{
			Schema:   rawData["schema"].(string),
			Table:    rawData["table"].(string),
			Column:   rawData["column"].(string),
			MaskType: api.SensitiveDataMaskType(rawData["mask_type"].(string)),
		})
	}

	return policy, nil
}

func convertAccessControlPolicy(d *schema.ResourceData) (*api.AccessControlPolicy, error) {
	rawList, ok := d.Get("access_control_policy").([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.Errorf("invalid access_control_policy")
	}

	raw := rawList[0].(map[string]interface{})
	rules := raw["disallow_rules"].([]interface{})
	policy := &api.AccessControlPolicy{}

	for _, rule := range rules {
		rawRule := rule.(map[string]interface{})
		policy.DisallowRules = append(policy.DisallowRules, &api.AccessControlRule{
			FullDatabase: rawRule["full_database"].(bool),
		})
	}

	return policy, nil
}

func convertSQLReviewPolicy(d *schema.ResourceData) (*api.SQLReviewPolicy, error) {
	rawList, ok := d.Get("sql_review_policy").([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.Errorf("invalid sql_review_policy")
	}

	raw := rawList[0].(map[string]interface{})
	rules := raw["rules"].([]interface{})
	if len(rules) == 0 {
		return nil, errors.Errorf("`rules` is required sql_review_policy")
	}

	title, ok := raw["title"].(string)
	if !ok || title == "" {
		return nil, errors.Errorf("`title` is required sql_review_policy")
	}

	policy := &api.SQLReviewPolicy{
		Title: title,
	}

	ruleMap := map[api.SQLReviewRuleType]bool{}
	for _, rule := range rules {
		rawRule := rule.(map[string]interface{})
		ruleType := api.SQLReviewRuleType(rawRule["type"].(string))
		if ruleMap[ruleType] {
			return nil, errors.Errorf("duplicate rule %v", ruleType)
		}
		ruleMap[ruleType] = true

		payloadRawList, ok := rawRule["payload"].([]interface{})
		if !ok || len(payloadRawList) != 1 {
			return nil, errors.Errorf("invalid sql_review_policy payload, only expect one payload for the rule %v", rawRule["type"].(string))
		}

		payloadRaw := payloadRawList[0].(map[string]interface{})
		payloadStr, err := mrshalSQLReviewRulePayload(ruleType, payloadRaw)
		if err != nil {
			return nil, err
		}

		policy.Rules = append(policy.Rules, &api.SQLReviewRule{
			Type:    ruleType,
			Level:   api.SQLReviewRuleLevel(rawRule["level"].(string)),
			Payload: payloadStr,
		})
	}

	return policy, nil
}

func validatePolicy(find *api.PolicyFindMessage, patch *api.PolicyPatchMessage) error {
	switch patch.Type {
	case api.PolicyTypeDeploymentApproval:
		if patch.DeploymentApprovalPolicy == nil {
			return errors.Errorf("must set deployment_approval_policy for %v policy", patch.Type)
		}
	case api.PolicyTypeBackupPlan:
		if patch.BackupPlanPolicy == nil {
			return errors.Errorf("must set backup_plan_policy for %v policy", patch.Type)
		}
	case api.PolicyTypeSensitiveData:
		if patch.SensitiveDataPolicy == nil {
			return errors.Errorf("must set sensitive_data_policy for %v policy", patch.Type)
		}
		if find.DatabaseName == nil {
			return errors.Errorf("%v policy only works for the database resource", patch.Type)
		}
	case api.PolicyTypeAccessControl:
		if patch.AccessControlPolicy == nil {
			return errors.Errorf("must set access_control_policy for %v policy", patch.Type)
		}
		if find.ProjectID != nil || (find.InstanceID != nil && find.DatabaseName == nil) {
			return errors.Errorf("%v policy only works for the environment and database resource", patch.Type)
		}
	case api.PolicyTypeSQLReview:
		if patch.SQLReviewPolicy == nil {
			return errors.Errorf("must set sql_review_policy for %v policy", patch.Type)
		}
	}

	if patch.DeploymentApprovalPolicy != nil {
		if patch.Type != api.PolicyTypeDeploymentApproval {
			return errors.Errorf("the policy payload deployment_approval_policy not matchs the policy type %v", patch.Type)
		}
	}

	if patch.BackupPlanPolicy != nil {
		if patch.Type != api.PolicyTypeBackupPlan {
			return errors.Errorf("the policy payload backup_plan_policy not matchs the policy type %v", patch.Type)
		}
	}

	if patch.SensitiveDataPolicy != nil {
		if patch.Type != api.PolicyTypeSensitiveData {
			return errors.Errorf("the policy payload sensitive_data_policy not matchs the policy type %v", patch.Type)
		}
	}

	if patch.AccessControlPolicy != nil {
		if patch.Type != api.PolicyTypeAccessControl {
			return errors.Errorf("the policy payload access_control_policy not matchs the policy type %v", patch.Type)
		}
	}

	if patch.SQLReviewPolicy != nil {
		if patch.Type != api.PolicyTypeSQLReview {
			return errors.Errorf("the policy payload sql_review_policy not matchs the policy type %v", patch.Type)
		}
	}

	return nil
}

func mrshalSQLReviewRulePayload(ruleType api.SQLReviewRuleType, payload map[string]interface{}) (string, error) {
	switch ruleType {
	case
		api.SchemaRuleTableNaming,
		api.SchemaRuleColumnNaming,
		api.SchemaRuleAutoIncrementColumnNaming,
		api.SchemaRuleFKNaming,
		api.SchemaRuleIDXNaming,
		api.SchemaRuleUKNaming:
		format, ok := payload["format"].(string)
		if !ok || format == "" {
			return "", errors.Errorf("`format` is required to set for SQL review rule %v", ruleType)
		}
		maxLength, ok := payload["max_length"].(int)
		if !ok || maxLength == 0 {
			return "", errors.Errorf("`max_length` is required to set for SQL review rule %v", ruleType)
		}
		res, err := json.Marshal(&api.NamingRulePayload{
			Format:    format,
			MaxLength: maxLength,
		})
		if err != nil {
			return "", errors.Errorf("failed to marshal payload for SQL review rule %v with error %v", ruleType, err.Error())
		}

		return string(res), nil
	case
		api.SchemaRuleColumnCommentConvention,
		api.SchemaRuleTableCommentConvention:
		required, ok := payload["required"].(bool)
		if !ok {
			return "", errors.Errorf("`required` is required to set for SQL review rule %v", ruleType)
		}
		maxLength, ok := payload["max_length"].(int)
		if !ok || maxLength == 0 {
			return "", errors.Errorf("`max_length` is required to set for SQL review rule %v", ruleType)
		}
		res, err := json.Marshal(&api.CommentConventionRulePayload{
			Required:  required,
			MaxLength: maxLength,
		})
		if err != nil {
			return "", errors.Errorf("failed to marshal payload for SQL review rule %v with error %v", ruleType, err.Error())
		}

		return string(res), nil
	case
		api.SchemaRuleIndexKeyNumberLimit,
		api.SchemaRuleStatementInsertRowLimit,
		api.SchemaRuleIndexTotalNumberLimit,
		api.SchemaRuleColumnMaximumCharacterLength,
		api.SchemaRuleColumnAutoIncrementInitialValue,
		api.SchemaRuleStatementAffectedRowLimit:
		number, ok := payload["number"].(int)
		if !ok {
			return "", errors.Errorf("`number` is required to set for SQL review rule %v", ruleType)
		}
		res, err := json.Marshal(&api.NumberTypeRulePayload{
			Number: number,
		})
		if err != nil {
			return "", errors.Errorf("failed to marshal payload for SQL review rule %v with error %v", ruleType, err.Error())
		}

		return string(res), nil
	case
		api.SchemaRuleRequiredColumn,
		api.SchemaRuleColumnTypeDisallowList,
		api.SchemaRuleCharsetAllowlist,
		api.SchemaRuleCollationAllowlist,
		api.SchemaRuleIndexPrimaryKeyTypeAllowlist:
		list, ok := payload["list"].([]interface{})
		if !ok {
			return "", errors.Errorf("`list` is required to set for SQL review rule %v", ruleType)
		}
		if len(list) == 0 {
			return "", errors.Errorf("`list` atleast contains one element for SQL review rule %v", ruleType)
		}

		stringArray := make([]string, len(list))
		for i, v := range list {
			stringArray[i] = fmt.Sprintf("%v", v)
		}

		res, err := json.Marshal(&api.StringArrayTypeRulePayload{
			List: stringArray,
		})
		if err != nil {
			return "", errors.Errorf("failed to marshal payload for SQL review rule %v with error %v", ruleType, err.Error())
		}

		return string(res), nil
	}

	if len(payload) > 0 {
		return "", errors.Errorf("cannot set payload for the SQL review rule %v", ruleType)
	}

	return "", nil
}
