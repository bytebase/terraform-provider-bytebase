package provider

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

// policyParentIdentificationMap is the map to identify a policy's parent.
var policyParentIdentificationMap = map[string]*schema.Schema{
	"project": {
		Type:         schema.TypeString,
		Optional:     true,
		Default:      "",
		ValidateFunc: internal.ResourceIDValidation,
		Description:  "The project resource id for the policy.",
	},
	"environment": {
		Type:         schema.TypeString,
		Optional:     true,
		Default:      "",
		ValidateFunc: internal.ResourceIDValidation,
		Description:  "The environment resource id for the policy.",
	},
	"instance": {
		Type:         schema.TypeString,
		Optional:     true,
		Default:      "",
		ValidateFunc: internal.ResourceIDValidation,
		Description:  "The instance resource id for the policy.",
	},
	"database": {
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "",
		Description: "The database name for the policy.",
	},
}

func dataSourcePolicy() *schema.Resource {
	return &schema.Resource{
		Description: "The policy data source.",
		ReadContext: dataSourcePolicyRead,
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
				Computed:    true,
				Description: "Decide if the policy should inherit from the parent.",
			},
			"deployment_approval_policy": getDeploymentApprovalPolicySchema(true),
			"backup_plan_policy":         getBackupPlanPolicySchema(true),
			"sensitive_data_policy":      getSensitiveDataPolicy(true),
			"access_control_policy":      getAccessControlPolicy(true),
			"sql_review_policy":          getSQLReviewPolicy(true),
		}),
	}
}

func dataSourcePolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	find, err := getPolicyFind(ctx, d, c)
	if err != nil {
		return diag.FromErr(err)
	}

	policy, err := c.GetPolicy(ctx, find)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(policy.Name)
	return setPolicyMessage(d, policy)
}

func getPolicyFind(ctx context.Context, d *schema.ResourceData, client api.Client) (*api.PolicyFindMessage, error) {
	projectID := d.Get("project").(string)
	environmentID := d.Get("environment").(string)
	if projectID != "" && environmentID != "" {
		return nil, errors.Errorf("cannot set both project and environment")
	}

	find := &api.PolicyFindMessage{}

	pType, ok := d.Get("type").(string)
	if ok {
		policyType := api.PolicyType(pType)
		find.Type = &policyType
	}

	if projectID != "" {
		find.ProjectID = &projectID
	} else if environmentID != "" {
		find.EnvironmentID = &environmentID

		if v := d.Get("instance").(string); v != "" {
			if find.EnvironmentID == nil {
				return nil, errors.Errorf("must set both environment and instance to find the instance policy")
			}
			find.InstanceID = &v
		}
		if v := d.Get("database").(string); v != "" {
			if find.EnvironmentID == nil || find.InstanceID == nil {
				return nil, errors.Errorf("must set both environment, instance and database to find the database policy")
			}
			find.DatabaseName = &v
		}
	}

	// Make sure the parent for the policy exists
	if find.DatabaseName != nil {
		if _, err := client.GetDatabase(ctx, &api.DatabaseFindMessage{
			EnvironmentID: *find.EnvironmentID,
			InstanceID:    *find.InstanceID,
			DatabaseName:  *find.DatabaseName,
		}); err != nil {
			return nil, errors.Errorf(
				"cannot find the database: environments/%s/instances/%s/databases/%s with error %v, please make sure the database synchronization is done",
				*find.EnvironmentID,
				*find.InstanceID,
				*find.DatabaseName,
				err.Error(),
			)
		}
	} else if find.InstanceID != nil {
		if _, err := client.GetInstance(ctx, &api.InstanceFindMessage{
			EnvironmentID: *find.EnvironmentID,
			InstanceID:    *find.InstanceID,
		}); err != nil {
			return nil, errors.Errorf(
				"cannot find the instance: environments/%s/instances/%s with error %v",
				*find.EnvironmentID,
				*find.InstanceID,
				err.Error(),
			)
		}
	} else if find.EnvironmentID != nil {
		if _, err := client.GetEnvironment(ctx, *find.EnvironmentID); err != nil {
			return nil, errors.Errorf(
				"cannot find the instance: environments/%s with error %v",
				*find.EnvironmentID,
				err.Error(),
			)
		}
	} else if find.ProjectID != nil {
		if _, err := client.GetProject(ctx, *find.ProjectID, false /* showDeleted */); err != nil {
			return nil, errors.Errorf(
				"cannot find the project: projects/%s with error %v",
				*find.ProjectID,
				err.Error(),
			)
		}
	}

	return find, nil
}

