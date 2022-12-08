package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func resourceInstance() *schema.Resource {
	return &schema.Resource{
		Description:   "The instance resource.",
		CreateContext: resourceInstanceCreate,
		ReadContext:   resourceInstanceRead,
		UpdateContext: resourceInstanceUpdate,
		DeleteContext: resourceInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The instance name.",
			},
			"engine": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"MYSQL",
					"POSTGRES",
					"TIDB",
					"SNOWFLAKE",
					"CLICKHOUSE",
				}, false),
				Description: "The instance engine. Support MYSQL, POSTGRES, TIDB, SNOWFLAKE, CLICKHOUSE.",
			},
			"engine_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version for instance engine.",
			},
			"external_link": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The external console URL managing this instance (e.g. AWS RDS console, your in-house DB instance console)",
			},
			"host": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "Host or socket for your instance, or the account name if the instance type is Snowflake.",
			},
			"port": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The port for your instance.",
			},
			"database": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The database for your instance.",
			},
			"environment": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The environment name for your instance.",
			},
			"data_source_list": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 3,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The data source unique id",
						},
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The unique data source name in this instance.",
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"ADMIN",
								"RO",
							}, false),
							Description: "The data source type. Should be ADMIN or RO.",
						},
						"username": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The connection user name used by Bytebase to perform DDL and DML operations.",
						},
						"password": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The connection user password used by Bytebase to perform DDL and DML operations.",
						},
						"ssl_ca": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The CA certificate. Optional, you can set this if the engine type is CLICKHOUSE.",
						},
						"ssl_cert": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The client certificate. Optional, you can set this if the engine type is CLICKHOUSE.",
						},
						"ssl_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The client key. Optional, you can set this if the engine type is CLICKHOUSE.",
						},
						"host_override": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The Read-replica Host. Only works for RO type data source",
						},
						"port_override": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The Read-replica Port. Only works for RO type data source",
						},
					},
				},
			},
		},
	}
}

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	instance, err := c.CreateInstance(ctx, &api.InstanceCreate{
		Environment:    d.Get("environment").(string),
		Name:           d.Get("name").(string),
		Engine:         d.Get("engine").(string),
		ExternalLink:   d.Get("external_link").(string),
		Database:       d.Get("database").(string),
		Host:           d.Get("host").(string),
		Port:           d.Get("port").(string),
		DataSourceList: convertDataSourceCreateList(d),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(instance.ID))

	return resourceInstanceRead(ctx, d, m)
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	instanceID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	instance, err := c.GetInstance(ctx, instanceID)
	if err != nil {
		return diag.FromErr(err)
	}

	return setInstance(d, instance)
}

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	instanceID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	patch := &api.InstancePatch{}
	if d.HasChange("name") {
		if name, ok := d.GetOk("name"); ok {
			val := name.(string)
			patch.Name = &val
		}
	}
	if d.HasChange("external_link") {
		if link, ok := d.GetOk("external_link"); ok {
			val := link.(string)
			patch.ExternalLink = &val
		}
	}
	if d.HasChange("host") {
		if host, ok := d.GetOk("host"); ok {
			val := host.(string)
			patch.Host = &val
		}
	}
	if d.HasChange("port") {
		if port, ok := d.GetOk("port"); ok {
			val := port.(string)
			patch.Port = &val
		}
	}
	if d.HasChange("database") {
		if database, ok := d.GetOk("database"); ok {
			val := database.(string)
			patch.Database = &val
		}
	}
	if d.HasChange("data_source_list") {
		patch.DataSourceList = convertDataSourceCreateList(d)
	}

	if patch.HasChange() {
		if _, err := c.UpdateInstance(ctx, instanceID, patch); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceInstanceRead(ctx, d, m)
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	instanceID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := c.DeleteInstance(ctx, instanceID); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func setInstance(d *schema.ResourceData, instance *api.Instance) diag.Diagnostics {
	if err := d.Set("name", instance.Name); err != nil {
		return diag.Errorf("cannot set name for instance: %s", err.Error())
	}
	if err := d.Set("engine", instance.Engine); err != nil {
		return diag.Errorf("cannot set engine for instance: %s", err.Error())
	}
	if err := d.Set("engine_version", instance.EngineVersion); err != nil {
		return diag.Errorf("cannot set engine_version for instance: %s", err.Error())
	}
	if err := d.Set("external_link", instance.ExternalLink); err != nil {
		return diag.Errorf("cannot set external_link for instance: %s", err.Error())
	}
	if err := d.Set("host", instance.Host); err != nil {
		return diag.Errorf("cannot set host for instance: %s", err.Error())
	}
	if err := d.Set("port", instance.Port); err != nil {
		return diag.Errorf("cannot set port for instance: %s", err.Error())
	}
	if err := d.Set("environment", instance.Environment); err != nil {
		return diag.Errorf("cannot set environment for instance: %s", err.Error())
	}
	if err := d.Set("data_source_list", flattenDataSourceList(instance.DataSourceList)); err != nil {
		return diag.Errorf("cannot set data_source_list for instance: %s", err.Error())
	}

	return nil
}

func flattenDataSourceList(dataSourceList []*api.DataSource) []interface{} {
	res := []interface{}{}
	for _, dataSource := range dataSourceList {
		raw := map[string]interface{}{}
		raw["id"] = dataSource.ID
		raw["name"] = dataSource.Name
		raw["type"] = dataSource.Type
		raw["username"] = dataSource.Username
		raw["host_override"] = dataSource.HostOverride
		raw["port_override"] = dataSource.PortOverride
		res = append(res, raw)
	}
	return res
}

func convertDataSourceCreateList(d *schema.ResourceData) []*api.DataSourceCreate {
	var dataSourceList []*api.DataSourceCreate
	if rawList, ok := d.Get("data_source_list").([]interface{}); ok {
		for _, raw := range rawList {
			obj := raw.(map[string]interface{})
			dataSource := &api.DataSourceCreate{
				Name: obj["name"].(string),
				Type: obj["type"].(string),
			}

			if v, ok := obj["username"].(string); ok {
				dataSource.Username = v
			}
			if v, ok := obj["password"].(string); ok && v != "" {
				dataSource.Password = v
			}
			if v, ok := obj["ssl_ca"].(string); ok {
				dataSource.SslCa = v
			}
			if v, ok := obj["ssl_cert"].(string); ok {
				dataSource.SslCert = v
			}
			if v, ok := obj["ssl_key"].(string); ok {
				dataSource.SslKey = v
			}
			if v, ok := obj["host_override"].(string); ok {
				dataSource.HostOverride = v
			}
			if v, ok := obj["port_override"].(string); ok {
				dataSource.PortOverride = v
			}
			dataSourceList = append(dataSourceList, dataSource)
		}
	}

	return dataSourceList
}
