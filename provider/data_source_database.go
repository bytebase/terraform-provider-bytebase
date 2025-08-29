package provider

import (
	"context"

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
													Set: columnHash,
												},
											},
										},
										Set: tableHash,
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
	databaseName := d.Get("name").(string)

	database, err := c.GetDatabase(ctx, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(databaseName)

	return setDatabase(ctx, c, d, database)
}

func columnHash(rawColumn interface{}) int {
	column := convertToV1ColumnCatalog(rawColumn)
	return internal.ToHash(column)
}

func tableHash(rawTable interface{}) int {
	table, err := convertToV1TableCatalog(rawTable)
	if err != nil {
		return 0
	}
	return internal.ToHash(table)
}

func schemaHash(rawSchema interface{}) int {
	schema, err := convertToV1SchemaCatalog(rawSchema)
	if err != nil {
		return 0
	}
	return internal.ToHash(schema)
}
