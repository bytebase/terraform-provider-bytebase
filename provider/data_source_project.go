package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
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
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The project full name in projects/{resource id} format.",
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
			"databases": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The databases in the project.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The database full name in instances/{instance id}/databases/{db name} format.",
						},
						"environment": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The database environment.",
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
	projectName := fmt.Sprintf("%s%s", internal.ProjectNamePrefix, d.Get("resource_id").(string))
	project, err := c.GetProject(ctx, projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	filter := fmt.Sprintf(`project == "%s"`, project.Name)
	response, err := c.ListDatabase(ctx, "-", filter)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(project.Name)
	return setProjectWithDatabases(d, project, response.Databases)
}

func setProjectWithDatabases(d *schema.ResourceData, project *v1pb.Project, databases []*v1pb.Database) diag.Diagnostics {
	d.SetId(project.Name)

	projectID, err := internal.GetProjectID(project.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resource_id", projectID); err != nil {
		return diag.Errorf("cannot set resource_id for project: %s", err.Error())
	}
	if err := d.Set("title", project.Title); err != nil {
		return diag.Errorf("cannot set title for project: %s", err.Error())
	}
	if err := d.Set("name", project.Name); err != nil {
		return diag.Errorf("cannot set name for project: %s", err.Error())
	}
	if err := d.Set("key", project.Key); err != nil {
		return diag.Errorf("cannot set key for project: %s", err.Error())
	}
	if err := d.Set("title", project.Title); err != nil {
		return diag.Errorf("cannot set title for project: %s", err.Error())
	}
	if err := d.Set("workflow", project.Workflow.String()); err != nil {
		return diag.Errorf("cannot set workflow for project: %s", err.Error())
	}

	dbList := []interface{}{}
	for _, database := range databases {
		db := map[string]interface{}{}
		db["name"] = database.Name
		db["environment"] = database.Environment
		db["sync_state"] = database.SyncState.String()
		db["successful_sync_time"] = database.SuccessfulSyncTime.AsTime().UTC().Format(time.RFC3339)
		db["schema_version"] = database.SchemaVersion
		db["labels"] = database.Labels
		dbList = append(dbList, db)
	}
	if err := d.Set("databases", dbList); err != nil {
		return diag.Errorf("cannot set databases for project: %s", err.Error())
	}

	return nil
}
