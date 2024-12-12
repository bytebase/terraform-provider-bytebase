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
	"google.golang.org/protobuf/types/known/durationpb"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"

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
					v1pb.Engine_CLICKHOUSE.String(),
					v1pb.Engine_MYSQL.String(),
					v1pb.Engine_POSTGRES.String(),
					v1pb.Engine_SNOWFLAKE.String(),
					v1pb.Engine_SQLITE.String(),
					v1pb.Engine_TIDB.String(),
					v1pb.Engine_MONGODB.String(),
					v1pb.Engine_REDIS.String(),
					v1pb.Engine_ORACLE.String(),
					v1pb.Engine_SPANNER.String(),
					v1pb.Engine_MSSQL.String(),
					v1pb.Engine_REDSHIFT.String(),
					v1pb.Engine_MARIADB.String(),
					v1pb.Engine_OCEANBASE.String(),
					v1pb.Engine_DM.String(),
					v1pb.Engine_RISINGWAVE.String(),
					v1pb.Engine_OCEANBASE_ORACLE.String(),
					v1pb.Engine_STARROCKS.String(),
					v1pb.Engine_DORIS.String(),
					v1pb.Engine_HIVE.String(),
					v1pb.Engine_ELASTICSEARCH.String(),
				}, false),
				Description: "The instance engine. Support MYSQL, POSTGRES, TIDB, SNOWFLAKE, CLICKHOUSE, MONGODB, SQLITE, REDIS, ORACLE, SPANNER, MSSQL, REDSHIFT, MARIADB, OCEANBASE.",
			},
			"engine_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The engine version.",
			},
			"external_link": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The external console URL managing this instance (e.g. AWS RDS console, your in-house DB instance console)",
			},
			"sync_interval": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "How often the instance is synced in seconds. Default 0, means never sync.",
			},
			"maximum_connections": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "The maximum number of connections.",
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
								v1pb.DataSourceType_ADMIN.String(),
								v1pb.DataSourceType_READ_ONLY.String(),
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
							Sensitive:   true,
							Default:     "",
							Description: "The connection user password used by Bytebase to perform DDL and DML operations.",
						},
						"ssl_ca": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Sensitive:   true,
							Description: "The CA certificate. Optional, you can set this if the engine type is MYSQL, POSTGRES, TIDB or CLICKHOUSE.",
						},
						"ssl_cert": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Sensitive:   true,
							Description: "The client certificate. Optional, you can set this if the engine type is MYSQL, POSTGRES, TIDB or CLICKHOUSE.",
						},
						"ssl_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Sensitive:   true,
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

	dataSourceList, err := convertDataSourceCreateList(d, true /* validate */)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID := d.Get("resource_id").(string)
	instanceName := fmt.Sprintf("%s%s", internal.InstanceNamePrefix, instanceID)
	title := d.Get("title").(string)
	externalLink := d.Get("external_link").(string)
	instanceOptions := &v1pb.InstanceOptions{
		SyncInterval: &durationpb.Duration{
			Seconds: int64(d.Get("sync_interval").(int)),
		},
		MaximumConnections: int32(d.Get("maximum_connections").(int)),
	}

	engineString := d.Get("engine").(string)
	engineValue, ok := v1pb.Engine_value[engineString]
	if !ok {
		return diag.Errorf("invalid engine type %v", engineString)
	}
	engine := v1pb.Engine(engineValue)

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

		if existedInstance.Engine != engine {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid argument",
				Detail:   fmt.Sprintf("cannot update instance %s engine to %s", instanceName, engine),
			})
			return diags
		}

		if existedInstance.State == v1pb.State_DELETED {
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

		updateMasks := []string{}
		if title != "" && title != existedInstance.Title {
			updateMasks = append(updateMasks, "title")
		}
		if externalLink != "" && externalLink != existedInstance.ExternalLink {
			updateMasks = append(updateMasks, "external_link")
		}
		if op := existedInstance.Options; op != nil {
			if instanceOptions.SyncInterval.GetSeconds() != op.SyncInterval.GetSeconds() {
				updateMasks = append(updateMasks, "options.sync_interval")
			}
			if instanceOptions.MaximumConnections != op.MaximumConnections {
				updateMasks = append(updateMasks, "options.maximum_connections")
			}
		}
		if len(dataSourceList) > 0 {
			updateMasks = append(updateMasks, "data_sources")
		}

		if len(updateMasks) > 0 {
			if _, err := c.UpdateInstance(ctx, &v1pb.Instance{
				Name:         instanceName,
				Title:        title,
				ExternalLink: externalLink,
				DataSources:  dataSourceList,
				State:        v1pb.State_ACTIVE,
				Options:      instanceOptions,
			}, updateMasks); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to update instance",
					Detail:   fmt.Sprintf("Update instance %s failed, error: %v", instanceName, err),
				})
				return diags
			}
		}
	} else {
		if _, err := c.CreateInstance(ctx, instanceID, &v1pb.Instance{
			Name:         instanceName,
			Title:        d.Get("title").(string),
			Engine:       engine,
			ExternalLink: d.Get("external_link").(string),
			State:        v1pb.State_ACTIVE,
			DataSources:  dataSourceList,
			Environment:  d.Get("environment").(string),
			Options:      instanceOptions,
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
	if existedInstance.State == v1pb.State_DELETED {
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

	dataSourceList, err := convertDataSourceCreateList(d, true /* validate */)
	if err != nil {
		return diag.FromErr(err)
	}

	paths := []string{}
	if d.HasChange("title") {
		paths = append(paths, "title")
	}
	if d.HasChange("external_link") {
		paths = append(paths, "external_link")
	}
	if d.HasChange("data_sources") {
		paths = append(paths, "data_sources")
	}
	if d.HasChange("sync_interval") {
		paths = append(paths, "options.sync_interval")
	}
	if d.HasChange("maximum_connections") {
		paths = append(paths, "options.maximum_connections")
	}

	if len(paths) > 0 {
		if _, err := c.UpdateInstance(ctx, &v1pb.Instance{
			Name:         instanceName,
			Title:        d.Get("title").(string),
			ExternalLink: d.Get("external_link").(string),
			DataSources:  dataSourceList,
			State:        v1pb.State_ACTIVE,
			Options: &v1pb.InstanceOptions{
				SyncInterval: &durationpb.Duration{
					Seconds: int64(d.Get("sync_interval").(int)),
				},
				MaximumConnections: int32(d.Get("maximum_connections").(int)),
			},
		}, paths); err != nil {
			return diag.FromErr(err)
		}
		if err := c.SyncInstanceSchema(ctx, instanceName); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Instance schema sync failed",
				Detail:   fmt.Sprintf("Failed to sync schema for instance %s with error: %v. You can try to trigger the sync manually via Bytebase UI.", instanceName, err.Error()),
			})
		}
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

func setInstanceMessage(d *schema.ResourceData, instance *v1pb.Instance) diag.Diagnostics {
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
	if err := d.Set("engine", instance.Engine.String()); err != nil {
		return diag.Errorf("cannot set engine for instance: %s", err.Error())
	}
	if err := d.Set("engine_version", instance.EngineVersion); err != nil {
		return diag.Errorf("cannot set engine_version for instance: %s", err.Error())
	}
	if err := d.Set("external_link", instance.ExternalLink); err != nil {
		return diag.Errorf("cannot set external_link for instance: %s", err.Error())
	}
	if op := instance.Options; op != nil {
		if err := d.Set("sync_interval", op.SyncInterval.GetSeconds()); err != nil {
			return diag.Errorf("cannot set sync_interval for instance: %s", err.Error())
		}
		if err := d.Set("maximum_connections", op.MaximumConnections); err != nil {
			return diag.Errorf("cannot set maximum_connections for instance: %s", err.Error())
		}
	}

	dataSources, err := flattenDataSourceList(d, instance.DataSources)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("data_sources", dataSources); err != nil {
		return diag.Errorf("cannot set data_sources for instance: %s", err.Error())
	}

	return nil
}

func flattenDataSourceList(d *schema.ResourceData, dataSourceList []*v1pb.DataSource) ([]interface{}, error) {
	oldDataSourceList, err := convertDataSourceCreateList(d, false)
	if err != nil {
		return nil, err
	}
	oldDataSourceMap := make(map[string]*v1pb.DataSource)
	for _, ds := range oldDataSourceList {
		oldDataSourceMap[ds.Id] = ds
	}

	res := []interface{}{}
	for _, dataSource := range dataSourceList {
		raw := map[string]interface{}{}
		raw["id"] = dataSource.Id
		raw["type"] = dataSource.Type.String()
		raw["username"] = dataSource.Username
		raw["host"] = dataSource.Host
		raw["port"] = dataSource.Port
		raw["database"] = dataSource.Database

		// These sensitive fields won't returned in the API. Propagate state value.
		if ds, ok := oldDataSourceMap[dataSource.Id]; ok {
			raw["password"] = ds.Password
			raw["ssl_ca"] = ds.SslCa
			raw["ssl_cert"] = ds.SslCert
			raw["ssl_key"] = ds.SslKey
		}
		res = append(res, raw)
	}
	return res, nil
}

func convertDataSourceCreateList(d *schema.ResourceData, validate bool) ([]*v1pb.DataSource, error) {
	var dataSourceList []*v1pb.DataSource
	if rawList, ok := d.Get("data_sources").([]interface{}); ok {
		dataSourceTypeMap := map[v1pb.DataSourceType]bool{}
		for _, raw := range rawList {
			obj := raw.(map[string]interface{})
			dataSource := &v1pb.DataSource{
				Id:   obj["id"].(string),
				Type: v1pb.DataSourceType(v1pb.DataSourceType_value[obj["type"].(string)]),
			}
			if dataSourceTypeMap[dataSource.Type] && dataSource.Type == v1pb.DataSourceType_ADMIN {
				return nil, errors.Errorf("duplicate data source type ADMIN")
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

		if !dataSourceTypeMap[v1pb.DataSourceType_ADMIN] && validate {
			return nil, errors.Errorf("data source \"%v\" is required", v1pb.DataSourceType_ADMIN.String())
		}
	}

	return dataSourceList, nil
}