func setPolicyMessage(d *schema.ResourceData, policy *api.PolicyMessage) diag.Diagnostics {
	if err := d.Set("name", policy.Name); err != nil {
		return diag.Errorf("cannot set name for policy: %s", err.Error())
	}
	if err := d.Set("inherit_from_parent", policy.InheritFromParent); err != nil {
		return diag.Errorf("cannot set inherit_from_parent for policy: %s", err.Error())
	}

	if p := policy.DeploymentApprovalPolicy; p != nil {
		if err := d.Set("deployment_approval_policy", flattenDeploymentApprovalPolicy(p)); err != nil {
			return diag.Errorf("cannot set deployment_approval_policy: %s", err.Error())
		}
	}

	if p := policy.BackupPlanPolicy; p != nil {
		backupPlan, err := flattenBackupPlanPolicy(p)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("backup_plan_policy", backupPlan); err != nil {
			return diag.Errorf("cannot set backup_plan_policy: %s", err.Error())
		}
	}

	if p := policy.SensitiveDataPolicy; p != nil {
		if err := d.Set("sensitive_data_policy", flattenSensitiveDataPolicy(p)); err != nil {
			return diag.Errorf("cannot set sensitive_data_policy: %s", err.Error())
		}
	}

	if p := policy.AccessControlPolicy; p != nil {
		if err := d.Set("access_control_policy", flattenAccessControlPolicy(p)); err != nil {
			return diag.Errorf("cannot set access_control_policy: %s", err.Error())
		}
	}

	if p := policy.SQLReviewPolicy; p != nil {
		sqlReviewPolicy, err := flattenSQLReviewPolicy(p)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("sql_review_policy", sqlReviewPolicy); err != nil {
			return diag.Errorf("cannot set sql_review_policy: %s", err.Error())
		}
	}

	return nil
}

func flattenDeploymentApprovalPolicy(p *api.DeploymentApprovalPolicy) []interface{} {
	strategies := []interface{}{}
	for _, strategy := range p.DeploymentApprovalStrategies {
		raw := map[string]interface{}{}
		raw["approval_group"] = strategy.ApprovalGroup
		raw["approval_strategy"] = strategy.ApprovalStrategy
		raw["deployment_type"] = strategy.DeploymentType
		strategies = append(strategies, raw)
	}
	policy := map[string]interface{}{
		"default_strategy":               p.DefaultStrategy,
		"deployment_approval_strategies": strategies,
	}

	return []interface{}{policy}
}

func flattenBackupPlanPolicy(p *api.BackupPlanPolicy) ([]interface{}, error) {
	duration := p.RetentionDuration
	if strings.HasSuffix(duration, "s") {
		duration = duration[:(len(duration) - 1)]
	}
	d, err := strconv.Atoi(duration)
	if err != nil {
		return nil, err
	}

	policy := map[string]interface{}{
		"schedule":           p.Schedule,
		"retention_duration": d,
	}
	return []interface{}{policy}, nil
}

func flattenSensitiveDataPolicy(p *api.SensitiveDataPolicy) []interface{} {
	sensitiveDataList := []interface{}{}
	for _, data := range p.SensitiveData {
		raw := map[string]interface{}{}
		raw["schema"] = data.Schema
		raw["table"] = data.Table
		raw["column"] = data.Column
		raw["mask_type"] = data.MaskType
		sensitiveDataList = append(sensitiveDataList, raw)
	}
	policy := map[string]interface{}{
		"sensitive_data": sensitiveDataList,
	}
	return []interface{}{policy}
}

