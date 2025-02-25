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
