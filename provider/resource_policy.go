package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

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
			"deployment_approval_policy": getDeploymentApprovalPolicySchema(false),
			"backup_plan_policy":         getBackupPlanPolicySchema(false),
			"sensitive_data_policy":      getSensitiveDataPolicy(false),
			"access_control_policy":      getAccessControlPolicy(false),
		},
	}
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	parent := fmt.Sprintf("%s/%s%s", d.Get("parent").(string), internal.PolicyNamePrefix, d.Get("type").(string))
	policy, err := c.GetPolicy(ctx, parent)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(policy.Name)
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

	inheritFromParent := d.Get("inherit_from_parent").(bool)
	enforce := d.Get("enforce").(bool)

	patch := &api.PolicyPatchMessage{
		Name:              policyName,
		InheritFromParent: &inheritFromParent,
		Enforce:           &enforce,
	}

	policyType := api.PolicyType(strings.ToUpper(d.Get("type").(string)))
	switch policyType {
	case api.PolicyTypeDeploymentApproval:
		if _, ok := d.GetOk("deployment_approval_policy"); ok {
			policy, err := convertDeploymentApprovalPolicy(d)
			if err != nil {
				return diag.FromErr(err)
			}
			patch.DeploymentApprovalPolicy = policy
		}
	case api.PolicyTypeBackupPlan:
		if _, ok := d.GetOk("backup_plan_policy"); ok {
			policy, err := convertBackupPlanPolicy(d)
			if err != nil {
				return diag.FromErr(err)
			}
			patch.BackupPlanPolicy = policy
		}
	case api.PolicyTypeSensitiveData:
		if _, ok := d.GetOk("sensitive_data_policy"); ok {
			policy, err := convertSensitiveDataPolicy(d)
			if err != nil {
				return diag.FromErr(err)
			}
			patch.SensitiveDataPolicy = policy
		}
	case api.PolicyTypeAccessControl:
		policy, err := convertAccessControlPolicy(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.AccessControlPolicy = policy
	}

	if err := validatePolicy(policyType, patch); err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	p, err := c.UpsertPolicy(ctx, patch)
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

	patch := &api.PolicyPatchMessage{
		Name: policyName,
	}

	if d.HasChange("inherit_from_parent") {
		v := d.Get("inherit_from_parent").(bool)
		patch.InheritFromParent = &v
	}
	if d.HasChange("enforce") {
		v := d.Get("enforce").(bool)
		patch.Enforce = &v
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

	policyType := api.PolicyType(strings.ToUpper(d.Get("type").(string)))
	if err := validatePolicy(policyType, patch); err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	if _, err := c.UpsertPolicy(ctx, patch); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to upsert policy",
			Detail:   fmt.Sprintf("Upsert policy %s failed, error: %v", policyName, err),
		})
		return diags
	}

	diag := resourcePolicyRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
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
	if !ok || len(rawList) > 1 {
		return nil, errors.Errorf("invalid access_control_policy")
	}

	if len(rawList) == 0 {
		if _, ok := d.GetOk("database"); !ok {
			return nil, errors.Errorf("access_control_policy is required")
		}

		return &api.AccessControlPolicy{
			DisallowRules: []*api.AccessControlRule{
				{
					FullDatabase: false,
				},
			},
		}, nil
	}

	raw := rawList[0].(map[string]interface{})
	rules := raw["disallow_rules"].([]interface{})
	policy := &api.AccessControlPolicy{}

	for _, rule := range rules {
		rawRule := rule.(map[string]interface{})
		policy.DisallowRules = append(policy.DisallowRules, &api.AccessControlRule{
			FullDatabase: rawRule["all_databases"].(bool),
		})
	}

	return policy, nil
}

func validatePolicy(policyType api.PolicyType, patch *api.PolicyPatchMessage) error {
	switch policyType {
	case api.PolicyTypeDeploymentApproval:
		if patch.DeploymentApprovalPolicy == nil {
			return errors.Errorf("must set deployment_approval_policy for %v policy", policyType)
		}
	case api.PolicyTypeBackupPlan:
		if patch.BackupPlanPolicy == nil {
			return errors.Errorf("must set backup_plan_policy for %v policy", policyType)
		}
	case api.PolicyTypeSensitiveData:
		if patch.SensitiveDataPolicy == nil {
			return errors.Errorf("must set sensitive_data_policy for %v policy", policyType)
		}
	case api.PolicyTypeAccessControl:
		if patch.AccessControlPolicy == nil {
			return errors.Errorf("must set access_control_policy for %v policy", policyType)
		}
	}

	return nil
}
