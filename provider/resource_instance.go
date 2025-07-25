package provider

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/durationpb"

	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func resourceInstance() *schema.Resource {
	return &schema.Resource{
		Description:          "The instance resource.",
		CreateWithoutTimeout: resourceInstanceCreate,
		ReadWithoutTimeout:   internal.ResourceRead(resourceInstanceRead),
		UpdateContext:        resourceInstanceUpdate,
		DeleteContext:        internal.ResourceDelete,
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
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: internal.ResourceNameValidation(
					fmt.Sprintf("^%s%s$", internal.EnvironmentNamePrefix, internal.ResourceIDPattern),
				),
				Description: "The environment full name for the instance in environments/{environment id} format.",
			},
			"title": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The instance title.",
			},
			"activation": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether assign license for this instance or not.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance full name in instances/{resource id} format.",
			},
			"engine": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: internal.EngineValidation,
				Description:  "The instance engine. Support MYSQL, POSTGRES, TIDB, SNOWFLAKE, CLICKHOUSE, MONGODB, SQLITE, REDIS, ORACLE, SPANNER, MSSQL, REDSHIFT, MARIADB, OCEANBASE, COCKROACHDB.",
			},
			"engine_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The engine version.",
			},
			"external_link": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Default:     nil,
				Description: "The external console URL managing this instance (e.g. AWS RDS console, your in-house DB instance console)",
			},
			"sync_interval": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "How often the instance is synced in seconds. Default 0, means never sync.",
			},
			"maximum_connections": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The maximum number of connections.",
			},
			"sync_databases": getSyncDatabasesSchema(true),
			"data_sources": {
				Type:        schema.TypeSet,
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
							Description: "The data source type. Should be ADMIN or READ_ONLY.",
						},
						"username": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The connection user name used by Bytebase to perform DDL and DML operations.",
						},
						"password": {
							Type:             schema.TypeString,
							Optional:         true,
							Sensitive:        true,
							Computed:         true,
							DiffSuppressFunc: suppressSensitiveFieldDiff,
							Description:      "The connection user password used by Bytebase to perform DDL and DML operations.",
						},
						"external_secret": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "The external secret to get the database password. Learn more: https://www.bytebase.com/docs/get-started/instance/#use-external-secret-manager",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"vault": {
										Type:        schema.TypeList,
										Optional:    true,
										MaxItems:    1,
										Description: "The Valut to get the database password. Reference doc https://developer.hashicorp.com/vault/api-docs/secret/kv/kv-v2",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"url": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The Vault URL.",
												},
												"token": {
													Type:        schema.TypeString,
													Optional:    true,
													Sensitive:   true,
													Description: "The root token without TTL. Learn more: https://developer.hashicorp.com/vault/docs/commands/operator/generate-root",
												},
												"app_role": {
													Type:        schema.TypeList,
													Optional:    true,
													MaxItems:    1,
													Description: "The Vault app role to get the password.",
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"role_id": {
																Type:        schema.TypeString,
																Required:    true,
																Sensitive:   true,
																Description: "The app role id.",
															},
															"secret": {
																Type:        schema.TypeString,
																Required:    true,
																Sensitive:   true,
																Description: "The secret id for the role without ttl.",
															},
															"secret_type": {
																Type:        schema.TypeString,
																Required:    true,
																Description: "The secret id type, can be PLAIN (plain text for the secret) or ENVIRONMENT (envirionment name for the secret).",
																ValidateFunc: validation.StringInSlice([]string{
																	v1pb.DataSourceExternalSecret_AppRoleAuthOption_PLAIN.String(),
																	v1pb.DataSourceExternalSecret_AppRoleAuthOption_ENVIRONMENT.String(),
																}, false),
															},
														},
													},
												},
												"engine_name": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The name for secret engine.",
												},
												"secret_name": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The secret name in the engine to store the password.",
												},
												"password_key_name": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The key name for the password.",
												},
											},
										},
									},
									"aws_secrets_manager": {
										Type:        schema.TypeList,
										Optional:    true,
										MaxItems:    1,
										Description: "The AWS Secrets Manager to get the database password. Reference doc https://docs.aws.amazon.com/secretsmanager/latest/userguide/intro.html",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"secret_name": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The secret name to store the password.",
												},
												"password_key_name": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The key name for the password.",
												},
											},
										},
									},
									"gcp_secret_manager": {
										Type:        schema.TypeList,
										Optional:    true,
										MaxItems:    1,
										Description: "The GCP Secret Manager to get the database password. Reference doc https://cloud.google.com/secret-manager/docs",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"secret_name": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The secret name should be like \"projects/{project-id}/secrets/{secret-id}\".",
												},
											},
										},
									},
								},
							},
						},
						"use_ssl": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Enable SSL connection. Required to use SSL certificates.",
						},
						"ssl_ca": {
							Type:             schema.TypeString,
							Optional:         true,
							Sensitive:        true,
							Computed:         true,
							DiffSuppressFunc: suppressSensitiveFieldDiff,
							Description:      "The CA certificate. Optional, you can set this if the engine type is MYSQL, POSTGRES, TIDB or CLICKHOUSE.",
						},
						"ssl_cert": {
							Type:             schema.TypeString,
							Optional:         true,
							Sensitive:        true,
							Computed:         true,
							DiffSuppressFunc: suppressSensitiveFieldDiff,
							Description:      "The client certificate. Optional, you can set this if the engine type is MYSQL, POSTGRES, TIDB or CLICKHOUSE.",
						},
						"ssl_key": {
							Type:             schema.TypeString,
							Optional:         true,
							Sensitive:        true,
							Computed:         true,
							DiffSuppressFunc: suppressSensitiveFieldDiff,
							Description:      "The client key. Optional, you can set this if the engine type is MYSQL, POSTGRES, TIDB or CLICKHOUSE.",
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
				Set: dataSourceHash,
			},
			"list_all_databases": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "List all databases in this instance. If false, will only list 500 databases.",
			},
			"databases": getDatabasesSchema(true),
		},
	}
}

