package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func dataSourceRisk() *schema.Resource {
	return &schema.Resource{
		Description: "The risk data source.",
		ReadContext: dataSourceRoleRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The risk full name in risks/{uid} format.",
			},
			"title": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The risk title.",
			},
			"source": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The risk source.",
			},
			"level": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The risk level.",
			},
			"active": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "The risk active.",
			},
			"condition": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The risk condition.",
			},
		},
	}
}

func dataSourceRiskRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	riskName := d.Get("name").(string)

	risk, err := c.GetRisk(ctx, riskName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(risk.Name)

	return setRisk(d, risk)
}
