package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

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
				Optional: true,
				Default:  "",
				ValidateFunc: internal.ResourceNameValidation(
					// workspace policy
					regexp.MustCompile("^$"),
					// environment policy
					regexp.MustCompile(fmt.Sprintf("^%s%s$", internal.EnvironmentNamePrefix, internal.ResourceIDPattern)),
					// instance policy
					regexp.MustCompile(fmt.Sprintf("^%s%s$", internal.InstanceNamePrefix, internal.ResourceIDPattern)),
					// project policy
					regexp.MustCompile(fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern)),
					// database policy
					regexp.MustCompile(fmt.Sprintf("^%s%s%s%s$", internal.InstanceNamePrefix, internal.ResourceIDPattern, internal.DatabaseIDPrefix, internal.ResourceIDPattern)),
				),
				Description: "The policy parent name for the policy, support projects/{resource id}, environments/{resource id}, instances/{resource id}, or instances/{resource id}/databases/{database name}",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(api.PolicyTypeDeploymentApproval),
					string(api.PolicyTypeBackupPlan),
					string(api.PolicyTypeSensitiveData),
					string(api.PolicyTypeAccessControl),
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
			"deployment_approval_policy": getDeploymentApprovalPolicySchema(true),
			"backup_plan_policy":         getBackupPlanPolicySchema(true),
			"sensitive_data_policy":      getSensitiveDataPolicy(true),
			"access_control_policy":      getAccessControlPolicy(true),
		},
	}
}

func dataSourcePolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	policyName := fmt.Sprintf("%s/%s%s", d.Get("parent").(string), internal.PolicyNamePrefix, d.Get("type").(string))
	policy, err := c.GetPolicy(ctx, policyName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(policy.Name)
	return setPolicyMessage(d, policy)
}

func setPolicyMessage(d *schema.ResourceData, policy *api.PolicyMessage) diag.Diagnostics {
	parent, _, err := internal.GetPolicyParentAndType(policy.Name)
	if err != nil {
		return diag.Errorf("cannot parse name for policy: %s", err.Error())
	}
	if err := d.Set("name", policy.Name); err != nil {
		return diag.Errorf("cannot set name for policy: %s", err.Error())
	}
	if err := d.Set("parent", parent); err != nil {
		return diag.Errorf("cannot set parent for policy: %s", err.Error())
	}
	if err := d.Set("inherit_from_parent", policy.InheritFromParent); err != nil {
		return diag.Errorf("cannot set inherit_from_parent for policy: %s", err.Error())
	}
	if err := d.Set("enforce", policy.Enforce); err != nil {
		return diag.Errorf("cannot set enforce for policy: %s", err.Error())
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
