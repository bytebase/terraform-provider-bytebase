package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"

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
					fmt.Sprintf(`^%s%s/%s\S+$`, internal.InstanceNamePrefix, internal.ResourceIDPattern, internal.DatabaseIDPrefix),
				),
				Description: "The policy parent name for the policy, support projects/{resource id}, environments/{resource id}, instances/{resource id}, or instances/{resource id}/databases/{database name}",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					v1pb.PolicyType_MASKING_EXEMPTION.String(),
					v1pb.PolicyType_MASKING_RULE.String(),
					v1pb.PolicyType_DATA_SOURCE_QUERY.String(),
					v1pb.PolicyType_ROLLOUT_POLICY.String(),
					v1pb.PolicyType_DATA_QUERY.String(),
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
			"masking_exemption_policy": getMaskingExemptionPolicySchema(true),
			"global_masking_policy":    getGlobalMaskingPolicySchema(true),
			"data_source_query_policy": getDataSourceQueryPolicySchema(true),
			"rollout_policy":           getRolloutPolicySchema(true),
			"query_data_policy":        getDataQueryPolicySchema(true),
		},
	}
}

func getMaskingExemptionPolicySchema(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed: computed,
		Optional: true,
		Default:  nil,
		Type:     schema.TypeList,
		MinItems: 0,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"exemptions": {
					Computed: computed,
					Optional: true,
					Default:  nil,
					MinItems: 0,
					Type:     schema.TypeSet,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"database": {
								Type:     schema.TypeString,
								Computed: computed,
								Optional: true,
								ValidateDiagFunc: internal.ResourceNameValidation(
									// database name format
									fmt.Sprintf(`^%s%s/%s\S+$`, internal.InstanceNamePrefix, internal.ResourceIDPattern, internal.DatabaseIDPrefix),
								),
								Description: "The database full name in instances/{instance resource id}/databases/{database name} format",
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
							"columns": {
								Type:     schema.TypeSet,
								Computed: computed,
								Optional: true,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: validation.StringIsNotEmpty,
								},
							},
							"members": {
								Type:     schema.TypeSet,
								Required: true,
								MinItems: 1,
								Elem: &schema.Schema{
									Type:        schema.TypeString,
									Description: "The member in user:{email} or group:{email} format.",
									ValidateDiagFunc: internal.ResourceNameValidation(
										"^user:",
										"^group:",
									),
								},
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
							"raw_expression": {
								Type:        schema.TypeString,
								Computed:    computed,
								Optional:    true,
								Description: `The raw CEL expression. We will use it as the masking exemption and ignore the "database"/"schema"/"table"/"columns"/"expire_timestamp" fields if you provide the raw expression.`,
							},
						},
					},
					Set: exemptionHash,
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

func getDataQueryPolicySchema(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed:    computed,
		Optional:    true,
		Default:     nil,
		Type:        schema.TypeList,
		MaxItems:    1,
		MinItems:    1,
		Description: "The policy for query data",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"maximum_result_size": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     100 * 1024 * 1024,
					Description: "The size limit in bytes. The default value is 100MB, we will use the default value if the limit <= 0.",
				},
				"maximum_result_rows": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     -1,
					Description: "The return rows limit. If the value <= 0, will be treated as no limit. The default value is -1.",
				},
				"disable_export": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Disable export data in the SQL editor",
				},
				"disable_copy_data": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Disable copying data in the SQL editor",
				},
				"timeout_in_seconds": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "The maximum time allowed for a query to run in SQL Editor. No limit when the value <= 0",
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
						Description: `Role full name in roles/{id} format.`,
						ValidateDiagFunc: internal.ResourceNameValidation(
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
	if err := d.Set("name", policy.Name); err != nil {
		return diag.Errorf("cannot set name for policy: %s", err.Error())
	}
	if err := d.Set("inherit_from_parent", policy.InheritFromParent); err != nil {
		return diag.Errorf("cannot set inherit_from_parent for policy: %s", err.Error())
	}
	if err := d.Set("enforce", policy.Enforce); err != nil {
		return diag.Errorf("cannot set enforce for policy: %s", err.Error())
	}

	key, payload, diags := flattenPolicyPayload(policy)
	if diags != nil {
		return diags
	}
	if err := d.Set(key, payload); err != nil {
		return diag.Errorf("cannot set %s for policy: %s", key, err.Error())
	}

	return nil
}

