package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceProjectList() *schema.Resource {
	return &schema.Resource{
		Description: "The project data source list.",
		ReadContext: dataSourceProjectListRead,
		Schema: map[string]*schema.Schema{
			"show_deleted": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Including removed project in the response.",
			},
			"projects": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The project unique resource id.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The project full name in projects/{resource id} format.",
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
				},
			},
		},
	}
}

func dataSourceProjectListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	response, err := c.ListProject(ctx, d.Get("show_deleted").(bool))
	if err != nil {
		return diag.FromErr(err)
	}

	projects := make([]map[string]interface{}, 0)
	for _, project := range response.Projects {
		projectID, err := internal.GetProjectID(project.Name)
		if err != nil {
			return diag.FromErr(err)
		}

		proj := make(map[string]interface{})
		proj["resource_id"] = projectID
		proj["name"] = project.Name
		proj["title"] = project.Title
		proj["key"] = project.Key
		proj["workflow"] = project.Workflow

		filter := fmt.Sprintf(`project == "%s"`, project.Name)
		response, err := c.ListDatabase(ctx, &api.DatabaseFindMessage{
			InstanceID: "-",
			Filter:     &filter,
		})
		if err != nil {
			return diag.FromErr(err)
		}

		dbList := []interface{}{}
		for _, database := range response.Databases {
			db := map[string]interface{}{}
			db["name"] = database.Name
			db["environment"] = database.Environment
			db["sync_state"] = database.SyncState
			db["successful_sync_time"] = database.SuccessfulSyncTime
			db["schema_version"] = database.SchemaVersion
			db["labels"] = database.Labels
			dbList = append(dbList, db)
		}
		proj["databases"] = dbList

		projects = append(projects, proj)
	}

	if err := d.Set("projects", projects); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
