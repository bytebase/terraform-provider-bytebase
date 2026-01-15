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
		Description:        "The instance data source.",
		ReadWithoutTimeout: dataSourceInstanceRead,
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
			"activation": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether assign license for this instance or not.",
			},
			"engine": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance engine. Supported engines: MYSQL, POSTGRES, TIDB, SNOWFLAKE, CLICKHOUSE, MONGODB, SQLITE, REDIS, ORACLE, SPANNER, MSSQL, REDSHIFT, MARIADB, OCEANBASE, STARROCKS, DORIS, HIVE, ELASTICSEARCH, BIGQUERY, DYNAMODB, DATABRICKS, COCKROACHDB, COSMOSDB, TRINO, CASSANDRA.",
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
			"data_sources": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: getDataSourceComputedSchema(),
				},
				Set: dataSourceHash,
			},
			"list_all_databases": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "List all databases in this instance. If false, will only list 500 databases.",
			},
			"databases":      getDatabasesSchema(true),
			"sync_databases": getSyncDatabasesSchema(true),
		},
	}
}

// getDataSourceComputedSchema returns the schema for data_sources in data source (read-only).
func getDataSourceComputedSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
			Description: "Enable SSL connection.",
		},
		"ssl_ca": {
			Type:        schema.TypeString,
			Computed:    true,
			Sensitive:   true,
			Description: "The CA certificate.",
		},
		"ssl_cert": {
			Type:        schema.TypeString,
			Computed:    true,
			Sensitive:   true,
			Description: "The client certificate.",
		},
		"ssl_key": {
			Type:        schema.TypeString,
			Computed:    true,
			Sensitive:   true,
			Description: "The client key.",
		},
		"verify_tls_certificate": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Enable TLS certificate verification for SSL connections.",
		},
		// MongoDB-specific
		"srv": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Use DNS SRV record for MongoDB connection.",
		},
		"authentication_database": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The database to authenticate against for MongoDB.",
		},
		"replica_set": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The replica set name for MongoDB.",
		},
		"direct_connection": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Use direct connection to MongoDB node.",
		},
		"additional_addresses": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Additional addresses for MongoDB replica set.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"host": {Type: schema.TypeString, Computed: true},
					"port": {Type: schema.TypeString, Computed: true},
				},
			},
		},
		// Oracle-specific
		"sid": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Oracle System Identifier (SID).",
		},
		"service_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Oracle service name.",
		},
		// SSH Tunneling
		"ssh_host": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "SSH tunnel server hostname.",
		},
		"ssh_port": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "SSH tunnel server port.",
		},
		"ssh_user": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "SSH tunnel username.",
		},
		"ssh_password": {
			Type:        schema.TypeString,
			Computed:    true,
			Sensitive:   true,
			Description: "SSH tunnel password.",
		},
		"ssh_private_key": {
			Type:        schema.TypeString,
			Computed:    true,
			Sensitive:   true,
			Description: "SSH tunnel private key.",
		},
		// Authentication
		"authentication_type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Authentication type: PASSWORD, GOOGLE_CLOUD_SQL_IAM, AWS_RDS_IAM, AZURE_IAM.",
		},
		"authentication_private_key": {
			Type:        schema.TypeString,
			Computed:    true,
			Sensitive:   true,
			Description: "PKCS#8 private key for authentication.",
		},
		"authentication_private_key_passphrase": {
			Type:        schema.TypeString,
			Computed:    true,
			Sensitive:   true,
			Description: "Passphrase for encrypted private key.",
		},
		// Redis-specific
		"redis_type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Redis deployment type: STANDALONE, SENTINEL, CLUSTER.",
		},
		"master_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Redis Sentinel master name.",
		},
		"master_username": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Redis Sentinel master username.",
		},
		"master_password": {
			Type:        schema.TypeString,
			Computed:    true,
			Sensitive:   true,
			Description: "Redis Sentinel master password.",
		},
		// Cloud-specific
		"region": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "AWS region for IAM auth.",
		},
		"warehouse_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Databricks warehouse ID.",
		},
		"cluster": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "CockroachDB cluster name.",
		},
		// IAM Credentials
		"azure_credential": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Azure IAM credential.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"tenant_id":     {Type: schema.TypeString, Computed: true},
					"client_id":     {Type: schema.TypeString, Computed: true},
					"client_secret": {Type: schema.TypeString, Computed: true, Sensitive: true},
				},
			},
		},
		"aws_credential": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "AWS IAM credential.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"access_key_id":     {Type: schema.TypeString, Computed: true},
					"secret_access_key": {Type: schema.TypeString, Computed: true, Sensitive: true},
					"session_token":     {Type: schema.TypeString, Computed: true, Sensitive: true},
					"role_arn":          {Type: schema.TypeString, Computed: true},
					"external_id":       {Type: schema.TypeString, Computed: true},
				},
			},
		},
		"gcp_credential": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "GCP IAM credential (service account JSON).",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"content": {Type: schema.TypeString, Computed: true, Sensitive: true},
				},
			},
		},
		// SASL Config
		"sasl_config": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "SASL authentication configuration for HIVE.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"kerberos": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"primary":                {Type: schema.TypeString, Computed: true},
								"instance":               {Type: schema.TypeString, Computed: true},
								"realm":                  {Type: schema.TypeString, Computed: true},
								"keytab":                 {Type: schema.TypeString, Computed: true, Sensitive: true},
								"kdc_host":               {Type: schema.TypeString, Computed: true},
								"kdc_port":               {Type: schema.TypeString, Computed: true},
								"kdc_transport_protocol": {Type: schema.TypeString, Computed: true},
							},
						},
					},
				},
			},
		},
		// Extra connection parameters
		"extra_connection_parameters": {
			Type:        schema.TypeMap,
			Computed:    true,
			Description: "Extra connection parameters as key-value pairs.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}
}

func getExternalSecretSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "The external secret to get the database password. Learn more: https://www.bytebase.com/docs/get-started/instance/#use-external-secret-manager",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"vault": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "The Valut to get the database password. Reference doc https://developer.hashicorp.com/vault/api-docs/secret/kv/kv-v2",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"url": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The Vault URL.",
							},
							"token": {
								Type:        schema.TypeString,
								Computed:    true,
								Sensitive:   true,
								Description: "The root token without TTL. Learn more: https://developer.hashicorp.com/vault/docs/commands/operator/generate-root",
							},
							"app_role": {
								Type:        schema.TypeList,
								Computed:    true,
								Description: "The Vault app role to get the password.",
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"role_id": {
											Type:        schema.TypeString,
											Computed:    true,
											Sensitive:   true,
											Description: "The app role id.",
										},
										"secret": {
											Type:        schema.TypeString,
											Computed:    true,
											Sensitive:   true,
											Description: "The secret id for the role without ttl.",
										},
										"secret_type": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "The secret id type, can be PLAIN (plain text for the secret) or ENVIRONMENT (envirionment name for the secret).",
										},
									},
								},
							},
							"engine_name": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The name for secret engine.",
							},
							"secret_name": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The secret name in the engine to store the password.",
							},
							"password_key_name": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The key name for the password.",
							},
						},
					},
				},
				"aws_secrets_manager": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "The AWS Secrets Manager to get the database password. Reference doc https://docs.aws.amazon.com/secretsmanager/latest/userguide/intro.html",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"secret_name": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The secret name to store the password.",
							},
							"password_key_name": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The key name for the password.",
							},
						},
					},
				},
				"gcp_secret_manager": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "The GCP Secret Manager to get the database password. Reference doc https://cloud.google.com/secret-manager/docs",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"secret_name": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The secret name should be like \"projects/{project-id}/secrets/{secret-id}\".",
							},
						},
					},
				},
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

	return setInstanceMessage(ctx, c, d, ins)
}