func flattenAccessControlPolicy(p *api.AccessControlPolicy) []interface{} {
	rules := []interface{}{}
	for _, rule := range p.DisallowRules {
		raw := map[string]interface{}{}
		raw["all_databases"] = rule.FullDatabase
		rules = append(rules, raw)
	}
	policy := map[string]interface{}{
		"disallow_rules": rules,
	}
	return []interface{}{policy}
}

func flattenSQLReviewPolicy(p *api.SQLReviewPolicy) ([]interface{}, error) {
	rules := []interface{}{}
	for _, rule := range p.Rules {
		raw := map[string]interface{}{}
		raw["type"] = rule.Type
		raw["level"] = rule.Level
		raw["engine"] = rule.Engine

		payload, err := unamrshalSQLReviewRulePayload(rule.Type, rule.Payload)
		if err != nil {
			return nil, err
		}
		payloadRaw := map[string]interface{}{}
		for key, val := range payload {
			payloadRaw[key] = val
		}
		raw["payload"] = []interface{}{payloadRaw}
		rules = append(rules, raw)
	}

	policy := map[string]interface{}{
		"title": p.Title,
		"rules": rules,
	}
	return []interface{}{policy}, nil
}

func getDeploymentApprovalPolicySchema(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed: computed,
		Optional: true,
		Default:  nil,
		Type:     schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"default_strategy": {
					Type:     schema.TypeString,
					Computed: computed,
					Optional: true,
					ValidateFunc: validation.StringInSlice([]string{
						string(api.ApprovalStrategyManual),
						string(api.ApprovalStrategyAutomatic),
					}, false),
				},
				"deployment_approval_strategies": {
					Type:     schema.TypeList,
					Computed: computed,
					Optional: true,
					MinItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"approval_group": {
								Type:     schema.TypeString,
								Computed: computed,
								Optional: true,
								ValidateFunc: validation.StringInSlice([]string{
									string(api.ApprovalGroupDBA),
									string(api.ApprovalGroupOwner),
								}, false),
							},
							"approval_strategy": {
								Type:     schema.TypeString,
								Computed: computed,
								Optional: true,
								ValidateFunc: validation.StringInSlice([]string{
									string(api.ApprovalStrategyManual),
									string(api.ApprovalStrategyAutomatic),
								}, false),
							},
							"deployment_type": {
								Type:     schema.TypeString,
								Computed: computed,
								Optional: true,
								ValidateFunc: validation.StringInSlice([]string{
									string(api.DeploymentTypeDatabaseCreate),
									string(api.DeploymentTypeDatabaseDDL),
									string(api.DeploymentTypeDatabaseDDLGhost),
									string(api.DeploymentTypeDatabaseDML),
									string(api.DeploymentTypeDatabaseRestorePITR),
									string(api.DeploymentTypeDatabaseDMLRollback),
								}, false),
							},
						},
					},
				},
			},
		},
	}
}

func getBackupPlanPolicySchema(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed: computed,
		Optional: true,
		Default:  nil,
		Type:     schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"schedule": {
					Type:     schema.TypeString,
					Computed: computed,
					Optional: true,
					ValidateFunc: validation.StringInSlice([]string{
						string(api.BackupPlanScheduleUnset),
						string(api.BackupPlanScheduleDaily),
						string(api.BackupPlanScheduleWeekly),
					}, false),
				},
				"retention_duration": {
					Type:        schema.TypeInt,
					Computed:    computed,
					Optional:    true,
					Description: "The minimum allowed seconds that backup data is kept for databases in an environment.",
				},
			},
		},
	}
}

