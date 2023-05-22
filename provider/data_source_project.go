package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceProject() *schema.Resource {
	return &schema.Resource{
		Description: "The project data source.",
		ReadContext: dataSourceProjectRead,
		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The project unique resource id.",
			},
			"title": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The project title.",
			},
			"key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The project key.",
			},
			"workflow": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The project workflow.",
			},
			"visibility": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The project visibility.",
			},
			"tenant_mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The project tenant mode.",
			},
			"db_name_template": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The project db name template.",
			},
			"schema_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The project schema version.",
			},
			"schema_change": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The project schema change type.",
			},
			"databases": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The databases in the project.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The database name.",
						},
						"instance": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The instance resource id for the database.",
						},
						"sync_state": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The existence of a database on latest sync.",
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
							Description: "The  deployment and policy control labels.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataSourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	project, err := c.GetProject(ctx, d.Get("resource_id").(string), false /* showDeleted */)
	if err != nil {
		return diag.FromErr(err)
	}

	filter := fmt.Sprintf(`project == "%s"`, project.Name)
	response, err := c.ListDatabase(ctx, &api.DatabaseFindMessage{
		InstanceID: "-",
		Filter:     &filter,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(project.Name)
	return setProjectWithDatabases(d, project, response.Databases)
}

func setProjectWithDatabases(d *schema.ResourceData, project *api.ProjectMessage, databases []*api.DatabaseMessage) diag.Diagnostics {
	projectID, err := internal.GetProjectID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resource_id", projectID); err != nil {
		return diag.Errorf("cannot set resource_id for project: %s", err.Error())
	}
	if err := d.Set("title", project.Title); err != nil {
		return diag.Errorf("cannot set title for project: %s", err.Error())
	}
	if err := d.Set("key", project.Key); err != nil {
		return diag.Errorf("cannot set key for project: %s", err.Error())
	}
	if err := d.Set("title", project.Title); err != nil {
		return diag.Errorf("cannot set title for project: %s", err.Error())
	}
	if err := d.Set("workflow", project.Workflow); err != nil {
		return diag.Errorf("cannot set workflow for project: %s", err.Error())
	}
	if err := d.Set("visibility", project.Visibility); err != nil {
		return diag.Errorf("cannot set visibility for project: %s", err.Error())
	}
	if err := d.Set("tenant_mode", project.TenantMode); err != nil {
		return diag.Errorf("cannot set tenant_mode for project: %s", err.Error())
	}
	if err := d.Set("db_name_template", project.DBNameTemplate); err != nil {
		return diag.Errorf("cannot set db_name_template for project: %s", err.Error())
	}
	if err := d.Set("schema_version", project.SchemaVersion); err != nil {
		return diag.Errorf("cannot set schema_version for project: %s", err.Error())
	}
	if err := d.Set("schema_change", project.SchemaChange); err != nil {
		return diag.Errorf("cannot set schema_change for project: %s", err.Error())
	}

	dbList := []interface{}{}
	for _, database := range databases {
		instanceID, databaseName, err := internal.GetInstanceDatabaseID(database.Name)
		if err != nil {
			return diag.FromErr(err)
		}

		db := map[string]interface{}{}
		db["name"] = databaseName
		db["instance"] = instanceID
		db["sync_state"] = database.SyncState
		db["successful_sync_time"] = database.SuccessfulSyncTime
		db["schema_version"] = database.SchemaVersion
		db["labels"] = database.Labels
		dbList = append(dbList, db)
	}
	if err := d.Set("databases", dbList); err != nil {
		return diag.Errorf("cannot set databases for project: %s", err.Error())
	}

	return nil
}
