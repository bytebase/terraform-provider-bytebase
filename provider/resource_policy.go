package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/type/expr"

	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

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
		Schema: map[string]*schema.Schema{
			"parent": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: internal.ResourceNameValidation(
					// workspace policy
					fmt.Sprintf("^%s$", internal.WorkspaceName),
					// environment policy
					fmt.Sprintf("^%s%s$", internal.EnvironmentNamePrefix, internal.ResourceIDPattern),
					// instance policy
					fmt.Sprintf("^%s%s$", internal.InstanceNamePrefix, internal.ResourceIDPattern),
					// project policy
					fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern),
					// database policy
					fmt.Sprintf(`^%s%s/%s\S+$`, internal.InstanceNamePrefix, internal.ResourceIDPattern, internal.DatabaseIDPrefix),
				),
				Description: "The policy parent name for the policy, support projects/{resource id}, environments/{resource id}, instances/{resource id}, or instances/{resource id}/databases/{database name}",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					v1pb.PolicyType_MASKING_EXCEPTION.String(),
					v1pb.PolicyType_MASKING_RULE.String(),
					v1pb.PolicyType_DISABLE_COPY_DATA.String(),
					v1pb.PolicyType_DATA_SOURCE_QUERY.String(),
					v1pb.PolicyType_ROLLOUT_POLICY.String(),
				}, false),
				Description: "The policy type.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy full name",
			},
			"enforce": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Decide if the policy is enforced.",
			},
			"inherit_from_parent": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Decide if the policy should inherit from the parent.",
			},
			"masking_exception_policy": getMaskingExceptionPolicySchema(false),
			"global_masking_policy":    getGlobalMaskingPolicySchema(false),
			"disable_copy_data_policy": getDisableCopyDataPolicySchema(false),
			"data_source_query_policy": getDataSourceQueryPolicySchema(false),
			"rollout_policy":           getRolloutPolicySchema(false),
		},
	}
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	policyName := d.Id()

	policy, err := c.GetPolicy(ctx, policyName)
	if err != nil {
		return diag.FromErr(err)
	}

	return setPolicyMessage(d, policy)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	policyName := d.Id()
	if err := c.DeletePolicy(ctx, policyName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	policyName := fmt.Sprintf("%s/%s%s", d.Get("parent").(string), internal.PolicyNamePrefix, d.Get("type").(string))
	if strings.HasPrefix(policyName, internal.WorkspaceName) {
		policyName = strings.TrimPrefix(policyName, fmt.Sprintf("%s/", internal.WorkspaceName))
	}

	_, policyType, err := internal.GetPolicyParentAndType(policyName)
	if err != nil {
		return diag.FromErr(err)
	}

	patch := &v1pb.Policy{
		Name:              policyName,
		InheritFromParent: d.Get("inherit_from_parent").(bool),
		Enforce:           d.Get("enforce").(bool),
		Type:              policyType,
	}

	switch policyType {
	case v1pb.PolicyType_MASKING_EXCEPTION:
		maskingExceptionPolicy, err := convertToMaskingExceptionPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.Policy = &v1pb.Policy_MaskingExceptionPolicy{
			MaskingExceptionPolicy: maskingExceptionPolicy,
		}
	case v1pb.PolicyType_MASKING_RULE:
		maskingRulePolicy, err := convertToMaskingRulePolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.Policy = &v1pb.Policy_MaskingRulePolicy{
			MaskingRulePolicy: maskingRulePolicy,
		}
	case v1pb.PolicyType_DISABLE_COPY_DATA:
		if !strings.HasPrefix(policyName, internal.EnvironmentNamePrefix) && !strings.HasPrefix(policyName, internal.ProjectNamePrefix) {
			return diag.Errorf("policy %v only support environment or project resource", policyName)
		}
		disableCopyDataPolicy, err := convertToDisableCopyDataPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.Policy = &v1pb.Policy_DisableCopyDataPolicy{
			DisableCopyDataPolicy: disableCopyDataPolicy,
		}
	case v1pb.PolicyType_DATA_SOURCE_QUERY:
		if !strings.HasPrefix(policyName, internal.EnvironmentNamePrefix) && !strings.HasPrefix(policyName, internal.ProjectNamePrefix) {
			return diag.Errorf("policy %v only support environment or project resource", policyName)
		}
		dataSourceQueryPolicy, err := convertToDataSourceQueryPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.Policy = &v1pb.Policy_DataSourceQueryPolicy{
			DataSourceQueryPolicy: dataSourceQueryPolicy,
		}
	case v1pb.PolicyType_ROLLOUT_POLICY:
		if !strings.HasPrefix(policyName, internal.EnvironmentNamePrefix) {
			return diag.Errorf("policy %v only support environment resource", policyName)
		}
		rolloutPolicy, err := convertToRolloutPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.Policy = &v1pb.Policy_RolloutPolicy{
			RolloutPolicy: rolloutPolicy,
		}
	default:
		return diag.Errorf("unsupport policy type: %v", policyName)
	}

	updateMasks := []string{"payload"}
	rawConfig := d.GetRawConfig()
	if config := rawConfig.GetAttr("inherit_from_parent"); !config.IsNull() {
		updateMasks = append(updateMasks, "inherit_from_parent")
	}
	if config := rawConfig.GetAttr("enforce"); !config.IsNull() {
		updateMasks = append(updateMasks, "enforce")
	}

	var diags diag.Diagnostics
	p, err := c.UpsertPolicy(ctx, patch, updateMasks)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to upsert policy",
			Detail:   fmt.Sprintf("Upsert policy %s failed, error: %v", policyName, err),
		})
		return diags
	}

	d.SetId(p.Name)

	diag := resourcePolicyRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	policyName := d.Id()

	_, policyType, err := internal.GetPolicyParentAndType(policyName)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("type") || d.HasChange("parent") {
		return diag.Errorf("cannot change policy type or parent")
	}

	patch := &v1pb.Policy{
		Name:              policyName,
		InheritFromParent: d.Get("inherit_from_parent").(bool),
		Enforce:           d.Get("enforce").(bool),
		Type:              policyType,
	}

	updateMasks := []string{}
	if d.HasChange("inherit_from_parent") {
		updateMasks = append(updateMasks, "inherit_from_parent")
	}
	if d.HasChange("enforce") {
		updateMasks = append(updateMasks, "enforce")
	}

	if d.HasChange("masking_exception_policy") {
		updateMasks = append(updateMasks, "payload")
		maskingExceptionPolicy, err := convertToMaskingExceptionPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.Policy = &v1pb.Policy_MaskingExceptionPolicy{
			MaskingExceptionPolicy: maskingExceptionPolicy,
		}
	}
	if d.HasChange("global_masking_policy") {
		updateMasks = append(updateMasks, "payload")
		maskingRulePolicy, err := convertToMaskingRulePolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.Policy = &v1pb.Policy_MaskingRulePolicy{
			MaskingRulePolicy: maskingRulePolicy,
		}
	}
	if d.HasChange("disable_copy_data_policy") {
		updateMasks = append(updateMasks, "payload")
		disableCopyDataPolicy, err := convertToDisableCopyDataPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.Policy = &v1pb.Policy_DisableCopyDataPolicy{
			DisableCopyDataPolicy: disableCopyDataPolicy,
		}
	}
	if d.HasChange("data_source_query_policy") {
		updateMasks = append(updateMasks, "payload")
		dataSourceQueryPolicy, err := convertToDataSourceQueryPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.Policy = &v1pb.Policy_DataSourceQueryPolicy{
			DataSourceQueryPolicy: dataSourceQueryPolicy,
		}
	}

	var diags diag.Diagnostics
	if len(updateMasks) > 0 {
		if _, err := c.UpsertPolicy(ctx, patch, updateMasks); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to upsert policy",
				Detail:   fmt.Sprintf("Upsert policy %s failed, error: %v", policyName, err),
			})
			return diags
		}
	}

	diag := resourcePolicyRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func convertToMaskingRulePolicy(d *schema.ResourceData) (*v1pb.MaskingRulePolicy, error) {
	rawList, ok := d.Get("global_masking_policy").([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.Errorf("invalid global_masking_policy")
	}

	raw := rawList[0].(map[string]interface{})
	ruleList, ok := raw["rules"].([]interface{})
	if !ok {
		return nil, errors.Errorf("invalid masking rules")
	}

	policy := &v1pb.MaskingRulePolicy{}

	for _, rule := range ruleList {
		rawRule := rule.(map[string]interface{})
		title := rawRule["title"].(string)
		policy.Rules = append(policy.Rules, &v1pb.MaskingRulePolicy_MaskingRule{
			Id:           rawRule["id"].(string),
			SemanticType: rawRule["semantic_type"].(string),
			Condition: &expr.Expr{
				Title:      title,
				Expression: rawRule["condition"].(string),
			},
		})
	}

	return policy, nil
}

func convertToMaskingExceptionPolicy(d *schema.ResourceData) (*v1pb.MaskingExceptionPolicy, error) {
	rawList, ok := d.Get("masking_exception_policy").([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.Errorf("invalid masking_exception_policy")
	}

	raw := rawList[0].(map[string]interface{})
	exceptionList, ok := raw["exceptions"].(*schema.Set)
	if !ok {
		return nil, errors.Errorf("invalid exceptions")
	}

	policy := &v1pb.MaskingExceptionPolicy{}

	for _, exception := range exceptionList.List() {
		rawException := exception.(map[string]interface{})

		expressions := []string{}
		databaseFullName := rawException["database"].(string)
		if databaseFullName != "" {
			instanceID, databaseName, err := internal.GetInstanceDatabaseID(databaseFullName)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid database full name: %v", databaseFullName)
			}
			expressions = append(
				expressions,
				fmt.Sprintf(`resource.instance_id == "%s"`, instanceID),
				fmt.Sprintf(`resource.database_name == "%s"`, databaseName),
			)

			if schema, ok := rawException["schema"].(string); ok && schema != "" {
				expressions = append(expressions, fmt.Sprintf(`resource.schema_name == "%s"`, schema))
			}
			if table, ok := rawException["table"].(string); ok && table != "" {
				expressions = append(expressions, fmt.Sprintf(`resource.table_name == "%s"`, table))
			}
			if column, ok := rawException["column"].(string); ok && column != "" {
				expressions = append(expressions, fmt.Sprintf(`resource.column_name == "%s"`, column))
			}
		}

		if expire, ok := rawException["expire_timestamp"].(string); ok && expire != "" {
			formattedTime, err := time.Parse(time.RFC3339, expire)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid time: %v", expire)
			}
			expressions = append(expressions, fmt.Sprintf(`request.time < timestamp("%s")`, formattedTime.Format(time.RFC3339)))
		}
		member := rawException["member"].(string)
		if member == "allUsers" {
			return nil, errors.Errorf("not support allUsers in masking_exception_policy")
		}
		if err := internal.ValidateMemberBinding(member); err != nil {
			return nil, err
		}
		policy.MaskingExceptions = append(policy.MaskingExceptions, &v1pb.MaskingExceptionPolicy_MaskingException{
			Member: member,
			Action: v1pb.MaskingExceptionPolicy_MaskingException_Action(
				v1pb.MaskingExceptionPolicy_MaskingException_Action_value[rawException["action"].(string)],
			),
			Condition: &expr.Expr{
				Description: rawException["reason"].(string),
				Expression:  strings.Join(expressions, " && "),
			},
		})
	}
	return policy, nil
}

