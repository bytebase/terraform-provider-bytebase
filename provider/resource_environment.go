package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func resourceEnvironment() *schema.Resource {
	return &schema.Resource{
		Description:   "The environment resource.",
		CreateContext: resourceEnvironmentCreate,
		ReadContext:   resourceEnvironmentRead,
		UpdateContext: resourceEnvironmentUpdate,
		DeleteContext: resourceEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The environment unique name.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"order": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "The environment sorting order.",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"environment_tier_policy": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"PROTECTED",
					"UNPROTECTED",
				}, false),
				Description: "If marked as PROTECTED, developers cannot execute any query on this environment's databases using SQL Editor by default.",
			},
			"pipeline_approval_policy": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(api.PipelineApprovalTypeNever),
					string(api.PipelineApprovalTypeByProjectOwner),
					string(api.PipelineApprovalTypeByWorkspaceOwnerORDBA),
				}, false),
				Description: "For updating schema on the existing database, this setting controls whether the task requires manual approval.",
			},
			"backup_plan_policy": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"UNSET",
					"DAILY",
					"WEEKLY",
				}, false),
				Description: "The database backup policy in this environment.",
			},
		},
	}
}

func resourceEnvironmentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name, ok := d.Get("name").(string)
	if !ok {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get the environment name",
			Detail:   "The environment name is required for creation",
		})
		return diags
	}

	order, ok := d.Get("order").(int)
	if !ok {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get the environment order",
			Detail:   "The environment order is required for creation",
		})
		return diags
	}

	pipelineApprovalPolicy, err := convertPipelineApprovalPolicy(d)
	if err != nil {
		return diag.Errorf("Invalid pipeline approval policy: %v", err.Error())
	}

	create := &api.EnvironmentUpsert{
		Name:                   &name,
		Order:                  &order,
		EnvironmentTierPolicy:  convertEnvironmentTierPolicy(d),
		PipelineApprovalPolicy: pipelineApprovalPolicy,
		BackupPlanPolicy:       convertBackupPlanPolicy(d),
	}

	env, err := c.CreateEnvironment(ctx, create)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(env.ID))

	return resourceEnvironmentRead(ctx, d, m)
}

func resourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	envID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	env, err := c.GetEnvironment(ctx, envID)
	if err != nil {
		return diag.FromErr(err)
	}

	return setEnvironment(d, env)
}

func resourceEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	envID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	patch := &api.EnvironmentUpsert{}
	if d.HasChange("name") {
		name, ok := d.Get("name").(string)
		if ok {
			patch.Name = &name
		}
	}

	if d.HasChange("order") {
		order, ok := d.Get("order").(int)
		if ok {
			patch.Order = &order
		}
	}

	if d.HasChange("environment_tier_policy") {
		patch.EnvironmentTierPolicy = convertEnvironmentTierPolicy(d)
	}
	if d.HasChange("pipeline_approval_policy") {
		pipelineApprovalPolicy, err := convertPipelineApprovalPolicy(d)
		if err != nil {
			return diag.Errorf("Invalid pipeline approval policy: %v", err.Error())
		}
		patch.PipelineApprovalPolicy = pipelineApprovalPolicy
	}
	if d.HasChange("backup_plan_policy") {
		patch.BackupPlanPolicy = convertBackupPlanPolicy(d)
	}

	if patch.HasChange() {
		if _, err := c.UpdateEnvironment(ctx, envID, patch); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceEnvironmentRead(ctx, d, m)
}

func resourceEnvironmentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	envID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := c.DeleteEnvironment(ctx, envID); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func setEnvironment(d *schema.ResourceData, env *api.Environment) diag.Diagnostics {
	if err := d.Set("name", env.Name); err != nil {
		return diag.Errorf("cannot set name for environment: %s", err.Error())
	}
	if err := d.Set("order", env.Order); err != nil {
		return diag.Errorf("cannot set order for environment: %s", err.Error())
	}
	if err := d.Set("environment_tier_policy", env.EnvironmentTierPolicy.EnvironmentTier); err != nil {
		return diag.Errorf("cannot set environment_tier_policy for environment: %s", err.Error())
	}
	if err := d.Set("pipeline_approval_policy", flattenPipelineApprovalPolicy(env.PipelineApprovalPolicy)); err != nil {
		return diag.Errorf("cannot set pipeline_approval_policy for environment: %s", err.Error())
	}
	if err := d.Set("backup_plan_policy", env.BackupPlanPolicy.Schedule); err != nil {
		return diag.Errorf("cannot set backup_plan_policy for environment: %s", err.Error())
	}

	return nil
}

func flattenPipelineApprovalPolicy(policy *api.PipelineApprovalPolicy) string {
	approvalType := api.PipelineApprovalTypeNever

	if len(policy.AssigneeGroupList) > 0 {
		switch policy.AssigneeGroupList[0].Value {
		case api.AssigneeGroupValueProjectOwner:
			approvalType = api.PipelineApprovalTypeByProjectOwner
		case api.AssigneeGroupValueWorkspaceOwnerOrDBA:
			approvalType = api.PipelineApprovalTypeByWorkspaceOwnerORDBA
		}
	}

	return string(approvalType)
}

func convertEnvironmentTierPolicy(d *schema.ResourceData) *api.EnvironmentTierPolicy {
	var policy *api.EnvironmentTierPolicy

	if v, ok := d.GetOk("environment_tier_policy"); ok {
		policy = &api.EnvironmentTierPolicy{
			EnvironmentTier: v.(string),
		}
	}

	return policy
}

func convertPipelineApprovalPolicy(d *schema.ResourceData) (*api.PipelineApprovalPolicy, error) {
	var policy *api.PipelineApprovalPolicy

	if v, ok := d.GetOk("pipeline_approval_policy"); ok {
		assigneeGroupList := []api.AssigneeGroup{}
		pipelineApprovalValue := api.PipelineApprovalValueManualNever

		switch api.PipelineApprovalType(v.(string)) {
		case api.PipelineApprovalTypeByProjectOwner:
			pipelineApprovalValue = api.PipelineApprovalValueManualAlways
			assigneeGroupList = []api.AssigneeGroup{
				{
					IssueType: api.IssueDatabaseDataUpdate,
					Value:     api.AssigneeGroupValueProjectOwner,
				},
				{
					IssueType: api.IssueDatabaseSchemaUpdate,
					Value:     api.AssigneeGroupValueProjectOwner,
				},
				{
					IssueType: api.IssueDatabaseSchemaUpdateGhost,
					Value:     api.AssigneeGroupValueProjectOwner,
				},
			}
		case api.PipelineApprovalTypeByWorkspaceOwnerORDBA:
			pipelineApprovalValue = api.PipelineApprovalValueManualAlways
			assigneeGroupList = []api.AssigneeGroup{
				{
					IssueType: api.IssueDatabaseDataUpdate,
					Value:     api.AssigneeGroupValueWorkspaceOwnerOrDBA,
				},
				{
					IssueType: api.IssueDatabaseSchemaUpdate,
					Value:     api.AssigneeGroupValueWorkspaceOwnerOrDBA,
				},
				{
					IssueType: api.IssueDatabaseSchemaUpdateGhost,
					Value:     api.AssigneeGroupValueWorkspaceOwnerOrDBA,
				},
			}
		}

		policy = &api.PipelineApprovalPolicy{
			Value:             pipelineApprovalValue,
			AssigneeGroupList: assigneeGroupList,
		}
	}

	return policy, nil
}

func convertBackupPlanPolicy(d *schema.ResourceData) *api.BackupPlanPolicy {
	var policy *api.BackupPlanPolicy

	if v, ok := d.GetOk("backup_plan_policy"); ok {
		policy = &api.BackupPlanPolicy{
			Schedule: v.(string),
		}
	}

	return policy
}