func flattenPolicyPayload(policy *v1pb.Policy) (string, interface{}, diag.Diagnostics) {
	_, policyType, err := internal.GetPolicyParentAndType(policy.Name)
	if err != nil {
		return "", nil, diag.Errorf("cannot parse name for policy: %s", err.Error())
	}
	switch policyType {
	case v1pb.PolicyType_MASKING_EXEMPTION:
		if p := policy.GetMaskingExemptionPolicy(); p != nil {
			exemptionPolicy, err := flattenMaskingExemptionPolicy(p)
			if err != nil {
				return "", nil, diag.FromErr(err)
			}
			return "masking_exemption_policy", exemptionPolicy, nil
		}
	case v1pb.PolicyType_MASKING_RULE:
		if p := policy.GetMaskingRulePolicy(); p != nil {
			maskingPolicy, err := flattenGlobalMaskingPolicy(p)
			if err != nil {
				return "", nil, diag.FromErr(err)
			}
			return "global_masking_policy", maskingPolicy, nil
		}
	case v1pb.PolicyType_DATA_SOURCE_QUERY:
		if p := policy.GetDataSourceQueryPolicy(); p != nil {
			dataSourceQueryPolicy := flattenDataSourceQueryPolicy(p)
			return "data_source_query_policy", dataSourceQueryPolicy, nil
		}
	case v1pb.PolicyType_ROLLOUT_POLICY:
		if p := policy.GetRolloutPolicy(); p != nil {
			rolloutPolicy := flattenRolloutPolicy(p)
			return "rollout_policy", rolloutPolicy, nil
		}
	case v1pb.PolicyType_DATA_QUERY:
		if p := policy.GetQueryDataPolicy(); p != nil {
			rolloutPolicy := flattenQueryDataPolicy(p)
			return "query_data_policy", rolloutPolicy, nil
		}
	}

	return "", nil, diag.Errorf("unsupported policy: %s", policy.Name)
}

func flattenRolloutPolicy(p *v1pb.RolloutPolicy) []interface{} {
	roles := []string{}
	roles = append(roles, p.Roles...)
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

func flattenQueryDataPolicy(p *v1pb.QueryDataPolicy) []interface{} {
	policy := map[string]interface{}{
		"maximum_result_size": int(p.MaximumResultSize),
		"maximum_result_rows": int(p.MaximumResultRows),
		"disable_export":      p.DisableExport,
		"disable_copy_data":   p.DisableCopyData,
		"timeout_in_seconds":  int(p.Timeout.Seconds),
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

func flattenMaskingExemptionPolicy(p *v1pb.MaskingExemptionPolicy) ([]interface{}, error) {
	exemptionList := []interface{}{}

	for _, exemption := range p.Exemptions {
		if exemption.Condition == nil || exemption.Condition.Expression == "" {
			// Skip invalid data.
			continue
		}

		// The new API uses Condition.Title for the reason/description
		reason := exemption.Condition.Title
		if reason == "" {
			reason = exemption.Condition.Description
		}

		// Convert members to interface slice for schema.Set
		members := []interface{}{}
		for _, member := range exemption.Members {
			members = append(members, member)
		}

		raw := map[string]interface{}{
			"members":        schema.NewSet(schema.HashString, members),
			"reason":         reason,
			"raw_expression": exemption.Condition.Expression,
		}

		expressions := strings.Split(exemption.Condition.Expression, " && ")
		instanceID := ""
		databaseName := ""
		columns := []interface{}{}

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
				columns = append(columns, strings.TrimSuffix(
					strings.TrimPrefix(expression, `resource.column_name == "`),
					`"`,
				))
			}
			if strings.HasPrefix(expression, "request.time < ") {
				raw["expire_timestamp"] = strings.TrimSuffix(
					strings.TrimPrefix(expression, `request.time < timestamp("`),
					`")`,
				)
			}
			if strings.HasPrefix(expression, "resource.column_name in [") {
				// rawColumnListString should be: "col1", "col2"
				rawColumnListString := strings.TrimSuffix(
					strings.TrimPrefix(expression, `resource.column_name in [`),
					`]`,
				)
				rawColumnList := strings.SplitSeq(rawColumnListString, ",")
				for rawColumn := range rawColumnList {
					column := strings.TrimSuffix(
						strings.TrimPrefix(strings.TrimSpace(rawColumn), `"`),
						`"`,
					)
					columns = append(columns, column)
				}
			}
		}
		if instanceID != "" && databaseName != "" {
			raw["database"] = fmt.Sprintf("%s%s/%s%s", internal.InstanceNamePrefix, instanceID, internal.DatabaseIDPrefix, databaseName)
		}
		if len(columns) > 0 {
			raw["columns"] = schema.NewSet(schema.HashString, columns)
		}
		exemptionList = append(exemptionList, raw)
	}

	policy := map[string]interface{}{
		"exemptions": exemptionList,
	}
	return []interface{}{policy}, nil
}

func exemptionHash(rawSchema interface{}) int {
	exemptions, err := convertToV1Exemptions(rawSchema)
	if err != nil {
		return 0
	}
	return internal.ToHash(&v1pb.MaskingExemptionPolicy{
		Exemptions: exemptions,
	})
}
