package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func dataSourceInstanceList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceInstanceRead,
		Schema: map[string]*schema.Schema{
			"instances": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"engine": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"engine_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"external_link": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"host": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"environment": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceInstanceRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	instanceList, err := c.ListInstance()
	if err != nil {
		return diag.FromErr(err)
	}

	instances := make([]map[string]interface{}, 0)
	for _, instance := range instanceList {
		ins := make(map[string]interface{})
		ins["id"] = instance.ID
		ins["name"] = instance.Name
		ins["engine"] = instance.Engine
		ins["engine_version"] = instance.EngineVersion
		ins["external_link"] = instance.ExternalLink
		ins["host"] = instance.Host
		ins["port"] = instance.Port
		ins["username"] = instance.Username
		ins["environment"] = instance.Environment

		instances = append(instances, ins)
	}

	if err := d.Set("instances", instances); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