func getSensitiveDataPolicy(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed: computed,
		Optional: true,
		Default:  nil,
		Type:     schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"sensitive_data": {
					Computed: computed,
					Optional: true,
					Type:     schema.TypeList,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"schema": {
								Type:     schema.TypeString,
								Computed: computed,
								Optional: true,
							},
							"table": {
								Type:     schema.TypeString,
								Computed: computed,
								Optional: true,
							},
							"column": {
								Type:     schema.TypeString,
								Computed: computed,
								Optional: true,
							},
							"mask_type": {
								Type:     schema.TypeString,
								Computed: computed,
								Optional: true,
								ValidateFunc: validation.StringInSlice([]string{
									string(api.SensitiveDataMaskTypeDefault),
								}, false),
							},
						},
					},
				},
			},
		},
	}
}

func getAccessControlPolicy(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed: computed,
		Optional: true,
		Default:  nil,
		Type:     schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"disallow_rules": {
					Computed: computed,
					Optional: true,
					Type:     schema.TypeList,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"all_databases": {
								Type:     schema.TypeBool,
								Computed: computed,
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
}

func getSQLReviewPolicy(computed bool) *schema.Schema {
	return &schema.Schema{
		Computed: computed,
		Optional: true,
		Default:  nil,
		Type:     schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"title": {
					Type:     schema.TypeString,
					Computed: computed,
					Optional: true,
				},
				"rules": {
					Computed:    computed,
					Optional:    true,
					Type:        schema.TypeList,
					Description: "The SQL review rules. Please check the doc for details: https://www.bytebase.com/docs/sql-review/review-policy/supported-rules",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"type": {
								Type:     schema.TypeString,
								Computed: computed,
								Optional: true,
								ValidateFunc: validation.StringInSlice([]string{
									string(api.SchemaRuleTableNaming),
									string(api.SchemaRuleColumnNaming),
									string(api.SchemaRulePKNaming),
									string(api.SchemaRuleUKNaming),
									string(api.SchemaRuleFKNaming),
									string(api.SchemaRuleIDXNaming),
									string(api.SchemaRuleAutoIncrementColumnNaming),
									string(api.SchemaRuleStatementNoSelectAll),
									string(api.SchemaRuleStatementRequireWhere),
									string(api.SchemaRuleStatementNoLeadingWildcardLike),
									string(api.SchemaRuleStatementDisallowCommit),
									string(api.SchemaRuleStatementDisallowLimit),
									string(api.SchemaRuleStatementDisallowOrderBy),
									string(api.SchemaRuleStatementMergeAlterTable),
									string(api.SchemaRuleStatementInsertRowLimit),
									string(api.SchemaRuleStatementInsertMustSpecifyColumn),
									string(api.SchemaRuleStatementInsertDisallowOrderByRand),
									string(api.SchemaRuleStatementAffectedRowLimit),
									string(api.SchemaRuleStatementDMLDryRun),
									string(api.SchemaRuleTableRequirePK),
									string(api.SchemaRuleTableNoFK),
									string(api.SchemaRuleTableDropNamingConvention),
									string(api.SchemaRuleTableCommentConvention),
									string(api.SchemaRuleTableDisallowPartition),
									string(api.SchemaRuleRequiredColumn),
									string(api.SchemaRuleColumnNotNull),
									string(api.SchemaRuleColumnDisallowChangeType),
									string(api.SchemaRuleColumnSetDefaultForNotNull),
									string(api.SchemaRuleColumnDisallowChange),
									string(api.SchemaRuleColumnDisallowChangingOrder),
									string(api.SchemaRuleColumnCommentConvention),
									string(api.SchemaRuleColumnAutoIncrementMustInteger),
									string(api.SchemaRuleColumnTypeDisallowList),
									string(api.SchemaRuleColumnDisallowSetCharset),
									string(api.SchemaRuleColumnMaximumCharacterLength),
									string(api.SchemaRuleColumnAutoIncrementInitialValue),
									string(api.SchemaRuleColumnAutoIncrementMustUnsigned),
									string(api.SchemaRuleCurrentTimeColumnCountLimit),
									string(api.SchemaRuleColumnRequireDefault),
									string(api.SchemaRuleSchemaBackwardCompatibility),
									string(api.SchemaRuleDropEmptyDatabase),
									string(api.SchemaRuleIndexNoDuplicateColumn),
									string(api.SchemaRuleIndexKeyNumberLimit),
									string(api.SchemaRuleIndexPKTypeLimit),
									string(api.SchemaRuleIndexTypeNoBlob),
									string(api.SchemaRuleIndexTotalNumberLimit),
									string(api.SchemaRuleIndexPrimaryKeyTypeAllowlist),
									string(api.SchemaRuleCharsetAllowlist),
									string(api.SchemaRuleCollationAllowlist),
									string(api.SchemaRuleCommentLength),
								}, false),
							},
							"level": {
								Type:     schema.TypeString,
								Computed: computed,
								Optional: true,
								ValidateFunc: validation.StringInSlice([]string{
									string(api.SQLReviewRuleLevelError),
									string(api.SQLReviewRuleLevelWarning),
									string(api.SQLReviewRuleLevelDisabled),
								}, false),
							},
							"engine": {
								Type:     schema.TypeString,
								Computed: computed,
								Optional: true,
								ValidateFunc: validation.StringInSlice([]string{
									string(api.EngineTypeMySQL),
									string(api.EngineTypePostgres),
									string(api.EngineTypeTiDB),
								}, false),
								Description: "The engine for this rule.",
							},
							"payload": {
								Computed: computed,
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 0,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"number": {
											Type:     schema.TypeInt,
											Computed: computed,
											Optional: true,
										},
										"format": {
											Type:     schema.TypeString,
											Computed: computed,
											Optional: true,
										},
										"required": {
											Type:     schema.TypeBool,
											Computed: computed,
											Optional: true,
										},
										"max_length": {
											Type:     schema.TypeInt,
											Computed: computed,
											Optional: true,
										},
										"list": {
											Type:     schema.TypeList,
											Computed: computed,
											Optional: true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func getPolicySchema(policySchemaMap map[string]*schema.Schema) map[string]*schema.Schema {
	for key, val := range policyParentIdentificationMap {
		policySchemaMap[key] = val
	}
	return policySchemaMap
}

func unamrshalSQLReviewRulePayload(ruleType api.SQLReviewRuleType, payload string) (map[string]interface{}, error) {
	switch ruleType {
	case
		api.SchemaRuleTableNaming,
		api.SchemaRuleColumnNaming,
		api.SchemaRuleAutoIncrementColumnNaming,
		api.SchemaRuleFKNaming,
		api.SchemaRuleIDXNaming,
		api.SchemaRuleUKNaming:
		var p api.NamingRulePayload
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"max_length": p.MaxLength,
			"format":     p.Format,
		}, nil
	case
		api.SchemaRuleColumnCommentConvention,
		api.SchemaRuleTableCommentConvention:
		var p api.CommentConventionRulePayload
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"max_length": p.MaxLength,
			"required":   p.Required,
		}, nil
	case
		api.SchemaRuleIndexKeyNumberLimit,
		api.SchemaRuleStatementInsertRowLimit,
		api.SchemaRuleIndexTotalNumberLimit,
		api.SchemaRuleColumnMaximumCharacterLength,
		api.SchemaRuleColumnAutoIncrementInitialValue,
		api.SchemaRuleStatementAffectedRowLimit:
		var p api.NumberTypeRulePayload
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"number": p.Number,
		}, nil
	case
		api.SchemaRuleRequiredColumn,
		api.SchemaRuleColumnTypeDisallowList,
		api.SchemaRuleCharsetAllowlist,
		api.SchemaRuleCollationAllowlist,
		api.SchemaRuleIndexPrimaryKeyTypeAllowlist:
		var p api.StringArrayTypeRulePayload
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"list": p.List,
		}, nil
	}

	return map[string]interface{}{}, nil
}
