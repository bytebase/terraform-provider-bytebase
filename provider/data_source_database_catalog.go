package provider

import (
	"bytes"
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceDatabaseCatalog() *schema.Resource {
	return &schema.Resource{
		Description: "The database catalog data source.",
		ReadContext: dataSourceDatabaseCatalogRead,
		Schema: map[string]*schema.Schema{
			"database": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The database full name in instances/{instance}/databases/{database} format",
				ValidateDiagFunc: internal.ResourceNameValidation(
					regexp.MustCompile(fmt.Sprintf("^%s%s/%s%s$", internal.InstanceNamePrefix, internal.ResourceIDPattern, internal.DatabaseIDPrefix, internal.ResourceIDPattern)),
				),
			},
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
	}
}

func dataSourceDatabaseCatalogRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	database := d.Get("database").(string)

	catalog, err := c.GetDatabaseCatalog(ctx, database)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(catalog.Name)

	return setDatabaseCatalog(d, catalog)
}

func setDatabaseCatalog(d *schema.ResourceData, catalog *v1pb.DatabaseCatalog) diag.Diagnostics {
	database := getDatabaseFullNameFromCatalog(catalog.Name)
	if err := d.Set("database", database); err != nil {
		return diag.Errorf("cannot set database: %s", err.Error())
	}

	schemaList := []interface{}{}
	for _, schemaCatalog := range catalog.Schemas {
		rawSchema := map[string]interface{}{}

		tableList := []interface{}{}
		for _, table := range schemaCatalog.Tables {
			rawTable := map[string]interface{}{}
			rawTable["name"] = table.Name
			rawTable["classification"] = table.Classification

			columnList := []interface{}{}
			for _, column := range table.GetColumns().Columns {
				rawColumn := map[string]interface{}{}
				rawColumn["name"] = column.Name
				rawColumn["semantic_type"] = column.SemanticType
				rawColumn["classification"] = column.Classification
				rawColumn["labels"] = column.Labels
				columnList = append(columnList, rawColumn)
			}
			rawTable["columns"] = schema.NewSet(func(i interface{}) int {
				return internal.ToHashcodeInt(columnHash(i))
			}, columnList)
			tableList = append(tableList, rawTable)
		}
		rawSchema["tables"] = schema.NewSet(func(i interface{}) int {
			return internal.ToHashcodeInt(tableHash(i))
		}, tableList)
		schemaList = append(schemaList, rawSchema)
	}

	if err := d.Set("schemas", schema.NewSet(schemaHash, schemaList)); err != nil {
		return diag.Errorf("cannot set schemas: %s", err.Error())
	}
	return nil
}

func columnHash(rawColumn interface{}) string {
	var buf bytes.Buffer
	column := rawColumn.(map[string]interface{})

	if v, ok := column["name"].(string); ok {
		_, _ = buf.WriteString(fmt.Sprintf("%s-", v))
	}

	return buf.String()
}

func tableHash(rawTable interface{}) string {
	var buf bytes.Buffer
	table := rawTable.(map[string]interface{})

	if v, ok := table["name"].(string); ok {
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
