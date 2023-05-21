package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceInstance() *schema.Resource {
	return &schema.Resource{
		Description: "The instance data source.",
		ReadContext: dataSourceInstanceRead,
		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The instance unique resource id.",
			},
			"environment": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `The environment name for your instance in "environments/{resource id}" format.`,
			},
			"title": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance title.",
			},
			"engine": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance engine. Support MYSQL, POSTGRES, TIDB, SNOWFLAKE, CLICKHOUSE, MONGODB, SQLITE, REDIS, ORACLE, SPANNER, MSSQL, REDSHIFT, MARIADB, OCEANBASE.",
			},
			"external_link": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The external console URL managing this instance (e.g. AWS RDS console, your in-house DB instance console)",
			},
			"data_sources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"title": {
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
						"host": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Read-replica Host. Only works for RO type data source",
						},
						"port": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Read-replica Port. Only works for RO type data source",
						},
						"database": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The database for the instance, you can set this if the engine type is POSTGRES.",
						},
					},
				},
			},
		},
	}
}

func dataSourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	ins, err := c.GetInstance(ctx, &api.InstanceFindMessage{
		InstanceID: d.Get("resource_id").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ins.Name)

	return setInstanceMessage(d, ins)
}
