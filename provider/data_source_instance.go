package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func dataSourceInstance() *schema.Resource {
	return &schema.Resource{
		Description: "The instance data source.",
		ReadContext: dataSourceInstanceRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The instance id.",
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The instance name.",
			},
			"engine": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance engine. Support MYSQL, POSTGRES, TIDB, SNOWFLAKE, CLICKHOUSE.",
			},
			"engine_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version for instance engine.",
			},
			"external_link": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The external console URL managing this instance (e.g. AWS RDS console, your in-house DB instance console)",
			},
			"host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Host or socket for your instance, or the account name if the instance type is Snowflake.",
			},
			"port": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The port for your instance.",
			},
			"database": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The database for your instance.",
			},
			"environment": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The environment name for your instance.",
			},
			"data_source_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique data source name in this instance.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The data source type. Should be ADMIN or RO.",
						},
						"username": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The connection user name used by Bytebase to perform DDL and DML operations.",
						},
						"host_override": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Read-replica Host. Only works for RO type data source",
						},
						"port_override": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Read-replica Port. Only works for RO type data source",
						},
					},
				},
			},
		},
	}
}

func dataSourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	name := d.Get("name").(string)
	var diags diag.Diagnostics

	instanceList, err := c.ListInstance(ctx, &api.InstanceFind{
		Name: name,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if len(instanceList) == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get the instance",
			Detail:   fmt.Sprintf("Cannot find the instance %s", name),
		})
		return diags
	}
	if len(instanceList) > 1 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get the instance",
			Detail:   fmt.Sprintf("The instance name is not unique %s", name),
		})
		return diags
	}

	ins := instanceList[0]
	d.SetId(strconv.Itoa(ins.ID))

	return setInstance(d, ins)
}
