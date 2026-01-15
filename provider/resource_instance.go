package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/durationpb"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func resourceInstance() *schema.Resource {
	return &schema.Resource{
		Description:          "The instance resource.",
		CreateWithoutTimeout: resourceInstanceCreate,
		ReadWithoutTimeout:   resourceInstanceRead,
		UpdateContext:        resourceInstanceUpdate,
		DeleteContext:        resourceInstanceDelete,
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
				Optional: true,
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
				Description:  "The instance engine. Supported engines: MYSQL, POSTGRES, TIDB, SNOWFLAKE, CLICKHOUSE, MONGODB, SQLITE, REDIS, ORACLE, SPANNER, MSSQL, REDSHIFT, MARIADB, OCEANBASE, STARROCKS, DORIS, HIVE, ELASTICSEARCH, BIGQUERY, DYNAMODB, DATABRICKS, COCKROACHDB, COSMOSDB, TRINO, CASSANDRA.",
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
				Description: "How often the instance is synced in seconds. Default 0, means never sync. Require instance license to enable this feature.",
			},
			"sync_databases": getSyncDatabasesSchema(false),
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
							Description: "The data source type. Should be ADMIN or READ_ONLY. The READ_ONLY data source requires the instance license.",
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
							MinItems:    0,
							Description: "The external secret to get the database password. Only available when authentication_type is PASSWORD. Requires instance license. Learn more: https://www.bytebase.com/docs/get-started/instance/#use-external-secret-manager",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"vault": {
										Type:        schema.TypeList,
										Optional:    true,
										MaxItems:    1,
										MinItems:    0,
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
													MinItems:    0,
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
										MinItems:    0,
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
										MinItems:    0,
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
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Host or socket for your instance, or the account name if the instance type is Snowflake. Not required for some engines like DYNAMODB.",
						},
						"port": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The port for your instance. Not required for some engines like SPANNER, BIGQUERY.",
						},
						"database": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The database for the instance, you can set this if the engine type is POSTGRES.",
						},
						// SSL/Security
						"verify_tls_certificate": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Enable TLS certificate verification for SSL connections.",
						},
						// MongoDB-specific (only available for MONGODB engine)
						"srv": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Use DNS SRV record for MongoDB connection. Only available for MONGODB engine.",
						},
						"authentication_database": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The database to authenticate against for MongoDB. Only available for MONGODB engine.",
						},
						"replica_set": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The replica set name for MongoDB. Only available for MONGODB engine.",
						},
						"direct_connection": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Use direct connection to MongoDB node. Only available for MONGODB engine.",
						},
						"additional_addresses": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Additional addresses for MongoDB replica set. Only available for MONGODB engine.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"host": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The hostname of the additional address.",
									},
									"port": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The port of the additional address.",
									},
								},
							},
						},
						// Oracle-specific (only available for ORACLE engine)
						"sid": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Oracle System Identifier (SID). Only available for ORACLE engine.",
						},
						"service_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Oracle service name. Only available for ORACLE engine.",
						},
						// SSH Tunneling (requires PASSWORD authentication_type and specific engines)
						"ssh_host": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "SSH tunnel server hostname. Only available for MYSQL, TIDB, MARIADB, OCEANBASE, POSTGRES, REDIS with PASSWORD authentication.",
						},
						"ssh_port": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "SSH tunnel server port. Only available for MYSQL, TIDB, MARIADB, OCEANBASE, POSTGRES, REDIS with PASSWORD authentication.",
						},
						"ssh_user": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "SSH tunnel username. Only available for MYSQL, TIDB, MARIADB, OCEANBASE, POSTGRES, REDIS with PASSWORD authentication.",
						},
						"ssh_password": {
							Type:             schema.TypeString,
							Optional:         true,
							Sensitive:        true,
							Computed:         true,
							DiffSuppressFunc: suppressSensitiveFieldDiff,
							Description:      "SSH tunnel password. Only available for MYSQL, TIDB, MARIADB, OCEANBASE, POSTGRES, REDIS with PASSWORD authentication.",
						},
						"ssh_private_key": {
							Type:             schema.TypeString,
							Optional:         true,
							Sensitive:        true,
							Computed:         true,
							DiffSuppressFunc: suppressSensitiveFieldDiff,
							Description:      "SSH tunnel private key. Only available for MYSQL, TIDB, MARIADB, OCEANBASE, POSTGRES, REDIS with PASSWORD authentication.",
						},
						// Authentication
						"authentication_type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								v1pb.DataSource_PASSWORD.String(),
								v1pb.DataSource_GOOGLE_CLOUD_SQL_IAM.String(),
								v1pb.DataSource_AWS_RDS_IAM.String(),
								v1pb.DataSource_AZURE_IAM.String(),
							}, false),
							Description: "Authentication type. Supported values depend on engine: " +
								"COSMOSDB only supports AZURE_IAM; " +
								"MSSQL supports PASSWORD, AZURE_IAM; " +
								"ELASTICSEARCH supports PASSWORD, AWS_RDS_IAM; " +
								"SPANNER, BIGQUERY only support GOOGLE_CLOUD_SQL_IAM; " +
								"Most other engines support PASSWORD, GOOGLE_CLOUD_SQL_IAM, AWS_RDS_IAM. " +
								"Default is PASSWORD.",
						},
						"authentication_private_key": {
							Type:             schema.TypeString,
							Optional:         true,
							Sensitive:        true,
							Computed:         true,
							DiffSuppressFunc: suppressSensitiveFieldDiff,
							Description:      "PKCS#8 private key for authentication.",
						},
						"authentication_private_key_passphrase": {
							Type:             schema.TypeString,
							Optional:         true,
							Sensitive:        true,
							Computed:         true,
							DiffSuppressFunc: suppressSensitiveFieldDiff,
							Description:      "Passphrase for encrypted private key.",
						},
						// Redis-specific (only available for REDIS engine)
						"redis_type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								v1pb.DataSource_STANDALONE.String(),
								v1pb.DataSource_SENTINEL.String(),
								v1pb.DataSource_CLUSTER.String(),
							}, false),
							Description: "Redis deployment type: STANDALONE, SENTINEL, CLUSTER. Only available for REDIS engine.",
						},
						"master_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Redis Sentinel master name. Only available for REDIS engine.",
						},
						"master_username": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Redis Sentinel master username. Only available for REDIS engine.",
						},
						"master_password": {
							Type:             schema.TypeString,
							Optional:         true,
							Sensitive:        true,
							Computed:         true,
							DiffSuppressFunc: suppressSensitiveFieldDiff,
							Description:      "Redis Sentinel master password. Only available for REDIS engine.",
						},
						// Cloud-specific
						"region": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "AWS region (e.g., us-east-1). Only available when authentication_type is AWS_RDS_IAM.",
						},
						"warehouse_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Databricks warehouse ID. Only available for DATABRICKS engine.",
						},
						"cluster": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "CockroachDB cluster name. Only available for COCKROACHDB engine.",
						},
						// IAM Credentials (each only valid for its respective authentication_type)
						"azure_credential": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Azure IAM credential. Only valid when authentication_type is AZURE_IAM.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"tenant_id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Azure tenant ID.",
									},
									"client_id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Azure client ID.",
									},
									"client_secret": {
										Type:        schema.TypeString,
										Required:    true,
										Sensitive:   true,
										Description: "Azure client secret.",
									},
								},
							},
						},
						"aws_credential": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "AWS IAM credential. Only valid when authentication_type is AWS_RDS_IAM.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"access_key_id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "AWS access key ID.",
									},
									"secret_access_key": {
										Type:        schema.TypeString,
										Required:    true,
										Sensitive:   true,
										Description: "AWS secret access key.",
									},
									"session_token": {
										Type:        schema.TypeString,
										Optional:    true,
										Sensitive:   true,
										Description: "AWS session token.",
									},
									"role_arn": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "ARN of IAM role to assume for cross-account access.",
									},
									"external_id": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "External ID for additional security when assuming role.",
									},
								},
							},
						},
						"gcp_credential": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "GCP IAM credential (service account JSON). Only valid when authentication_type is GOOGLE_CLOUD_SQL_IAM.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"content": {
										Type:        schema.TypeString,
										Required:    true,
										Sensitive:   true,
										Description: "GCP service account JSON content.",
									},
								},
							},
						},
						// SASL Config (only available for HIVE engine)
						"sasl_config": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "SASL authentication configuration. Only available for HIVE engine.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"kerberos": {
										Type:        schema.TypeList,
										Optional:    true,
										MaxItems:    1,
										Description: "Kerberos configuration.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"primary": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The primary component of the Kerberos principal.",
												},
												"instance": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "The instance component of the Kerberos principal.",
												},
												"realm": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The Kerberos realm.",
												},
												"keytab": {
													Type:        schema.TypeString,
													Required:    true,
													Sensitive:   true,
													Description: "The keytab file contents for authentication (base64 encoded).",
												},
												"kdc_host": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The hostname of the Key Distribution Center (KDC).",
												},
												"kdc_port": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "The port of the Key Distribution Center (KDC).",
												},
												"kdc_transport_protocol": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "The transport protocol for KDC communication (tcp or udp).",
												},
											},
										},
									},
								},
							},
						},
						// Extra connection parameters
						"extra_connection_parameters": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "Extra connection parameters as key-value pairs. Only available for MYSQL, MARIADB, OCEANBASE, POSTGRES, ORACLE, MSSQL, MONGODB.",
							Elem:        &schema.Schema{Type: schema.TypeString},
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
		Activation:    d.Get("activation").(bool),
		State:         v1pb.State_ACTIVE,
		Engine:        v1pb.Engine(v1pb.Engine_value[d.Get("engine").(string)]),
		SyncDatabases:      getSyncDatabases(d),
	}
	environment := d.Get("environment").(string)
	if environment != "" {
		instance.Environment = &environment
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
		// Check if the resource was deleted outside of Terraform
		if internal.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Resource %s not found, removing from state", instanceName))
			// Remove from state to trigger recreation on next apply
			d.SetId("")
			return nil
		}
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
	environment := d.Get("environment").(string)

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

	if len(paths) > 0 {
		if _, err := c.UpdateInstance(ctx, &v1pb.Instance{
			Name:         instanceName,
			Title:        d.Get("title").(string),
			ExternalLink: d.Get("external_link").(string),
			Environment:  &environment,
			Activation:   d.Get("activation").(bool),
			DataSources:  dataSourceList,
			State:        v1pb.State_ACTIVE,
			SyncInterval: &durationpb.Duration{
				Seconds: int64(d.Get("sync_interval").(int)),
			},
			SyncDatabases: getSyncDatabases(d),
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
	if v := instance.Environment; v != nil {
		if err := d.Set("environment", *v); err != nil {
			return diag.Errorf("cannot set environment for instance: %s", err.Error())
		}
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

		// SSL/Security
		raw["verify_tls_certificate"] = dataSource.VerifyTlsCertificate

		// MongoDB
		raw["srv"] = dataSource.Srv
		raw["authentication_database"] = dataSource.AuthenticationDatabase
		raw["replica_set"] = dataSource.ReplicaSet
		raw["direct_connection"] = dataSource.DirectConnection

		// Additional addresses for MongoDB
		if len(dataSource.AdditionalAddresses) > 0 {
			addresses := make([]interface{}, 0, len(dataSource.AdditionalAddresses))
			for _, addr := range dataSource.AdditionalAddresses {
				addresses = append(addresses, map[string]interface{}{
					"host": addr.Host,
					"port": addr.Port,
				})
			}
			raw["additional_addresses"] = addresses
		}

		// Oracle
		raw["sid"] = dataSource.Sid
		raw["service_name"] = dataSource.ServiceName

		// SSH Tunneling - non-sensitive fields
		raw["ssh_host"] = dataSource.SshHost
		raw["ssh_port"] = dataSource.SshPort
		raw["ssh_user"] = dataSource.SshUser

		// Authentication - non-sensitive fields
		if dataSource.AuthenticationType != v1pb.DataSource_AUTHENTICATION_UNSPECIFIED {
			raw["authentication_type"] = dataSource.AuthenticationType.String()
		}

		// Redis - non-sensitive fields
		if dataSource.RedisType != v1pb.DataSource_REDIS_TYPE_UNSPECIFIED {
			raw["redis_type"] = dataSource.RedisType.String()
		}
		raw["master_name"] = dataSource.MasterName
		raw["master_username"] = dataSource.MasterUsername

		// Cloud-specific
		raw["region"] = dataSource.Region
		raw["warehouse_id"] = dataSource.WarehouseId
		raw["cluster"] = dataSource.Cluster

		// Extra connection parameters
		if len(dataSource.ExtraConnectionParameters) > 0 {
			raw["extra_connection_parameters"] = dataSource.ExtraConnectionParameters
		}

		// These sensitive fields won't returned in the API. Propagate state value.
		if ds, ok := oldDataSourceMap[dataSource.Id]; ok {
			raw["ssl_ca"] = ds.SslCa
			raw["ssl_cert"] = ds.SslCert
			raw["ssl_key"] = ds.SslKey
			// SSH sensitive fields
			raw["ssh_password"] = ds.SshPassword
			raw["ssh_private_key"] = ds.SshPrivateKey
			// Authentication sensitive fields
			raw["authentication_private_key"] = ds.AuthenticationPrivateKey
			raw["authentication_private_key_passphrase"] = ds.AuthenticationPrivateKeyPassphrase
			// Redis sensitive fields
			raw["master_password"] = ds.MasterPassword

			// Propagate password or IAM credentials based on authentication_type
			switch dataSource.AuthenticationType {
			case v1pb.DataSource_PASSWORD, v1pb.DataSource_AUTHENTICATION_UNSPECIFIED:
				raw["password"] = ds.Password
			case v1pb.DataSource_AZURE_IAM:
				if ds.GetAzureCredential() != nil {
					raw["azure_credential"] = []any{
						map[string]interface{}{
							"tenant_id":     ds.GetAzureCredential().GetTenantId(),
							"client_id":     ds.GetAzureCredential().GetClientId(),
							"client_secret": ds.GetAzureCredential().GetClientSecret(),
						},
					}
				}
			case v1pb.DataSource_AWS_RDS_IAM:
				if ds.GetAwsCredential() != nil {
					awsCred := map[string]interface{}{
						"access_key_id":     ds.GetAwsCredential().GetAccessKeyId(),
						"secret_access_key": ds.GetAwsCredential().GetSecretAccessKey(),
					}
					if ds.GetAwsCredential().GetSessionToken() != "" {
						awsCred["session_token"] = ds.GetAwsCredential().GetSessionToken()
					}
					if ds.GetAwsCredential().GetRoleArn() != "" {
						awsCred["role_arn"] = ds.GetAwsCredential().GetRoleArn()
					}
					if ds.GetAwsCredential().GetExternalId() != "" {
						awsCred["external_id"] = ds.GetAwsCredential().GetExternalId()
					}
					raw["aws_credential"] = []any{awsCred}
				}
			case v1pb.DataSource_GOOGLE_CLOUD_SQL_IAM:
				if ds.GetGcpCredential() != nil {
					raw["gcp_credential"] = []any{
						map[string]interface{}{
							"content": ds.GetGcpCredential().GetContent(),
						},
					}
				}
			}

			// SASL config - propagate from old state
			if ds.GetSaslConfig() != nil && ds.GetSaslConfig().GetKrbConfig() != nil {
				krbConfig := ds.GetSaslConfig().GetKrbConfig()
				raw["sasl_config"] = []any{
					map[string]interface{}{
						"kerberos": []any{
							map[string]interface{}{
								"primary":                krbConfig.GetPrimary(),
								"instance":               krbConfig.GetInstance(),
								"realm":                  krbConfig.GetRealm(),
								"keytab":                 string(krbConfig.GetKeytab()),
								"kdc_host":               krbConfig.GetKdcHost(),
								"kdc_port":               krbConfig.GetKdcPort(),
								"kdc_transport_protocol": krbConfig.GetKdcTransportProtocol(),
							},
						},
					},
				}
			}
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
	ds, err := convertToV1DataSource(rawDataSource)
	if err != nil {
		return 0
	}
	return internal.ToHash(ds)
}

func convertToV1DataSource(raw interface{}) (*v1pb.DataSource, error) {
	obj := raw.(map[string]interface{})
	dataSource := &v1pb.DataSource{
		Id:   obj["id"].(string),
		Type: v1pb.DataSourceType(v1pb.DataSourceType_value[obj["type"].(string)]),
	}

	if v, ok := obj["username"].(string); ok {
		dataSource.Username = v
	}

	// Determine authentication type early for validation
	authType := v1pb.DataSource_PASSWORD // default to PASSWORD
	if v, ok := obj["authentication_type"].(string); ok && v != "" {
		authType = v1pb.DataSource_AuthenticationType(v1pb.DataSource_AuthenticationType_value[v])
	}
	dataSource.AuthenticationType = authType

	if v, ok := obj["password"].(string); ok && v != "" {
		dataSource.Password = v
	}

	// External secret requires PASSWORD authentication
	hasExternalSecret := false
	if v, ok := obj["external_secret"].([]interface{}); ok && len(v) == 1 {
		hasExternalSecret = true
	}
	if hasExternalSecret && authType != v1pb.DataSource_PASSWORD && authType != v1pb.DataSource_AUTHENTICATION_UNSPECIFIED {
		return nil, errors.Errorf("external_secret is only available when authentication_type is PASSWORD")
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
	// SSL/Security
	if v, ok := obj["verify_tls_certificate"].(bool); ok {
		dataSource.VerifyTlsCertificate = v
	}
	// MongoDB
	if v, ok := obj["srv"].(bool); ok {
		dataSource.Srv = v
	}
	if v, ok := obj["authentication_database"].(string); ok {
		dataSource.AuthenticationDatabase = v
	}
	if v, ok := obj["replica_set"].(string); ok {
		dataSource.ReplicaSet = v
	}
	if v, ok := obj["direct_connection"].(bool); ok {
		dataSource.DirectConnection = v
	}
	if v, ok := obj["additional_addresses"].([]interface{}); ok && len(v) > 0 {
		for _, addr := range v {
			addrMap := addr.(map[string]interface{})
			dataSource.AdditionalAddresses = append(dataSource.AdditionalAddresses, &v1pb.DataSource_Address{
				Host: addrMap["host"].(string),
				Port: addrMap["port"].(string),
			})
		}
	}
	// Oracle
	if v, ok := obj["sid"].(string); ok {
		dataSource.Sid = v
	}
	if v, ok := obj["service_name"].(string); ok {
		dataSource.ServiceName = v
	}
	// SSH Tunneling
	sshHost, _ := obj["ssh_host"].(string)
	sshPort, _ := obj["ssh_port"].(string)
	sshUser, _ := obj["ssh_user"].(string)
	sshPassword, _ := obj["ssh_password"].(string)
	sshPrivateKey, _ := obj["ssh_private_key"].(string)
	hasSSHConfig := sshHost != "" || sshPort != "" || sshUser != "" || sshPassword != "" || sshPrivateKey != ""

	dataSource.SshHost = sshHost
	dataSource.SshPort = sshPort
	dataSource.SshUser = sshUser
	if sshPassword != "" {
		dataSource.SshPassword = sshPassword
	}
	if sshPrivateKey != "" {
		dataSource.SshPrivateKey = sshPrivateKey
	}
	// SSH tunneling requires PASSWORD authentication
	if hasSSHConfig && authType != v1pb.DataSource_PASSWORD && authType != v1pb.DataSource_AUTHENTICATION_UNSPECIFIED {
		return nil, errors.Errorf("SSH tunneling is only available when authentication_type is PASSWORD")
	}

	// Check which credentials are provided
	hasPassword := dataSource.Password != ""
	hasAzureCredential := false
	hasAwsCredential := false
	hasGcpCredential := false
	if v, ok := obj["azure_credential"].([]interface{}); ok && len(v) == 1 {
		hasAzureCredential = true
	}
	if v, ok := obj["aws_credential"].([]interface{}); ok && len(v) == 1 {
		hasAwsCredential = true
	}
	if v, ok := obj["gcp_credential"].([]interface{}); ok && len(v) == 1 {
		hasGcpCredential = true
	}

	// Validate credentials based on authentication type
	// Rules:
	// - PASSWORD: password allowed, IAM credentials NOT allowed
	// - AZURE_IAM: password NOT allowed, azure_credential optional, other IAM credentials NOT allowed
	// - AWS_RDS_IAM: password NOT allowed, aws_credential optional, other IAM credentials NOT allowed
	// - GOOGLE_CLOUD_SQL_IAM: password NOT allowed, gcp_credential optional, other IAM credentials NOT allowed
	switch authType {
	case v1pb.DataSource_PASSWORD, v1pb.DataSource_AUTHENTICATION_UNSPECIFIED:
		if hasAzureCredential || hasAwsCredential || hasGcpCredential {
			return nil, errors.Errorf("cannot set azure_credential, aws_credential, or gcp_credential when authentication_type is PASSWORD")
		}
		// Password is handled above, nothing more to do
	case v1pb.DataSource_AZURE_IAM:
		if hasPassword {
			return nil, errors.Errorf("cannot set password when authentication_type is AZURE_IAM")
		}
		if hasAwsCredential || hasGcpCredential {
			return nil, errors.Errorf("cannot set aws_credential or gcp_credential when authentication_type is AZURE_IAM")
		}
		// azure_credential is optional for AZURE_IAM
		if hasAzureCredential {
			azureMap := obj["azure_credential"].([]interface{})[0].(map[string]interface{})
			dataSource.IamExtension = &v1pb.DataSource_AzureCredential_{
				AzureCredential: &v1pb.DataSource_AzureCredential{
					TenantId:     azureMap["tenant_id"].(string),
					ClientId:     azureMap["client_id"].(string),
					ClientSecret: azureMap["client_secret"].(string),
				},
			}
		}
	case v1pb.DataSource_AWS_RDS_IAM:
		if hasPassword {
			return nil, errors.Errorf("cannot set password when authentication_type is AWS_RDS_IAM")
		}
		if hasAzureCredential || hasGcpCredential {
			return nil, errors.Errorf("cannot set azure_credential or gcp_credential when authentication_type is AWS_RDS_IAM")
		}
		// aws_credential is optional for AWS_RDS_IAM
		if hasAwsCredential {
			awsMap := obj["aws_credential"].([]interface{})[0].(map[string]interface{})
			awsCred := &v1pb.DataSource_AWSCredential{
				AccessKeyId:     awsMap["access_key_id"].(string),
				SecretAccessKey: awsMap["secret_access_key"].(string),
			}
			if sessionToken, ok := awsMap["session_token"].(string); ok && sessionToken != "" {
				awsCred.SessionToken = sessionToken
			}
			if roleArn, ok := awsMap["role_arn"].(string); ok && roleArn != "" {
				awsCred.RoleArn = roleArn
			}
			if externalID, ok := awsMap["external_id"].(string); ok && externalID != "" {
				awsCred.ExternalId = externalID
			}
			dataSource.IamExtension = &v1pb.DataSource_AwsCredential{
				AwsCredential: awsCred,
			}
		}
	case v1pb.DataSource_GOOGLE_CLOUD_SQL_IAM:
		if hasPassword {
			return nil, errors.Errorf("cannot set password when authentication_type is GOOGLE_CLOUD_SQL_IAM")
		}
		if hasAzureCredential || hasAwsCredential {
			return nil, errors.Errorf("cannot set azure_credential or aws_credential when authentication_type is GOOGLE_CLOUD_SQL_IAM")
		}
		// gcp_credential is optional for GOOGLE_CLOUD_SQL_IAM
		if hasGcpCredential {
			gcpMap := obj["gcp_credential"].([]interface{})[0].(map[string]interface{})
			dataSource.IamExtension = &v1pb.DataSource_GcpCredential{
				GcpCredential: &v1pb.DataSource_GCPCredential{
					Content: gcpMap["content"].(string),
				},
			}
		}
	}

	if v, ok := obj["authentication_private_key"].(string); ok && v != "" {
		dataSource.AuthenticationPrivateKey = v
	}
	if v, ok := obj["authentication_private_key_passphrase"].(string); ok && v != "" {
		dataSource.AuthenticationPrivateKeyPassphrase = v
	}
	// Redis
	if v, ok := obj["redis_type"].(string); ok && v != "" {
		dataSource.RedisType = v1pb.DataSource_RedisType(v1pb.DataSource_RedisType_value[v])
	}
	if v, ok := obj["master_name"].(string); ok {
		dataSource.MasterName = v
	}
	if v, ok := obj["master_username"].(string); ok {
		dataSource.MasterUsername = v
	}
	if v, ok := obj["master_password"].(string); ok && v != "" {
		dataSource.MasterPassword = v
	}
	// Cloud-specific
	if v, ok := obj["region"].(string); ok {
		dataSource.Region = v
	}
	if v, ok := obj["warehouse_id"].(string); ok {
		dataSource.WarehouseId = v
	}
	if v, ok := obj["cluster"].(string); ok {
		dataSource.Cluster = v
	}
	// SASL Config
	if v, ok := obj["sasl_config"].([]interface{}); ok && len(v) == 1 {
		saslMap := v[0].(map[string]interface{})
		if kerberos, ok := saslMap["kerberos"].([]interface{}); ok && len(kerberos) == 1 {
			kerberosMap := kerberos[0].(map[string]interface{})
			krbConfig := &v1pb.KerberosConfig{
				Primary: kerberosMap["primary"].(string),
				Realm:   kerberosMap["realm"].(string),
				KdcHost: kerberosMap["kdc_host"].(string),
			}
			if instance, ok := kerberosMap["instance"].(string); ok {
				krbConfig.Instance = instance
			}
			if keytab, ok := kerberosMap["keytab"].(string); ok {
				krbConfig.Keytab = []byte(keytab)
			}
			if kdcPort, ok := kerberosMap["kdc_port"].(string); ok {
				krbConfig.KdcPort = kdcPort
			}
			if kdcTransportProtocol, ok := kerberosMap["kdc_transport_protocol"].(string); ok {
				krbConfig.KdcTransportProtocol = kdcTransportProtocol
			}
			dataSource.SaslConfig = &v1pb.SASLConfig{
				Mechanism: &v1pb.SASLConfig_KrbConfig{
					KrbConfig: krbConfig,
				},
			}
		}
	}
	// Extra connection parameters
	if v, ok := obj["extra_connection_parameters"].(map[string]interface{}); ok && len(v) > 0 {
		dataSource.ExtraConnectionParameters = make(map[string]string)
		for key, value := range v {
			dataSource.ExtraConnectionParameters[key] = value.(string)
		}
	}
	return dataSource, nil
}

func convertDataSourceCreateList(d *schema.ResourceData, validate bool) ([]*v1pb.DataSource, error) {
	var dataSourceList []*v1pb.DataSource
	dataSourceSet, ok := d.Get("data_sources").(*schema.Set)
	if !ok {
		return dataSourceList, nil
	}

	// Get engine type for validation
	engineStr := d.Get("engine").(string)
	engine := v1pb.Engine(v1pb.Engine_value[engineStr])

	dataSourceTypeMap := map[v1pb.DataSourceType]bool{}
	for _, raw := range dataSourceSet.List() {
		dataSource, err := convertToV1DataSource(raw)
		if err != nil {
			return nil, err
		}

		// Validate authentication_type is supported for the engine and engine-specific fields
		if validate {
			if err := validateAuthenticationTypeForEngine(engine, dataSource.AuthenticationType); err != nil {
				return nil, err
			}
			if err := validateDataSourceFieldsForEngine(engine, dataSource); err != nil {
				return nil, err
			}
		}

		if dataSourceTypeMap[dataSource.Type] && dataSource.Type == v1pb.DataSourceType_ADMIN {
			return nil, errors.Errorf("duplicate data source type ADMIN")
		}
		dataSourceTypeMap[dataSource.Type] = true
		dataSourceList = append(dataSourceList, dataSource)
	}

	if !dataSourceTypeMap[v1pb.DataSourceType_ADMIN] && validate {
		return nil, errors.Errorf("data source \"%v\" is required", v1pb.DataSourceType_ADMIN.String())
	}

	return dataSourceList, nil
}

// validateAuthenticationTypeForEngine validates that the authentication type is supported for the given engine
// Engine-specific rules:
// - COSMOSDB: only AZURE_IAM
// - MSSQL: PASSWORD, AZURE_IAM
// - ELASTICSEARCH: PASSWORD, AWS_RDS_IAM
// - SPANNER, BIGQUERY: only GOOGLE_CLOUD_SQL_IAM
// - Others: PASSWORD, GOOGLE_CLOUD_SQL_IAM, AWS_RDS_IAM.
func validateAuthenticationTypeForEngine(engine v1pb.Engine, authType v1pb.DataSource_AuthenticationType) error {
	// Treat AUTHENTICATION_UNSPECIFIED as PASSWORD
	if authType == v1pb.DataSource_AUTHENTICATION_UNSPECIFIED {
		authType = v1pb.DataSource_PASSWORD
	}

	switch engine {
	case v1pb.Engine_COSMOSDB:
		if authType != v1pb.DataSource_AZURE_IAM {
			return errors.Errorf("COSMOSDB only supports AZURE_IAM authentication")
		}
	case v1pb.Engine_MSSQL:
		if authType != v1pb.DataSource_PASSWORD && authType != v1pb.DataSource_AZURE_IAM {
			return errors.Errorf("MSSQL only supports PASSWORD or AZURE_IAM authentication")
		}
	case v1pb.Engine_ELASTICSEARCH:
		if authType != v1pb.DataSource_PASSWORD && authType != v1pb.DataSource_AWS_RDS_IAM {
			return errors.Errorf("ELASTICSEARCH only supports PASSWORD or AWS_RDS_IAM authentication")
		}
	case v1pb.Engine_SPANNER, v1pb.Engine_BIGQUERY:
		if authType != v1pb.DataSource_GOOGLE_CLOUD_SQL_IAM {
			return errors.Errorf("%s only supports GOOGLE_CLOUD_SQL_IAM authentication", engine.String())
		}
	default:
		// Most engines support PASSWORD, GOOGLE_CLOUD_SQL_IAM, AWS_RDS_IAM
		if authType != v1pb.DataSource_PASSWORD &&
			authType != v1pb.DataSource_GOOGLE_CLOUD_SQL_IAM &&
			authType != v1pb.DataSource_AWS_RDS_IAM {
			return errors.Errorf("authentication_type %s is not supported for engine %s", authType.String(), engine.String())
		}
	}
	return nil
}

// validateDataSourceFieldsForEngine validates that engine-specific fields are only set for appropriate engines.
func validateDataSourceFieldsForEngine(engine v1pb.Engine, ds *v1pb.DataSource) error {
	// MongoDB-specific fields
	if ds.Srv || ds.AuthenticationDatabase != "" || ds.ReplicaSet != "" || ds.DirectConnection || len(ds.AdditionalAddresses) > 0 {
		if engine != v1pb.Engine_MONGODB {
			return errors.Errorf("srv, authentication_database, replica_set, direct_connection, and additional_addresses are only available for MONGODB")
		}
	}

	// Oracle-specific fields
	if ds.Sid != "" || ds.ServiceName != "" {
		if engine != v1pb.Engine_ORACLE {
			return errors.Errorf("sid and service_name are only available for ORACLE")
		}
	}

	// Redis-specific fields
	if ds.RedisType != v1pb.DataSource_REDIS_TYPE_UNSPECIFIED || ds.MasterName != "" || ds.MasterUsername != "" || ds.MasterPassword != "" {
		if engine != v1pb.Engine_REDIS {
			return errors.Errorf("redis_type, master_name, master_username, and master_password are only available for REDIS")
		}
	}

	// Databricks-specific fields
	if ds.WarehouseId != "" {
		if engine != v1pb.Engine_DATABRICKS {
			return errors.Errorf("warehouse_id is only available for DATABRICKS")
		}
	}

	// CockroachDB-specific fields
	if ds.Cluster != "" {
		if engine != v1pb.Engine_COCKROACHDB {
			return errors.Errorf("cluster is only available for COCKROACHDB")
		}
	}

	// SASL/Kerberos is only for HIVE
	if ds.SaslConfig != nil {
		if engine != v1pb.Engine_HIVE {
			return errors.Errorf("sasl_config is only available for HIVE")
		}
	}

	// Region is only for AWS_RDS_IAM authentication
	if ds.Region != "" {
		if ds.AuthenticationType != v1pb.DataSource_AWS_RDS_IAM {
			return errors.Errorf("region is only available when authentication_type is AWS_RDS_IAM")
		}
	}

	// Determine effective authentication type (UNSPECIFIED treated as PASSWORD)
	effectiveAuthType := ds.AuthenticationType
	if effectiveAuthType == v1pb.DataSource_AUTHENTICATION_UNSPECIFIED {
		effectiveAuthType = v1pb.DataSource_PASSWORD
	}

	// External secret is only for PASSWORD authentication
	if ds.ExternalSecret != nil {
		if effectiveAuthType != v1pb.DataSource_PASSWORD {
			return errors.Errorf("external_secret is only available when authentication_type is PASSWORD")
		}
	}

	// SSH tunnel is only for PASSWORD authentication and specific engines
	if ds.SshHost != "" || ds.SshPort != "" || ds.SshUser != "" || ds.SshPassword != "" || ds.SshPrivateKey != "" {
		if effectiveAuthType != v1pb.DataSource_PASSWORD {
			return errors.Errorf("SSH tunnel (ssh_host, ssh_port, ssh_user, ssh_password, ssh_private_key) is only available when authentication_type is PASSWORD")
		}
		// SSH is only supported for specific engines
		sshSupportedEngines := map[v1pb.Engine]bool{
			v1pb.Engine_MYSQL:     true,
			v1pb.Engine_TIDB:      true,
			v1pb.Engine_MARIADB:   true,
			v1pb.Engine_OCEANBASE: true,
			v1pb.Engine_POSTGRES:  true,
			v1pb.Engine_REDIS:     true,
		}
		if !sshSupportedEngines[engine] {
			return errors.Errorf("SSH tunnel is only available for MYSQL, TIDB, MARIADB, OCEANBASE, POSTGRES, REDIS")
		}
	}

	// Extra connection parameters is only for specific engines
	if len(ds.ExtraConnectionParameters) > 0 {
		extraParamsSupportedEngines := map[v1pb.Engine]bool{
			v1pb.Engine_MYSQL:     true,
			v1pb.Engine_MARIADB:   true,
			v1pb.Engine_OCEANBASE: true,
			v1pb.Engine_POSTGRES:  true,
			v1pb.Engine_ORACLE:    true,
			v1pb.Engine_MSSQL:     true,
			v1pb.Engine_MONGODB:   true,
		}
		if !extraParamsSupportedEngines[engine] {
			return errors.Errorf("extra_connection_parameters is only available for MYSQL, MARIADB, OCEANBASE, POSTGRES, ORACLE, MSSQL, MONGODB")
		}
	}

	return nil
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	return internal.ResourceDelete(ctx, d, c.DeleteInstance)
}
