package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceEnvironmentList() *schema.Resource {
	return &schema.Resource{
		Description: "The environment data source list.",
		ReadContext: dataSourceEnvironmentListRead,
		Schema: map[string]*schema.Schema{
			"show_deleted": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Including removed instance in the response.",
			},
			"environments": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The environment unique resource id.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The environment full name in environments/{resource id} format.",
						},
						"title": {
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

	response, err := c.ListEnvironment(ctx, d.Get("show_deleted").(bool))
	if err != nil {
		return diag.FromErr(err)
	}

	environments := []map[string]interface{}{}
	for _, environment := range response.Environments {
		envID, err := internal.GetEnvironmentID(environment.Name)
		if err != nil {
			return diag.FromErr(err)
		}

		env := make(map[string]interface{})
		env["resource_id"] = envID
		env["title"] = environment.Title
		env["name"] = environment.Name
		env["order"] = environment.Order
		env["environment_tier_policy"] = string(environment.Tier)

		environments = append(environments, env)
	}

	if err := d.Set("environments", environments); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
