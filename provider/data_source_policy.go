package provider

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourcePolicy() *schema.Resource {
	return &schema.Resource{
		Description: "The policy data source.",
		ReadContext: dataSourcePolicyRead,
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
					fmt.Sprintf("^%s%s/%s%s$", internal.InstanceNamePrefix, internal.ResourceIDPattern, internal.DatabaseIDPrefix, internal.ResourceIDPattern),
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
			"inherit_from_parent": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Decide if the policy should inherit from the parent.",
			},
			"enforce": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Decide if the policy is enforced.",
			},
			"masking_exception_policy": getMaskingExceptionPolicySchema(true),
			"global_masking_policy":    getGlobalMaskingPolicySchema(true),
			"disable_copy_data_policy": getDisableCopyDataPolicySchema(true),
			"data_source_query_policy": getDataSourceQueryPolicySchema(true),
			"rollout_policy":           getRolloutPolicySchema(true),
		},
	}
}

func getMaskingExceptionPolicySchema(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed: computed,
		Optional: true,
		Default:  nil,
		Type:     schema.TypeList,
		MinItems: 0,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"exceptions": {
					Computed: computed,
					Optional: true,
					Default:  nil,
					MinItems: 0,
					Type:     schema.TypeSet,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"database": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringIsNotEmpty,
								Description:  "The database full name in instances/{instance resource id}/databases/{database name} format",
							},
							"schema": {
								Type:     schema.TypeString,
								Computed: computed,
								Optional: true,
							},
							"table": {
								Type:         schema.TypeString,
								Computed:     computed,
								Optional:     true,
								ValidateFunc: validation.StringIsNotEmpty,
							},
							"column": {
								Type:         schema.TypeString,
								Computed:     computed,
								Optional:     true,
								ValidateFunc: validation.StringIsNotEmpty,
							},
							"member": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringIsNotEmpty,
								Description:  "The member in user:{email} or group:{email} format.",
							},
							"action": {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.StringInSlice([]string{
									v1pb.MaskingExceptionPolicy_MaskingException_QUERY.String(),
									v1pb.MaskingExceptionPolicy_MaskingException_EXPORT.String(),
								}, false),
							},
							"reason": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The reason for the masking exemption",
							},
							"expire_timestamp": {
								Type:        schema.TypeString,
								Computed:    computed,
								Optional:    true,
								Description: "The expiration timestamp in YYYY-MM-DDThh:mm:ss.000Z format",
							},
						},
					},
					Set: exceptionHash,
				},
			},
		},
	}
}

func getGlobalMaskingPolicySchema(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed: computed,
		Optional: true,
		Default:  nil,
		Type:     schema.TypeList,
		MinItems: 0,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"rules": {
					Computed: computed,
					Optional: true,
					Default:  nil,
					MinItems: 0,
					Type:     schema.TypeList,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringIsNotEmpty,
								Description:  "The unique rule id",
							},
							"title": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The title for the rule",
							},
							"semantic_type": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringIsNotEmpty,
								Description:  "The semantic type id",
							},
							"condition": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringIsNotEmpty,
								Description:  "The condition expression",
							},
						},
					},
				},
			},
		},
	}
}

func getDisableCopyDataPolicySchema(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed:    computed,
		Optional:    true,
		Default:     nil,
		Type:        schema.TypeList,
		MinItems:    0,
		MaxItems:    1,
		Description: "Restrict data copying in SQL Editor (Admins/DBAs allowed)",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enable": {
					Type:        schema.TypeBool,
					Required:    true,
					Description: "Restrict data copying",
				},
			},
		},
	}
}

func getDataSourceQueryPolicySchema(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed:    computed,
		Optional:    true,
		Default:     nil,
		Type:        schema.TypeList,
		MinItems:    0,
		MaxItems:    1,
		Description: "Restrict querying admin data sources",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"restriction": {
					Type:     schema.TypeString,
					Optional: true,
					ValidateFunc: validation.StringInSlice([]string{
						v1pb.DataSourceQueryPolicy_FALLBACK.String(),
						v1pb.DataSourceQueryPolicy_DISALLOW.String(),
						v1pb.DataSourceQueryPolicy_RESTRICTION_UNSPECIFIED.String(),
					}, false),
					Description: "RESTRICTION_UNSPECIFIED means no restriction; FALLBACK will allows to query admin data sources when there is no read-only data source; DISALLOW will always disallow to query admin data sources.",
				},
				"disallow_ddl": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Disallow running DDL statements in the SQL editor.",
				},
				"disallow_dml": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Disallow running DML statements in the SQL editor.",
				},
			},
		},
	}
}

