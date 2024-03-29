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

func dataSourceInstanceList() *schema.Resource {
	return &schema.Resource{
		Description: "The instance data source list.",
		ReadContext: dataSourceInstanceListRead,
		Schema: map[string]*schema.Schema{
			"show_deleted": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Including removed instance in the response.",
			},
			"instances": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The instance unique resource id.",
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
				},
			},
		},
	}
}

func dataSourceInstanceListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	response, err := c.ListInstance(ctx, &api.InstanceFindMessage{
		ShowDeleted: d.Get("show_deleted").(bool),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	instances := make([]map[string]interface{}, 0)
	for _, instance := range response.Instances {
		instanceID, err := internal.GetInstanceID(instance.Name)
		if err != nil {
			return diag.FromErr(err)
		}

		ins := make(map[string]interface{})
		ins["resource_id"] = instanceID
		ins["title"] = instance.Title
		ins["engine"] = instance.Engine
		ins["external_link"] = instance.ExternalLink
		ins["data_sources"] = flattenDataSourceList(instance.DataSources)

		envID, err := internal.GetEnvironmentID(instance.Environment)
		if err != nil {
			return diag.FromErr(err)
		}
		ins["environment"] = envID

		instances = append(instances, ins)
	}

	if err := d.Set("instances", instances); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
