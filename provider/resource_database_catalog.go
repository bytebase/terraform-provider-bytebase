package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func resourceDatabaseCatalog() *schema.Resource {
	return &schema.Resource{
		Description:   "The database catalog resource.",
		CreateContext: resourceDatabaseCatalogCreate,
		ReadContext:   resourceDatabaseCatalogRead,
		UpdateContext: resourceDatabaseCatalogUpdate,
		DeleteContext: resourceDatabaseCatalogDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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
				Required: true,
				Type:     schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"tables": {
							Required: true,
							Type:     schema.TypeSet,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"classification": {
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "",
										Description: "The classification id",
									},
									"columns": {
										Required: true,
										Type:     schema.TypeSet,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"semantic_type": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "The semantic type id",
												},
												"classification": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "The classification id",
												},
												"labels": {
													Type:     schema.TypeMap,
													Optional: true,
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

func resourceDatabaseCatalogRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	catelogName := d.Id()
	database := getDatabaseFullNameFromCatalog(catelogName)

	catalog, err := c.GetDatabaseCatalog(ctx, database)
	if err != nil {
		return diag.FromErr(err)
	}

	return setDatabaseCatalog(d, catalog)
}

func getDatabaseFullNameFromCatalog(catalog string) string {
	return strings.TrimSuffix(catalog, internal.DatabaseCatalogNameSuffix)
}

func resourceDatabaseCatalogCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	database := d.Get("database").(string)

	catalog, err := convertToDatabaseCatalog(d)
	if err != nil {
		return diag.Errorf("failed to convert catalog %v with error: %v", database, err.Error())
	}

	if _, err := c.UpdateDatabaseCatalog(ctx, catalog, []string{}); err != nil {
		return diag.Errorf("failed to update catalog %v with error: %v", database, err.Error())
	}

	d.SetId(catalog.Name)

	var diags diag.Diagnostics
	diag := resourceDatabaseCatalogRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceDatabaseCatalogUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	catelogName := d.Id()
	database := getDatabaseFullNameFromCatalog(catelogName)

	catalog, err := convertToDatabaseCatalog(d)
	if err != nil {
		return diag.Errorf("failed to convert catalog %v with error: %v", database, err.Error())
	}

	if _, err := c.UpdateDatabaseCatalog(ctx, catalog, []string{}); err != nil {
		return diag.Errorf("failed to update catalog %v with error: %v", database, err.Error())
	}

	if _, err := c.UpdateDatabaseCatalog(ctx, catalog, []string{}); err != nil {
		return diag.Errorf("failed to update catalog %v with error: %v", database, err.Error())
	}

	var diags diag.Diagnostics
	diag := resourceDatabaseCatalogRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceDatabaseCatalogDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}

func convertToDatabaseCatalog(d *schema.ResourceData) (*v1pb.DatabaseCatalog, error) {
	database, ok := d.Get("database").(string)
	if !ok || database == "" {
		return nil, errors.Errorf("invalid database")
	}
	rawSchemaList, ok := d.Get("schemas").(*schema.Set)
	if !ok {
		return nil, errors.Errorf("invalid schemas")
	}

	catalog := &v1pb.DatabaseCatalog{
		Name:    fmt.Sprintf("%s%s", database, internal.DatabaseCatalogNameSuffix),
		Schemas: []*v1pb.SchemaCatalog{},
	}

	for _, raw := range rawSchemaList.List() {
		rawSchema := raw.(map[string]interface{})
		schemaCatalog := &v1pb.SchemaCatalog{
			Name: rawSchema["name"].(string),
		}

		rawTableList, ok := rawSchema["tables"].(*schema.Set)
		if !ok {
			return nil, errors.Errorf("invalid tables")
		}
		for _, table := range rawTableList.List() {
			rawTable := table.(map[string]interface{})
			table := &v1pb.TableCatalog{
				Name:           rawTable["name"].(string),
				Classification: rawTable["classification"].(string),
			}

			columnList := []*v1pb.ColumnCatalog{}
			rawColumnList, ok := rawTable["columns"].(*schema.Set)
			if !ok {
				return nil, errors.Errorf("invalid columns")
			}
			for _, column := range rawColumnList.List() {
				rawColumn := column.(map[string]interface{})
				labels := map[string]string{}
				for key, val := range rawColumn["labels"].(map[string]interface{}) {
					labels[key] = val.(string)
				}

				column := &v1pb.ColumnCatalog{
					Name:           rawColumn["name"].(string),
					SemanticType:   rawColumn["semantic_type"].(string),
					Classification: rawColumn["classification"].(string),
					Labels:         labels,
				}
				columnList = append(columnList, column)
			}

			table.Kind = &v1pb.TableCatalog_Columns_{
				Columns: &v1pb.TableCatalog_Columns{
					Columns: columnList,
				},
			}

			schemaCatalog.Tables = append(schemaCatalog.Tables, table)
		}

		catalog.Schemas = append(catalog.Schemas, schemaCatalog)
	}

	return catalog, nil
}