const (
	issueLastApproverRole = "roles/LAST_APPROVER"
	issueCreatorRole      = "roles/CREATOR"
)

func getRolloutPolicySchema(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed:    computed,
		Optional:    true,
		Default:     nil,
		Type:        schema.TypeList,
		MinItems:    0,
		MaxItems:    1,
		Description: "Control issue rollout. Learn more: https://docs.bytebase.com/administration/environment-policy/rollout-policy",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"automatic": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "If all check pass, the change will be rolled out and executed automatically.",
				},
				"roles": {
					Optional:    true,
					Type:        schema.TypeSet,
					MinItems:    0,
					Description: "If any roles are specified, Bytebase requires users with those roles to manually roll out the change.",
					Elem: &schema.Schema{
						Type:        schema.TypeString,
						Description: fmt.Sprintf(`Role full name in roles/{id} format. You can also use the "%s" for the last approver of the issue, or "%s" for the creator of the issue.`, issueLastApproverRole, issueCreatorRole),
						ValidateDiagFunc: internal.ResourceNameValidation(
							fmt.Sprintf("^%s$", issueLastApproverRole),
							fmt.Sprintf("^%s$", issueCreatorRole),
							fmt.Sprintf("^%s", internal.RoleNamePrefix),
						),
					},
				},
			},
		},
	}
}

func dataSourcePolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	policyName := fmt.Sprintf("%s/%s%s", d.Get("parent").(string), internal.PolicyNamePrefix, d.Get("type").(string))
	if strings.HasPrefix(policyName, internal.WorkspaceName) {
		policyName = strings.TrimPrefix(policyName, fmt.Sprintf("%s/", internal.WorkspaceName))
	}

	policy, err := c.GetPolicy(ctx, policyName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(policy.Name)
	return setPolicyMessage(d, policy)
}

func setPolicyMessage(d *schema.ResourceData, policy *v1pb.Policy) diag.Diagnostics {
	_, policyType, err := internal.GetPolicyParentAndType(policy.Name)
	if err != nil {
		return diag.Errorf("cannot parse name for policy: %s", err.Error())
	}
	if err := d.Set("name", policy.Name); err != nil {
		return diag.Errorf("cannot set name for policy: %s", err.Error())
	}
	if err := d.Set("inherit_from_parent", policy.InheritFromParent); err != nil {
		return diag.Errorf("cannot set inherit_from_parent for policy: %s", err.Error())
	}
	if err := d.Set("enforce", policy.Enforce); err != nil {
		return diag.Errorf("cannot set enforce for policy: %s", err.Error())
	}

	switch policyType {
	case v1pb.PolicyType_MASKING_EXCEPTION:
		if p := policy.GetMaskingExceptionPolicy(); p != nil {
			exceptionPolicy, err := flattenMaskingExceptionPolicy(p)
			if err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("masking_exception_policy", exceptionPolicy); err != nil {
				return diag.Errorf("cannot set masking_exception_policy: %s", err.Error())
			}
		}
	case v1pb.PolicyType_MASKING_RULE:
		if p := policy.GetMaskingRulePolicy(); p != nil {
			maskingPolicy, err := flattenGlobalMaskingPolicy(p)
			if err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("global_masking_policy", maskingPolicy); err != nil {
				return diag.Errorf("cannot set global_masking_policy: %s", err.Error())
			}
		}
	case v1pb.PolicyType_DISABLE_COPY_DATA:
		if p := policy.GetDisableCopyDataPolicy(); p != nil {
			disableCopyDataPolicy := flattenDisableCopyDataPolicy(p)
			if err := d.Set("disable_copy_data_policy", disableCopyDataPolicy); err != nil {
				return diag.Errorf("cannot set disable_copy_data_policy: %s", err.Error())
			}
		}
	case v1pb.PolicyType_DATA_SOURCE_QUERY:
		if p := policy.GetDataSourceQueryPolicy(); p != nil {
			dataSourceQueryPolicy := flattenDataSourceQueryPolicy(p)
			if err := d.Set("data_source_query_policy", dataSourceQueryPolicy); err != nil {
				return diag.Errorf("cannot set data_source_query_policy: %s", err.Error())
			}
		}
	case v1pb.PolicyType_ROLLOUT_POLICY:
		if p := policy.GetRolloutPolicy(); p != nil {
			rolloutPolicy := flattenRolloutPolicy(p)
			if err := d.Set("rollout_policy", rolloutPolicy); err != nil {
				return diag.Errorf("cannot set rollout_policy: %s", err.Error())
			}
		}
	}

	return nil
}