func convertToRolloutPolicy(d *schema.ResourceData) (*v1pb.RolloutPolicy, error) {
	rawList, ok := d.Get("rollout_policy").([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.Errorf("invalid rollout_policy")
	}

	raw := rawList[0].(map[string]interface{})
	policy := &v1pb.RolloutPolicy{
		Automatic: raw["automatic"].(bool),
	}

	roles, ok := raw["roles"].(*schema.Set)
	if !ok {
		return policy, nil
	}

	for _, rawRole := range roles.List() {
		role := rawRole.(string)
		if role == issueLastApproverRole || role == issueCreatorRole {
			policy.IssueRoles = append(policy.IssueRoles, role)
		} else {
			policy.Roles = append(policy.Roles, role)
		}
	}

	return policy, nil
}

func convertToDisableCopyDataPolicy(d *schema.ResourceData) (*v1pb.DisableCopyDataPolicy, error) {
	rawList, ok := d.Get("disable_copy_data_policy").([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.Errorf("invalid disable_copy_data_policy")
	}

	raw := rawList[0].(map[string]interface{})
	return &v1pb.DisableCopyDataPolicy{
		Active: raw["enable"].(bool),
	}, nil
}

func convertToDataSourceQueryPolicy(d *schema.ResourceData) (*v1pb.DataSourceQueryPolicy, error) {
	rawList, ok := d.Get("data_source_query_policy").([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.Errorf("invalid data_source_query_policy")
	}

	raw := rawList[0].(map[string]interface{})
	return &v1pb.DataSourceQueryPolicy{
		AdminDataSourceRestriction: v1pb.DataSourceQueryPolicy_Restriction(
			v1pb.DataSourceQueryPolicy_Restriction_value[raw["restriction"].(string)],
		),
		DisallowDdl: raw["disallow_ddl"].(bool),
		DisallowDml: raw["disallow_dml"].(bool),
	}, nil
}
