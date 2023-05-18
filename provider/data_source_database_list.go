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

func dataSourceDatabaseList() *schema.Resource {
	return &schema.Resource{
		Description: "The database data source list.",
		ReadContext: dataSourceDatabaseListRead,
		Schema: map[string]*schema.Schema{
			"instance": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "-",
				Description: `The instance resource id for the database. You can use "-" to list databases in all instances.`,
			},
			"project": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "-",
				Description: `The project resource id for the database. You can use "-" to list databases in all projects.`,
			},
			"databases": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The database name.",
						},
						"instance": {
							Type:     schema.TypeString,
							Computed: true,

							Description: "The instance resource id for the database.",
						},
						"project": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The project for a database with projects/{project} format.",
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

func dataSourceDatabaseListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	project := d.Get("project").(string)
	find := &api.DatabaseFindMessage{
		InstanceID: d.Get("instance").(string),
	}
	if project != "" && project != "-" {
		filter := fmt.Sprintf(`project = "projects/%s".`, project)
		find.Filter = &filter
	}
	response, err := c.ListDatabase(ctx, find)
	if err != nil {
		return diag.FromErr(err)
	}

	databases := make([]map[string]interface{}, 0)
	for _, database := range response.Databases {
		instanceID, databaseName, err := internal.GetInstanceDatabaseID(database.Name)
		if err != nil {
			return diag.FromErr(err)
		}

		db := make(map[string]interface{})
		db["name"] = databaseName
		db["instance"] = instanceID

		project := database.Project
		if project != "" {
			projectID, err := internal.GetProjectID(project)
			if err != nil {
				return diag.Errorf("failed to parse project id %s with error", database.Project, err.Error())
			}
			project = projectID
		}
		db["project"] = project
		db["sync_state"] = database.SyncState
		db["successful_sync_time"] = database.SuccessfulSyncTime
		db["schema_version"] = database.SchemaVersion
		db["labels"] = database.Labels

		databases = append(databases, db)
	}

	if err := d.Set("databases", databases); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}
