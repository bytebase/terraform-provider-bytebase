package provider

import (
	"context"
	"fmt"

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
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance full name in instances/{resource id} format.",
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
			"engine_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The engine version.",
			},
			"external_link": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The external console URL managing this instance (e.g. AWS RDS console, your in-house DB instance console)",
			},
			"sync_interval": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "How often the instance is synced in seconds. Default 0, means never sync.",
			},
			"maximum_connections": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The maximum number of connections. The default value is 10.",
			},
			"data_sources": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique data source id in this instance.",
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
						"password": {
							Type:        schema.TypeString,
							Computed:    true,
							Sensitive:   true,
							Description: "The connection user password used by Bytebase to perform DDL and DML operations.",
						},
						"ssl_ca": {
							Type:        schema.TypeString,
							Computed:    true,
							Sensitive:   true,
							Description: "The CA certificate. Optional, you can set this if the engine type is MYSQL, POSTGRES, TIDB or CLICKHOUSE.",
						},
						"ssl_cert": {
							Type:        schema.TypeString,
							Computed:    true,
							Sensitive:   true,
							Description: "The client certificate. Optional, you can set this if the engine type is MYSQL, POSTGRES, TIDB or CLICKHOUSE.",
						},
						"ssl_key": {
							Type:        schema.TypeString,
							Computed:    true,
							Sensitive:   true,
							Description: "The client key. Optional, you can set this if the engine type is MYSQL, POSTGRES, TIDB or CLICKHOUSE.",
						},
					},
				},
				Set: dataSourceHash,
			},
		},
	}
}

func dataSourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	instanceName := fmt.Sprintf("%s%s", internal.InstanceNamePrefix, d.Get("resource_id").(string))

	ins, err := c.GetInstance(ctx, instanceName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ins.Name)

	return setInstanceMessage(d, ins)
}
