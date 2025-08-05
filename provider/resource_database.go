package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		Description:   "The database resource.",
		CreateContext: resourceDatabaseUpdate,
		ReadContext:   resourceDatabaseRead,
		UpdateContext: resourceDatabaseUpdate,
		DeleteContext: resourceDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: internal.ResourceNameValidation(
					// database name format
					fmt.Sprintf("^%s%s/%s%s$", internal.InstanceNamePrefix, internal.ResourceIDPattern, internal.DatabaseIDPrefix, internal.ResourceIDPattern),
				),
				Description: "The database full name in instances/{instance}/databases/{database} format",
			},
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: internal.ResourceNameValidation(
					// project name format
					fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern),
				),
				Description: "The project full name for the database in projects/{project} format.",
			},
			"environment": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateDiagFunc: internal.ResourceNameValidation(
					// environment name format
					fmt.Sprintf("^%s%s$", internal.EnvironmentNamePrefix, internal.ResourceIDPattern),
				),
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
				Optional:    true,
				Description: "The deployment and policy control labels.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"catalog": {
				Type:        schema.TypeList,
				Computed:    true,
				Optional:    true,
				MaxItems:    1,
				Description: "The databases catalog.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
													Computed:    true,
													Default:     nil,
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
																Computed:    true,
																Description: "The semantic type id",
															},
															"classification": {
																Type:        schema.TypeString,
																Optional:    true,
																Computed:    true,
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
				},
			},
		},
	}
}

func resourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	databaseName := d.Get("name").(string)
	projectName := d.Get("project").(string)

	database := &v1pb.Database{
		Name:    databaseName,
		Project: projectName,
	}
	updateMasks := []string{"project"}
	rawConfig := d.GetRawConfig()
	if config := rawConfig.GetAttr("environment"); !config.IsNull() {
		database.Environment = d.Get("environment").(string)
		updateMasks = append(updateMasks, "environment")
	}
	if config := rawConfig.GetAttr("labels"); !config.IsNull() {
		labels := map[string]string{}
		for key, val := range d.Get("labels").(map[string]interface{}) {
			labels[key] = val.(string)
		}
		database.Labels = labels
		updateMasks = append(updateMasks, "labels")
	}

	if _, err := c.UpdateDatabase(ctx, database, updateMasks); err != nil {
		return diag.Errorf("failed to update the database %s with error: %v", databaseName, err.Error())
	}

	if config := rawConfig.GetAttr("catalog"); !config.IsNull() {
		catalog, err := convertToV1DatabaseCatalog(d, databaseName)
		if err != nil {
			return diag.Errorf("failed to convert database catalog %v with error: %v", databaseName, err.Error())
		}
		if catalog != nil {
			if _, err := c.UpdateDatabaseCatalog(ctx, catalog); err != nil {
				return diag.Errorf("failed to update database catalog %v with error: %v", databaseName, err.Error())
			}
		}
	}

	d.SetId(databaseName)

	return resourceDatabaseRead(ctx, d, m)
}

func resourceDatabaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	databaseName := d.Id()

	database, err := c.GetDatabase(ctx, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	return setDatabase(ctx, c, d, database)
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	databaseName := d.Id()

	var diags diag.Diagnostics

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "Unsupport delete database",
		Detail:   fmt.Sprintf("We don't support delete the database, will transfer the database %s to the default project", databaseName),
	})

	if _, err := c.UpdateDatabase(ctx, &v1pb.Database{
		Name:    databaseName,
		Project: defaultProj,
	}, []string{"project"}); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to unassign database",
			Detail:   fmt.Sprintf("Unassign database %s failed, error: %v", databaseName, err),
		})
		return diags
	}

	d.SetId("")
	return diags
}

func setDatabase(
	ctx context.Context,
	client api.Client,
	d *schema.ResourceData,
	database *v1pb.Database,
) diag.Diagnostics {
	catalog, err := client.GetDatabaseCatalog(ctx, database.Name)
	if err != nil {
		return diag.Errorf("failed to get catalog for database %s with error: %v", database.Name, err.Error())
	}

	d.SetId(database.Name)
	if err := d.Set("project", database.Project); err != nil {
		return diag.Errorf("cannot set project for database: %s", err.Error())
	}
	if err := d.Set("environment", database.EffectiveEnvironment); err != nil {
		return diag.Errorf("cannot set environment for database: %s", err.Error())
	}
	if err := d.Set("state", database.State.String()); err != nil {
		return diag.Errorf("cannot set state for database: %s", err.Error())
	}
	if err := d.Set("successful_sync_time", database.SuccessfulSyncTime.AsTime().UTC().Format(time.RFC3339)); err != nil {
		return diag.Errorf("cannot set successful_sync_time for database: %s", err.Error())
	}
	if err := d.Set("schema_version", database.SchemaVersion); err != nil {
		return diag.Errorf("cannot set schema_version for database: %s", err.Error())
	}
	if err := d.Set("labels", database.Labels); err != nil {
		return diag.Errorf("cannot set labels for database: %s", err.Error())
	}

	if err := d.Set("catalog", flattenDatabaseCatalog(catalog)); err != nil {
		return diag.Errorf("cannot set catalog for database: %s", err.Error())
	}

	return nil
}

func flattenDatabaseCatalog(catalog *v1pb.DatabaseCatalog) []interface{} {
	schemaList := []interface{}{}
	for _, schemaCatalog := range catalog.Schemas {
		rawSchema := map[string]interface{}{
			"name": schemaCatalog.Name,
		}

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

	rawCatalog := map[string]interface{}{
		"schemas": schema.NewSet(schemaHash, schemaList),
	}
	return []interface{}{rawCatalog}
}

func convertToV1DatabaseCatalog(d *schema.ResourceData, databaseName string) (*v1pb.DatabaseCatalog, error) {
	catalogs, ok := d.Get("catalog").([]interface{})
	if !ok || len(catalogs) != 1 {
		return nil, nil
	}
	rawCatalog := catalogs[0].(map[string]interface{})

	rawSchemaList, ok := rawCatalog["schemas"].(*schema.Set)
	if !ok {
		return nil, errors.Errorf("invalid schemas")
	}

	catalog := &v1pb.DatabaseCatalog{
		Name:    fmt.Sprintf("%s%s", databaseName, internal.DatabaseCatalogNameSuffix),
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
