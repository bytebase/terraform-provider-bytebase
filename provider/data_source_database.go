package provider

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceDatabase() *schema.Resource {
	return &schema.Resource{
		Description: "The database data source.",
		ReadContext: dataSourceDatabaseRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The database full name in instances/{instance}/databases/{database} format",
			},
			"project": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The project full name for the database in projects/{project} format.",
			},
			"environment": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The database environment, will follow the instance environment by default",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The existence of a database.",
			},
			"successful_sync_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The latest synchronization time.",
			},
			"schema_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version of database schema.",
			},
			"labels": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The deployment and policy control labels.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"catalog": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The databases catalog.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"schemas": {
							Computed: true,
							Type:     schema.TypeSet,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"tables": {
										Computed: true,
										Type:     schema.TypeSet,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"classification": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The classification id",
												},
												"columns": {
													Computed: true,
													Type:     schema.TypeSet,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Type:     schema.TypeString,
																Computed: true,
															},
															"semantic_type": {
																Type:        schema.TypeString,
																Computed:    true,
																Description: "The semantic type id",
															},
															"classification": {
																Type:        schema.TypeString,
																Computed:    true,
																Description: "The classification id",
															},
															"labels": {
																Type:     schema.TypeMap,
																Computed: true,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
														},
													},
													Set: func(i interface{}) int {
														return internal.ToHashcodeInt(columnHash(i))
													},
												},
											},
										},
										Set: func(i interface{}) int {
											return internal.ToHashcodeInt(tableHash(i))
										},
									},
								},
							},
							Set: schemaHash,
						},
					},
				},
			},
		},
	}
}

func dataSourceDatabaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	databaseName := d.Get("database").(string)

	database, err := c.GetDatabase(ctx, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(databaseName)

	return setDatabase(ctx, c, d, database)
}

func columnHash(rawColumn interface{}) string {
	var buf bytes.Buffer
	column := rawColumn.(map[string]interface{})

	if v, ok := column["name"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := column["semantic_type"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := column["classification"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := column["classification"].(map[string]interface{}); ok {
		for key, val := range v {
			_, _ = buf.WriteString(fmt.Sprintf("[%s:%s]-", key, val.(string)))
		}
	}
	return buf.String()
}

func tableHash(rawTable interface{}) string {
	var buf bytes.Buffer
	table := rawTable.(map[string]interface{})

	if v, ok := table["name"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := table["classification"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if columns, ok := table["columns"].(*schema.Set); ok {
		for _, column := range columns.List() {
			rawColumn := column.(map[string]interface{})
			_, _ = buf.WriteString(columnHash(rawColumn))
		}
	}

	return buf.String()
}

func schemaHash(rawSchema interface{}) int {
	var buf bytes.Buffer
	raw := rawSchema.(map[string]interface{})

	if v, ok := raw["name"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if tables, ok := raw["tables"].(*schema.Set); ok {
		for _, table := range tables.List() {
			rawTable := table.(map[string]interface{})
			_, _ = buf.WriteString(tableHash(rawTable))
		}
	}

	return internal.ToHashcodeInt(buf.String())
}