func getSyncDatabasesSchema(computed bool) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Computed:    computed,
		Optional:    !computed,
		Description: "Enable sync for following databases. Default empty, means sync all schemas & databases.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

// suppressSensitiveFieldDiff suppresses diffs for write-only sensitive fields.
func suppressSensitiveFieldDiff(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	// If the field was previously set (exists in state) and the new value is empty,
	// suppress the diff because the API doesn't return these fields
	if oldValue != "" && newValue == "" {
		return true
	}
	// If both are equal, suppress the diff
	return oldValue == newValue
}

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	dataSourceList, err := convertDataSourceCreateList(d, true /* validate */)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID := d.Get("resource_id").(string)
	instanceName := fmt.Sprintf("%s%s", internal.InstanceNamePrefix, instanceID)
	existedInstance, err := c.GetInstance(ctx, instanceName)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get instance %s failed with error: %v", instanceName, err))
	}

	instance := &v1pb.Instance{
		Name:               instanceName,
		Title:              d.Get("title").(string),
		ExternalLink:       d.Get("external_link").(string),
		DataSources:        dataSourceList,
		Environment:        d.Get("environment").(string),
		Activation:         d.Get("activation").(bool),
		State:              v1pb.State_ACTIVE,
		MaximumConnections: int32(d.Get("maximum_connections").(int)),
		Engine:             v1pb.Engine(v1pb.Engine_value[d.Get("engine").(string)]),
		SyncDatabases:      getSyncDatabases(d),
	}
	rawConfig := d.GetRawConfig()
	if config := rawConfig.GetAttr("sync_interval"); !config.IsNull() {
		instance.SyncInterval = &durationpb.Duration{
			Seconds: int64(d.Get("sync_interval").(int)),
		}
	}

	var diags diag.Diagnostics
	if existedInstance != nil && err == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Instance already exists",
			Detail:   fmt.Sprintf("Instance %s already exists, try to exec the update operation", instanceName),
		})

		if existedInstance.Engine != instance.Engine {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid argument",
				Detail:   fmt.Sprintf("cannot update instance %s engine to %s", instanceName, instance.Engine),
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
		if instance.Title != existedInstance.Title {
			updateMasks = append(updateMasks, "title")
		}
		if config := rawConfig.GetAttr("external_link"); !config.IsNull() && instance.ExternalLink != existedInstance.ExternalLink {
			updateMasks = append(updateMasks, "external_link")
		}
		if instance.Environment != existedInstance.Environment {
			updateMasks = append(updateMasks, "environment")
		}
		if config := rawConfig.GetAttr("activation"); !config.IsNull() && instance.Activation != existedInstance.Activation {
			updateMasks = append(updateMasks, "activation")
		}
		if config := rawConfig.GetAttr("sync_interval"); !config.IsNull() && instance.SyncInterval.GetSeconds() != existedInstance.GetSyncInterval().GetSeconds() {
			updateMasks = append(updateMasks, "sync_interval")
		}
		if config := rawConfig.GetAttr("maximum_connections"); !config.IsNull() && instance.MaximumConnections != existedInstance.GetMaximumConnections() {
			updateMasks = append(updateMasks, "maximum_connections")
		}
		if config := rawConfig.GetAttr("sync_databases"); !config.IsNull() {
			updateMasks = append(updateMasks, "sync_databases")
		}
		if len(dataSourceList) > 0 {
			updateMasks = append(updateMasks, "data_sources")
		}

		if len(updateMasks) > 0 {
			if _, err := c.UpdateInstance(ctx, instance, updateMasks); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to update instance",
					Detail:   fmt.Sprintf("Update instance %s failed, error: %v", instanceName, err),
				})
				return diags
			}
		}
	} else {
		if _, err := c.CreateInstance(ctx, instanceID, instance); err != nil {
			return diag.FromErr(err)
		}
	}

	tflog.Debug(ctx, "[upsert instance] instance created, start to sync schema", map[string]interface{}{
		"instance": instanceName,
	})

	if err := c.SyncInstanceSchema(ctx, instanceName); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Instance schema sync failed",
			Detail:   fmt.Sprintf("Failed to sync schema for instance %s with error: %v. You can try to trigger the sync manually via Bytebase UI.", instanceName, err.Error()),
		})
	}
	d.SetId(instanceName)

	tflog.Debug(ctx, "[upsert instance] sync schema finished. now reading instance", map[string]interface{}{
		"instance": instanceName,
	})

	diag := resourceInstanceRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	tflog.Debug(ctx, "[upsert instance] upsert instance finished", map[string]interface{}{
		"instance": instanceName,
	})

	return diags
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	instanceName := d.Id()

	instance, err := c.GetInstance(ctx, instanceName)
	if err != nil {
		return diag.FromErr(err)
	}

	resp := setInstanceMessage(ctx, c, d, instance)
	tflog.Debug(ctx, "[read instance] read instance finished", map[string]interface{}{
		"instance": instance.Name,
	})
	return resp
}

