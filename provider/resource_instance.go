package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
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
			"resource_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The instance unique resource id.",
			},
			"environment": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: internal.ResourceNameValidation(regexp.MustCompile(fmt.Sprintf("^%s%s$", internal.EnvironmentNamePrefix, internal.ResourceIDPattern))),
				Description:      "The environment full name for the instance in environments/{environment id} format.",
			},
			"title": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The instance title.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance full name in instances/{resource id} format.",
			},
			"engine": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(api.EngineTypeMySQL),
					string(api.EngineTypePostgres),
					string(api.EngineTypeTiDB),
					string(api.EngineTypeSnowflake),
					string(api.EngineTypeClickHouse),
					string(api.EngineTypeMongoDB),
					string(api.EngineTypeSQLite),
					string(api.EngineTypeRedis),
					string(api.EngineTypeOracle),
					string(api.EngineTypeSpanner),
					string(api.EngineTypeMSSQL),
					string(api.EngineTypeRedshift),
					string(api.EngineTypeMariaDB),
					string(api.EngineTypeOceanbase),
				}, false),
				Description: "The instance engine. Support MYSQL, POSTGRES, TIDB, SNOWFLAKE, CLICKHOUSE, MONGODB, SQLITE, REDIS, ORACLE, SPANNER, MSSQL, REDSHIFT, MARIADB, OCEANBASE.",
			},
			"external_link": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The external console URL managing this instance (e.g. AWS RDS console, your in-house DB instance console)",
			},
			"data_sources": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Description: "The connection for the instance. You can configure read-only or admin connection account here.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The unique data source id in this instance.",
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(api.DataSourceAdmin),
								string(api.DataSourceRO),
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
							Description: "The CA certificate. Optional, you can set this if the engine type is MYSQL, POSTGRES, TIDB or CLICKHOUSE.",
						},
						"ssl_cert": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The client certificate. Optional, you can set this if the engine type is MYSQL, POSTGRES, TIDB or CLICKHOUSE.",
						},
						"ssl_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The client key. Optional, you can set this if the engine type is MYSQL, POSTGRES, TIDB or CLICKHOUSE.",
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
							Description: "The database for the instance, you can set this if the engine type is POSTGRES.",
						},
					},
				},
			},
		},
	}
}

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	dataSourceList, err := convertDataSourceCreateList(d)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID := d.Get("resource_id").(string)
	instanceName := fmt.Sprintf("%s%s", internal.InstanceNamePrefix, instanceID)

	existedInstance, err := c.GetInstance(ctx, instanceName)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get instance %s failed with error: %v", instanceName, err))
	}

	var diags diag.Diagnostics
	if existedInstance != nil && err == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Instance already exists",
			Detail:   fmt.Sprintf("Instance %s already exists, try to exec the update operation", instanceName),
		})

		engine := d.Get("engine").(string)
		if string(existedInstance.Engine) != engine {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid argument",
				Detail:   fmt.Sprintf("cannot update instance %s engine to %s", instanceName, engine),
			})
			return diags
		}

		if existedInstance.State == api.Deleted {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Instance is deleted",
				Detail:   fmt.Sprintf("Instance %s already deleted, try to undelete the instance", instanceName),
			})
			if _, err := c.UndeleteInstance(ctx, instanceName); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to undelete instance",
					Detail:   fmt.Sprintf("Undelete instance %s failed, error: %v", instanceName, err),
				})
				return diags
			}
		}

		title := d.Get("title").(string)
		externalLink := d.Get("external_link").(string)
		if _, err := c.UpdateInstance(ctx, &api.InstancePatchMessage{
			Name:         instanceName,
			Title:        &title,
			ExternalLink: &externalLink,
			DataSources:  dataSourceList,
		}); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update instance",
				Detail:   fmt.Sprintf("Update instance %s failed, error: %v", instanceName, err),
			})
			return diags
		}
	} else {
		if _, err := c.CreateInstance(ctx, instanceID, &api.InstanceMessage{
			Name:         instanceName,
			Title:        d.Get("title").(string),
			Engine:       api.EngineType(d.Get("engine").(string)),
			ExternalLink: d.Get("external_link").(string),
			State:        api.Active,
			DataSources:  dataSourceList,
			Environment:  d.Get("environment").(string),
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := c.SyncInstanceSchema(ctx, instanceName); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Instance schema sync failed",
			Detail:   fmt.Sprintf("Failed to sync schema for instance %s with error: %v. You can try to trigger the sync manually via Bytebase UI.", instanceName, err.Error()),
		})
	}
	d.SetId(instanceName)

	diag := resourceInstanceRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	instanceName := d.Id()

	instance, err := c.GetInstance(ctx, instanceName)
	if err != nil {
		return diag.FromErr(err)
	}

	return setInstanceMessage(d, instance)
}

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("resource_id") {
		return diag.Errorf("cannot change the resource id")
	}
	if d.HasChange("environment") {
		return diag.Errorf("cannot change the environment in instance")
	}
	if d.HasChange("engine") {
		return diag.Errorf("cannot change the engine in instance")
	}

	c := m.(api.Client)
	instanceName := d.Id()

	existedInstance, err := c.GetInstance(ctx, instanceName)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get instance %s failed with error: %v", instanceName, err))
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	if existedInstance.State == api.Deleted {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Instance is deleted",
			Detail:   fmt.Sprintf("Instance %s already deleted, try to undelete the instance", instanceName),
		})
		if _, err := c.UndeleteInstance(ctx, instanceName); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to undelete instance",
				Detail:   fmt.Sprintf("Undelete instance %s failed, error: %v", instanceName, err),
			})
			return diags
		}
	}

	patch := &api.InstancePatchMessage{
		Name: instanceName,
	}
	if d.HasChange("title") {
		v := d.Get("title").(string)
		patch.Title = &v
	}
	if d.HasChange("external_link") {
		v := d.Get("external_link").(string)
		patch.ExternalLink = &v
	}
	if d.HasChange("data_sources") {
		dataSourceList, err := convertDataSourceCreateList(d)
		if err != nil {
			return diag.FromErr(err)
		}
		patch.DataSources = dataSourceList
	}

	if _, err := c.UpdateInstance(ctx, patch); err != nil {
		return diag.FromErr(err)
	}
	if err := c.SyncInstanceSchema(ctx, instanceName); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Instance schema sync failed",
			Detail:   fmt.Sprintf("Failed to sync schema for instance %s with error: %v. You can try to trigger the sync manually via Bytebase UI.", instanceName, err.Error()),
		})
	}

	diag := resourceInstanceRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	instanceName := d.Id()

	if err := c.DeleteInstance(ctx, instanceName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func setInstanceMessage(d *schema.ResourceData, instance *api.InstanceMessage) diag.Diagnostics {
	instanceID, err := internal.GetInstanceID(instance.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resource_id", instanceID); err != nil {
		return diag.Errorf("cannot set resource_id for instance: %s", err.Error())
	}
	if err := d.Set("title", instance.Title); err != nil {
		return diag.Errorf("cannot set title for instance: %s", err.Error())
	}
	if err := d.Set("name", instance.Name); err != nil {
		return diag.Errorf("cannot set name for instance: %s", err.Error())
	}
	if err := d.Set("environment", instance.Environment); err != nil {
		return diag.Errorf("cannot set environment for instance: %s", err.Error())
	}
	if err := d.Set("engine", instance.Engine); err != nil {
		return diag.Errorf("cannot set engine for instance: %s", err.Error())
	}
	if err := d.Set("external_link", instance.ExternalLink); err != nil {
		return diag.Errorf("cannot set external_link for instance: %s", err.Error())
	}
	if err := d.Set("data_sources", flattenDataSourceList(instance.DataSources)); err != nil {
		return diag.Errorf("cannot set data_sources for instance: %s", err.Error())
	}

	return nil
}

func flattenDataSourceList(dataSourceList []*api.DataSourceMessage) []interface{} {
	res := []interface{}{}
	for _, dataSource := range dataSourceList {
		raw := map[string]interface{}{}
		raw["id"] = dataSource.ID
		raw["type"] = dataSource.Type
		raw["username"] = dataSource.Username
		raw["host"] = dataSource.Host
		raw["port"] = dataSource.Port
		res = append(res, raw)
	}
	return res
}

func convertDataSourceCreateList(d *schema.ResourceData) ([]*api.DataSourceMessage, error) {
	var dataSourceList []*api.DataSourceMessage
	if rawList, ok := d.Get("data_sources").([]interface{}); ok {
		dataSourceTypeMap := map[api.DataSourceType]bool{}
		for _, raw := range rawList {
			obj := raw.(map[string]interface{})
			dataSource := &api.DataSourceMessage{
				ID:   obj["id"].(string),
				Type: api.DataSourceType(obj["type"].(string)),
			}

			if dataSourceTypeMap[dataSource.Type] {
				return nil, errors.Errorf("duplicate data source type %s", dataSource.Type)
			}
			dataSourceTypeMap[dataSource.Type] = true

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
			if v, ok := obj["host"].(string); ok {
				dataSource.Host = v
			}
			if v, ok := obj["port"].(string); ok {
				dataSource.Port = v
			}
			if v, ok := obj["database"].(string); ok {
				dataSource.Database = v
			}
			dataSourceList = append(dataSourceList, dataSource)
		}

		if !dataSourceTypeMap[api.DataSourceAdmin] {
			return nil, errors.Errorf("data source \"%v\" is required", api.DataSourceAdmin)
		}
	}

	return dataSourceList, nil
}
