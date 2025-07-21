package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceDatabaseList() *schema.Resource {
	return &schema.Resource{
		Description: "The database data source list.",
		ReadContext: dataSourceDatabaseListRead,
		Schema: map[string]*schema.Schema{
			"parent": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: internal.ResourceNameValidation(
					fmt.Sprintf("^%s$", internal.WorkspaceName),
					fmt.Sprintf("^%s%s$", internal.InstanceNamePrefix, internal.ResourceIDPattern),
					fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern),
				),
			},
			"query": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter databases by name with wildcard",
			},
			"exclude_unassigned": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "If not include unassigned databases in the response.",
			},
			"environment": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The environment full name. Filter databases by environment.",
				ValidateDiagFunc: internal.ResourceNameValidation(
					fmt.Sprintf("^%s%s$", internal.EnvironmentNamePrefix, internal.ResourceIDPattern),
				),
			},
			"project": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The project full name. Filter databases by project.",
				ValidateDiagFunc: internal.ResourceNameValidation(
					fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern),
				),
			},
			"instance": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The instance full name. Filter databases by instance.",
				ValidateDiagFunc: internal.ResourceNameValidation(
					fmt.Sprintf("^%s%s$", internal.InstanceNamePrefix, internal.ResourceIDPattern),
				),
			},
			"engines": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: internal.EngineValidation,
				},
				Description: "Filter databases by engines.",
			},
			"labels": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Filter databases by labels",
			},
			"databases": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
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
					},
				},
			},
		},
	}
}

func dataSourceDatabaseListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(api.Client)
	parent := d.Get("parent").(string)

	filter := &api.DatabaseFilter{
		Query:             d.Get("query").(string),
		Environment:       d.Get("environment").(string),
		Project:           d.Get("project").(string),
		Instance:          d.Get("instance").(string),
		ExcludeUnassigned: d.Get("exclude_unassigned").(bool),
	}

	engines := d.Get("engines").(*schema.Set)
	for _, engine := range engines.List() {
		engineString := engine.(string)
		engineValue, ok := v1pb.Engine_value[engineString]
		if ok {
			filter.Engines = append(filter.Engines, v1pb.Engine(engineValue))
		}
	}
	for key, val := range d.Get("labels").(map[string]interface{}) {
		filter.Labels = append(filter.Labels, &api.Label{
			Key:   key,
			Value: val.(string),
		})
	}

	databases, err := client.ListDatabase(ctx, parent, filter, true)
	if err != nil {
		return diag.FromErr(err)
	}

	dbList := []interface{}{}
	for _, database := range databases {
		db := map[string]interface{}{}
		db["name"] = database.Name
		db["project"] = database.Project
		db["environment"] = database.Environment
		db["state"] = database.State.String()
		db["successful_sync_time"] = database.SuccessfulSyncTime.AsTime().UTC().Format(time.RFC3339)
		db["schema_version"] = database.SchemaVersion
		db["labels"] = database.Labels

		dbList = append(dbList, db)
	}

	if err := d.Set("databases", dbList); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}