func flattenRolloutPolicy(p *v1pb.RolloutPolicy) []interface{} {
	roles := []string{}
	roles = append(roles, p.Roles...)
	roles = append(roles, p.IssueRoles...)
	policy := map[string]interface{}{
		"automatic": p.Automatic,
		"roles":     roles,
	}
	return []interface{}{policy}
}

func flattenDataSourceQueryPolicy(p *v1pb.DataSourceQueryPolicy) []interface{} {
	policy := map[string]interface{}{
		"restriction":  p.AdminDataSourceRestriction.String(),
		"disallow_ddl": p.DisallowDdl,
		"disallow_dml": p.DisallowDml,
	}
	return []interface{}{policy}
}

func flattenDisableCopyDataPolicy(p *v1pb.DisableCopyDataPolicy) []interface{} {
	policy := map[string]interface{}{
		"enable": p.Active,
	}
	return []interface{}{policy}
}

func flattenGlobalMaskingPolicy(p *v1pb.MaskingRulePolicy) ([]interface{}, error) {
	ruleList := []interface{}{}

	for _, rule := range p.Rules {
		if rule.Condition == nil || rule.Condition.Expression == "" {
			return nil, errors.Errorf("invalid global masking policy condition")
		}
		raw := map[string]interface{}{}
		raw["id"] = rule.Id
		raw["semantic_type"] = rule.SemanticType
		raw["condition"] = rule.Condition.Expression
		raw["title"] = rule.Condition.Title

		ruleList = append(ruleList, raw)
	}

	policy := map[string]interface{}{
		"rules": ruleList,
	}
	return []interface{}{policy}, nil
}

func flattenMaskingExceptionPolicy(p *v1pb.MaskingExceptionPolicy) ([]interface{}, error) {
	exceptionList := []interface{}{}
	for _, exception := range p.MaskingExceptions {
		raw := map[string]interface{}{}
		raw["member"] = exception.Member
		raw["action"] = exception.Action.String()

		if exception.Condition == nil || exception.Condition.Expression == "" {
			return nil, errors.Errorf("invalid exception policy condition")
		}
		raw["reason"] = exception.Condition.Description

		expressions := strings.Split(exception.Condition.Expression, " && ")
		instanceID := ""
		databaseName := ""
		for _, expression := range expressions {
			if strings.HasPrefix(expression, "resource.instance_id == ") {
				instanceID = strings.TrimSuffix(
					strings.TrimPrefix(expression, `resource.instance_id == "`),
					`"`,
				)
			}
			if strings.HasPrefix(expression, "resource.database_name == ") {
				databaseName = strings.TrimSuffix(
					strings.TrimPrefix(expression, `resource.database_name == "`),
					`"`,
				)
			}
			if strings.HasPrefix(expression, "resource.table_name == ") {
				raw["table"] = strings.TrimSuffix(
					strings.TrimPrefix(expression, `resource.table_name == "`),
					`"`,
				)
			}
			if strings.HasPrefix(expression, "resource.schema_name == ") {
				raw["schema"] = strings.TrimSuffix(
					strings.TrimPrefix(expression, `resource.schema_name == "`),
					`"`,
				)
			}
			if strings.HasPrefix(expression, "resource.column_name == ") {
				raw["column"] = strings.TrimSuffix(
					strings.TrimPrefix(expression, `resource.column_name == "`),
					`"`,
				)
			}
			if strings.HasPrefix(expression, "request.time < ") {
				raw["expire_timestamp"] = strings.TrimSuffix(
					strings.TrimPrefix(expression, `request.time < timestamp("`),
					`")`,
				)
			}
		}
		if instanceID == "" || databaseName == "" {
			return nil, errors.Errorf("invalid exception policy condition: %v", exception.Condition.Expression)
		}
		raw["database"] = fmt.Sprintf("%s%s/%s%s", internal.InstanceNamePrefix, instanceID, internal.DatabaseIDPrefix, databaseName)
		exceptionList = append(exceptionList, raw)
	}
	policy := map[string]interface{}{
		"exceptions": schema.NewSet(exceptionHash, exceptionList),
	}
	return []interface{}{policy}, nil
}

func exceptionHash(rawException interface{}) int {
	var buf bytes.Buffer
	exception := rawException.(map[string]interface{})

	if v, ok := exception["database"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := exception["schema"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := exception["table"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := exception["column"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := exception["member"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := exception["action"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := exception["expire_timestamp"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}

	return internal.ToHashcodeInt(buf.String())
}
