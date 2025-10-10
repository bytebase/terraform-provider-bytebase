package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func dataSourceRiskList() *schema.Resource {
	return &schema.Resource{
		Description: "The risk data source list.",
		ReadContext: dataSourceRiskListRead,
		Schema: map[string]*schema.Schema{
			"risks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The risk full name in risks/{resource id} format.",
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
							Type:        schema.TypeString,
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
				},
			},
		},
	}
}

func dataSourceRiskListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	risks, err := c.ListRisk(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	dataList := []map[string]interface{}{}
	for _, risk := range risks {
		raw := make(map[string]interface{})
		raw["name"] = risk.Name
		raw["title"] = risk.Title
		raw["source"] = risk.Source.String()
		raw["level"] = risk.Level.String()
		raw["active"] = risk.Active
		raw["condition"] = risk.Condition.Expression

		dataList = append(dataList, raw)
	}

	if err := d.Set("risks", dataList); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
