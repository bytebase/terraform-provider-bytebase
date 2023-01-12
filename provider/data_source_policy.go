package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

var (
	deploymentApprovalPolicySchema = &schema.Schema{
		Computed: true,
		Optional: true,
		Type:     schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"default_strategy": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"deployment_approval_strategies": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"approval_group": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"approval_strategy": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"deployment_type": {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
			},
		},
	}

	backupPlanPolicySchema = &schema.Schema{
		Computed: true,
		Optional: true,
		Type:     schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"schedule": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"retention_duration": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}

	sensitiveDataPolicy = &schema.Schema{
		Computed: true,
		Optional: true,
		Type:     schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"sensitive_data": {
					Computed: true,
					Type:     schema.TypeList,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"schema": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"table": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"column": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"mask_type": {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
			},
		},
	}

	accessControlPolicy = &schema.Schema{
		Computed: true,
		Optional: true,
		Type:     schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"disallow_rules": {
					Computed: true,
					Type:     schema.TypeList,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"full_database": {
								Type:     schema.TypeBool,
								Computed: true,
							},
						},
					},
				},
			},
		},
	}

	sqlReviewPolicy = &schema.Schema{
		Computed: true,
		Optional: true,
		Type:     schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"title": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"rules": {
					Computed: true,
					Type:     schema.TypeList,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"type": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"level": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"payload": {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
			},
		},
	}
)

func dataSourcePolicy() *schema.Resource {
	return &schema.Resource{
		Description: "The policy data source.",
		ReadContext: dataSourcePolicyRead,
		Schema: map[string]*schema.Schema{
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
			"deployment_approval_policy": deploymentApprovalPolicySchema,
			"backup_plan_policy":         backupPlanPolicySchema,
			"sensitive_data_policy":      sensitiveDataPolicy,
			"access_control_policy":      accessControlPolicy,
			"sql_review_policy":          sqlReviewPolicy,
		},
	}
}

func dataSourcePolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	find, err := getPolicyFind(d)
	if err != nil {
		return diag.FromErr(err)
	}

	policyType := api.PolicyType(d.Get("type").(string))
	find.Type = &policyType

	policy, err := c.GetPolicy(ctx, find)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(policy.Name)
	return setPolicyMessage(d, policy)
}

func getPolicyFind(d *schema.ResourceData) (*api.PolicyFindMessage, error) {
	projectID := d.Get("project").(string)
	environmentID := d.Get("environment").(string)
	if projectID != "" && environmentID != "" {
		return nil, errors.Errorf("cannot set both project and environment")
	}

	find := &api.PolicyFindMessage{}
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
		if err := d.Set("backup_plan_policy", flattenBackupPlanPolicy(p)); err != nil {
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
		if err := d.Set("sql_review_policy", flattenSQLReviewPolicy(p)); err != nil {
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

func flattenBackupPlanPolicy(p *api.BackupPlanPolicy) []interface{} {
	policy := map[string]interface{}{
		"schedule":           p.Schedule,
		"retention_duration": p.RetentionDuration,
	}
	return []interface{}{policy}
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
		raw["full_database"] = rule.FullDatabase
		rules = append(rules, raw)
	}
	policy := map[string]interface{}{
		"disallow_rules": rules,
	}
	return []interface{}{policy}
}

func flattenSQLReviewPolicy(p *api.SQLReviewPolicy) []interface{} {
	rules := []interface{}{}
	for _, rule := range p.Rules {
		raw := map[string]interface{}{}
		raw["type"] = rule.Type
		raw["level"] = rule.Level
		raw["payload"] = rule.Payload
		rules = append(rules, raw)
	}

	policy := map[string]interface{}{
		"title": p.Title,
		"rules": rules,
	}
	return []interface{}{policy}
}