func getSyncDatabases(d *schema.ResourceData) []string {
	rawSet, ok := d.Get("sync_databases").(*schema.Set)
	if !ok {
		return nil
	}
	dbList := []string{}
	for _, raw := range rawSet.List() {
		dbList = append(dbList, raw.(string))
	}
	return dbList
}

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("resource_id") {
		return diag.Errorf("cannot change the resource id")
	}
	if d.HasChange("engine") {
		return diag.Errorf("cannot change the engine in instance")
	}

	c := m.(api.Client)
	instanceName := d.Id()

	existedInstance, err := c.GetInstance(ctx, instanceName)
	if err != nil {
		return diag.Errorf("get instance %s failed with error: %v", instanceName, err)
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

	paths := []string{"data_sources"}
	if d.HasChange("title") {
		paths = append(paths, "title")
	}
	if d.HasChange("external_link") {
		paths = append(paths, "external_link")
	}
	if d.HasChange("environment") {
		paths = append(paths, "environment")
	}
	if d.HasChange("activation") {
		paths = append(paths, "activation")
	}
	if d.HasChange("data_sources") {
		paths = append(paths, "data_sources")
	}
	if d.HasChange("sync_interval") {
		paths = append(paths, "sync_interval")
	}
	if d.HasChange("sync_databases") {
		paths = append(paths, "sync_databases")
	}
	if d.HasChange("maximum_connections") {
		paths = append(paths, "maximum_connections")
	}

	if len(paths) > 0 {
		if _, err := c.UpdateInstance(ctx, &v1pb.Instance{
			Name:         instanceName,
			Title:        d.Get("title").(string),
			ExternalLink: d.Get("external_link").(string),
			Environment:  d.Get("environment").(string),
			Activation:   d.Get("activation").(bool),
			DataSources:  dataSourceList,
			State:        v1pb.State_ACTIVE,
			SyncInterval: &durationpb.Duration{
				Seconds: int64(d.Get("sync_interval").(int)),
			},
			MaximumConnections: int32(d.Get("maximum_connections").(int)),
			SyncDatabases:      getSyncDatabases(d),
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

func setInstanceMessage(
	ctx context.Context,
	client api.Client,
	d *schema.ResourceData,
	instance *v1pb.Instance,
) diag.Diagnostics {
	tflog.Debug(ctx, "[read instance] start reading instance", map[string]interface{}{
		"instance": instance.Name,
	})

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
	if err := d.Set("activation", instance.Activation); err != nil {
		return diag.Errorf("cannot set activation for instance: %s", err.Error())
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
	if v := instance.GetSyncInterval(); v != nil {
		if err := d.Set("sync_interval", v.GetSeconds()); err != nil {
			return diag.Errorf("cannot set sync_interval for instance: %s", err.Error())
		}
	}
	if err := d.Set("maximum_connections", instance.GetMaximumConnections()); err != nil {
		return diag.Errorf("cannot set maximum_connections for instance: %s", err.Error())
	}

	dataSources, err := flattenDataSourceList(d, instance.DataSources)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("data_sources", schema.NewSet(dataSourceHash, dataSources)); err != nil {
		return diag.Errorf("cannot set data_sources for instance: %s", err.Error())
	}

	tflog.Debug(ctx, "[read instance] start set instance databases", map[string]interface{}{
		"instance": instance.Name,
	})

	listAllDatabases := d.Get("list_all_databases").(bool)
	databases, err := client.ListDatabase(ctx, instance.Name, &api.DatabaseFilter{}, listAllDatabases)
	if err != nil {
		return diag.FromErr(err)
	}
	databaseList := flattenDatabaseList(databases)
	if err := d.Set("databases", databaseList); err != nil {
		return diag.Errorf("cannot set databases for instance: %s", err.Error())
	}

	if err := d.Set("sync_databases", instance.SyncDatabases); err != nil {
		return diag.Errorf("cannot set sync_databases for instance: %s", err.Error())
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
		raw["use_ssl"] = dataSource.UseSsl

		// These sensitive fields won't returned in the API. Propagate state value.
		if ds, ok := oldDataSourceMap[dataSource.Id]; ok {
			raw["password"] = ds.Password
			raw["ssl_ca"] = ds.SslCa
			raw["ssl_cert"] = ds.SslCert
			raw["ssl_key"] = ds.SslKey
		}

		if v := dataSource.ExternalSecret; v != nil {
			switch v.SecretType {
			case v1pb.DataSourceExternalSecret_GCP_SECRET_MANAGER:
				gcp := map[string]interface{}{
					"secret_name": v.SecretName,
				}
				raw["external_secret"] = []any{
					map[string]interface{}{
						"gcp_secret_manager": []any{gcp},
					},
				}
			case v1pb.DataSourceExternalSecret_AWS_SECRETS_MANAGER:
				aws := map[string]interface{}{
					"secret_name":       v.SecretName,
					"password_key_name": v.PasswordKeyName,
				}
				raw["external_secret"] = []any{
					map[string]interface{}{
						"aws_secrets_manager": []any{aws},
					},
				}
			case v1pb.DataSourceExternalSecret_VAULT_KV_V2:
				vault := map[string]interface{}{
					"url":               v.Url,
					"engine_name":       v.EngineName,
					"secret_name":       v.SecretName,
					"password_key_name": v.PasswordKeyName,
				}
				switch v.AuthType {
				case v1pb.DataSourceExternalSecret_TOKEN:
					if ds, ok := oldDataSourceMap[dataSource.Id]; ok && ds.GetExternalSecret() != nil {
						vault["token"] = ds.GetExternalSecret().GetToken()
					}
				case v1pb.DataSourceExternalSecret_VAULT_APP_ROLE:
					appRole := map[string]interface{}{
						"secret_type": v.GetAppRole().Type.String(),
					}
					if ds, ok := oldDataSourceMap[dataSource.Id]; ok && ds.GetExternalSecret() != nil {
						appRole["role_id"] = ds.GetExternalSecret().GetAppRole().GetRoleId()
						appRole["secret"] = ds.GetExternalSecret().GetAppRole().GetSecretId()
					}
					vault["app_role"] = []any{appRole}
				}
				raw["external_secret"] = []any{
					map[string]interface{}{
						"vault": []any{vault},
					},
				}
			}
		}
		res = append(res, raw)
	}
	return res, nil
}

func dataSourceHash(rawDataSource interface{}) int {
	var buf bytes.Buffer
	raw := rawDataSource.(map[string]interface{})

	if v, ok := raw["id"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := raw["username"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := raw["password"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := raw["host"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := raw["port"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := raw["database"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}

	// Include use_ssl in hash to detect SSL enablement changes
	if v, ok := raw["use_ssl"].(bool); ok {
		_, _ = buf.WriteString(fmt.Sprintf("ssl_%v-", v))
	}
	// Include whether SSL certificates are present (not the values themselves)
	if v, ok := raw["ssl_ca"].(string); ok && v != "" {
		_, _ = buf.WriteString(fmt.Sprintf("ca_present_%s-", v))
	}
	if v, ok := raw["ssl_cert"].(string); ok && v != "" {
		_, _ = buf.WriteString(fmt.Sprintf("cert_present_%s-", v))
	}
	if v, ok := raw["ssl_key"].(string); ok && v != "" {
		_, _ = buf.WriteString(fmt.Sprintf("key_present_%s-", v))
	}

	return internal.ToHashcodeInt(buf.String())
}

func convertDataSourceCreateList(d *schema.ResourceData, validate bool) ([]*v1pb.DataSource, error) {
	var dataSourceList []*v1pb.DataSource
	dataSourceSet, ok := d.Get("data_sources").(*schema.Set)
	if !ok {
		return dataSourceList, nil
	}

	dataSourceTypeMap := map[v1pb.DataSourceType]bool{}
	for _, raw := range dataSourceSet.List() {
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
		if v, ok := obj["external_secret"].([]interface{}); ok && len(v) == 1 {
			externalSecret := &v1pb.DataSourceExternalSecret{}
			rawExternalSecret := v[0].(map[string]interface{})
			if v, ok := rawExternalSecret["vault"].([]interface{}); ok && len(v) == 1 {
				rawVault := v[0].(map[string]interface{})
				externalSecret.SecretType = v1pb.DataSourceExternalSecret_VAULT_KV_V2
				externalSecret.Url = rawVault["url"].(string)
				externalSecret.EngineName = rawVault["engine_name"].(string)
				externalSecret.SecretName = rawVault["secret_name"].(string)
				externalSecret.PasswordKeyName = rawVault["password_key_name"].(string)

				if token, ok := rawVault["token"].(string); ok && token != "" {
					externalSecret.AuthType = v1pb.DataSourceExternalSecret_TOKEN
					externalSecret.AuthOption = &v1pb.DataSourceExternalSecret_Token{
						Token: token,
					}
				} else if v, ok := rawVault["app_role"].([]interface{}); ok && len(v) == 1 {
					rawAppRole := v[0].(map[string]interface{})
					externalSecret.AuthType = v1pb.DataSourceExternalSecret_VAULT_APP_ROLE
					externalSecret.AuthOption = &v1pb.DataSourceExternalSecret_AppRole{
						AppRole: &v1pb.DataSourceExternalSecret_AppRoleAuthOption{
							RoleId:   rawAppRole["role_id"].(string),
							SecretId: rawAppRole["secret"].(string),
							Type:     v1pb.DataSourceExternalSecret_AppRoleAuthOption_SecretType(v1pb.DataSourceExternalSecret_AppRoleAuthOption_SecretType_value[rawAppRole["secret_type"].(string)]),
						},
					}
				} else {
					return nil, errors.Errorf("require token or app_role for Vault")
				}
			} else if v, ok := rawExternalSecret["aws_secrets_manager"].([]interface{}); ok && len(v) == 1 {
				rawAWS := v[0].(map[string]interface{})
				externalSecret.SecretType = v1pb.DataSourceExternalSecret_AWS_SECRETS_MANAGER
				externalSecret.SecretName = rawAWS["secret_name"].(string)
				externalSecret.PasswordKeyName = rawAWS["password_key_name"].(string)
			} else if v, ok := rawExternalSecret["gcp_secret_manager"].([]interface{}); ok && len(v) == 1 {
				rawGCP := v[0].(map[string]interface{})
				externalSecret.SecretType = v1pb.DataSourceExternalSecret_GCP_SECRET_MANAGER
				externalSecret.SecretName = rawGCP["secret_name"].(string)
			} else {
				return nil, errors.Errorf("must set one of vault, aws_secrets_manager or gcp_secret_manager")
			}
			dataSource.ExternalSecret = externalSecret
		}
		if dataSource.Password != "" && dataSource.ExternalSecret != nil {
			return nil, errors.Errorf("cannot set both password and external_secret")
		}

		if v, ok := obj["use_ssl"].(bool); ok {
			dataSource.UseSsl = v
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

	return dataSourceList, nil
}
