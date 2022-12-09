package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func dataSourceEnvironmentList() *schema.Resource {
	return &schema.Resource{
		Description: "The environment data source list.",
		ReadContext: dataSourceEnvironmentListRead,
		Schema: map[string]*schema.Schema{
			"environments": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The environment id.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The environment unique name.",
						},
						"order": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The environment sorting order.",
						},
						"environment_tier_policy": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "If marked as PROTECTED, developers cannot execute any query on this environment's databases using SQL Editor by default.",
						},
						"pipeline_approval_policy": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "For updating schema on the existing database, this setting controls whether the task requires manual approval.",
						},
						"backup_plan_policy": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The database backup policy in this environment.",
						},
					},
				},
			},
		},
	}
}

func dataSourceEnvironmentListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	environmentList, err := c.ListEnvironment(ctx, &api.EnvironmentFind{})
	if err != nil {
		return diag.FromErr(err)
	}

	environments := []map[string]interface{}{}
	for _, environment := range environmentList {
		env := make(map[string]interface{})
		env["id"] = environment.ID
		env["name"] = environment.Name
		env["order"] = environment.Order
		env["environment_tier_policy"] = environment.EnvironmentTierPolicy.EnvironmentTier
		env["pipeline_approval_policy"] = flattenPipelineApprovalPolicy(environment.PipelineApprovalPolicy)
		env["backup_plan_policy"] = environment.BackupPlanPolicy.Schedule

		environments = append(environments, env)
	}

	if err := d.Set("environments", environments); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
