package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceEnvironment() *schema.Resource {
	return &schema.Resource{
		Description: "The environment data source.",
		ReadContext: dataSourceEnvironmentRead,
		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The environment unique resource id.",
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
	}
}

func dataSourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	environment, err := c.GetEnvironment(ctx, d.Get("resource_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(environment.Name)

	return setEnvironment(d, environment)
}
