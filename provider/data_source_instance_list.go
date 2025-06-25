package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceInstanceList() *schema.Resource {
	return &schema.Resource{
		Description:        "The instance data source list.",
		ReadWithoutTimeout: dataSourceInstanceListRead,
		Schema: map[string]*schema.Schema{
			"query": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter instances by name or resource id with wildcard",
			},
			"environment": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The environment full name. Filter instances by environment.",
				ValidateDiagFunc: internal.ResourceNameValidation(
					fmt.Sprintf("^%s%s$", internal.EnvironmentNamePrefix, internal.ResourceIDPattern),
				),
			},
			"project": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The project full name. Filter instances by project.",
				ValidateDiagFunc: internal.ResourceNameValidation(
					fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern),
				),
			},
			"host": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter instances by host.",
			},
			"port": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter instances by port.",
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  v1pb.State_ACTIVE.String(),
				ValidateFunc: validation.StringInSlice([]string{
					v1pb.State_ACTIVE.String(),
					v1pb.State_DELETED.String(),
				}, false),
				Description: "Filter instances by state. Default ACTIVE.",
			},
			"engines": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: internal.EngineValidation,
				},
				Description: "Filter instances by engines.",
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
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The instance full name in instances/{resource id} format.",
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
						"activation": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether assign license for this instance or not.",
						},
						"engine": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The instance engine. Support MYSQL, POSTGRES, TIDB, SNOWFLAKE, CLICKHOUSE, MONGODB, SQLITE, REDIS, ORACLE, SPANNER, MSSQL, REDSHIFT, MARIADB, OCEANBASE, COCKROACHDB.",
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
										Description: "The data source type. Should be ADMIN or READ_ONLY.",
									},
									"username": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The connection user name used by Bytebase to perform DDL and DML operations.",
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
										Description: "The database for the instance, you can set this if the engine type is POSTGRES.",
									},
									"password": {
										Type:        schema.TypeString,
										Computed:    true,
										Sensitive:   true,
										Description: "The connection user password used by Bytebase to perform DDL and DML operations.",
									},
									"external_secret": getExternalSecretSchema(),
									"use_ssl": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Enable SSL connection. Required to use SSL certificates.",
									},
									"ssl_ca": {
										Type:        schema.TypeString,
										Computed:    true,
										Sensitive:   true,
										Description: "The CA certificate. Optional, you can set this if the engine type is MYSQL, POSTGRES, TIDB, CLICKHOUSE or COCKROACHDB.",
									},
									"ssl_cert": {
										Type:        schema.TypeString,
										Computed:    true,
										Sensitive:   true,
										Description: "The client certificate. Optional, you can set this if the engine type is MYSQL, POSTGRES, TIDB, CLICKHOUSE or COCKROACHDB.",
									},
									"ssl_key": {
										Type:        schema.TypeString,
										Computed:    true,
										Sensitive:   true,
										Description: "The client key. Optional, you can set this if the engine type is MYSQL, POSTGRES, TIDB, CLICKHOUSE or COCKROACHDB.",
									},
								},
							},
							Set: dataSourceHash,
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

	filter := &api.InstanceFilter{
		Query:       d.Get("query").(string),
		Environment: d.Get("environment").(string),
		Project:     d.Get("project").(string),
		Host:        d.Get("host").(string),
		Port:        d.Get("port").(string),
	}
	stateString := d.Get("state").(string)
	stateValue, ok := v1pb.State_value[stateString]
	if ok {
		filter.State = v1pb.State(stateValue)
	}

	engines := d.Get("engines").(*schema.Set)
	for _, engine := range engines.List() {
		engineString := engine.(string)
		engineValue, ok := v1pb.Engine_value[engineString]
		if ok {
			filter.Engines = append(filter.Engines, v1pb.Engine(engineValue))
		}
	}

	response, err := c.ListInstance(ctx, filter)
	if err != nil {
		return diag.FromErr(err)
	}

	instances := make([]map[string]interface{}, 0)
	for _, instance := range response {
		instanceID, err := internal.GetInstanceID(instance.Name)
		if err != nil {
			return diag.FromErr(err)
		}

		ins := make(map[string]interface{})
		ins["resource_id"] = instanceID
		ins["title"] = instance.Title
		ins["name"] = instance.Name
		ins["activation"] = instance.Activation
		ins["engine"] = instance.Engine.String()
		ins["engine_version"] = instance.EngineVersion
		ins["external_link"] = instance.ExternalLink
		ins["environment"] = instance.Environment
		ins["sync_interval"] = instance.GetSyncInterval().GetSeconds()
		ins["maximum_connections"] = instance.GetMaximumConnections()

		dataSources, err := flattenDataSourceList(d, instance.DataSources)
		if err != nil {
			return diag.FromErr(err)
		}
		ins["data_sources"] = schema.NewSet(dataSourceHash, dataSources)

		instances = append(instances, ins)
	}

	if err := d.Set("instances", instances); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
