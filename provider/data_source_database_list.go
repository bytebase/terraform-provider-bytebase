package provider

import (
	"context"
	"fmt"
	"regexp"
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
			"parent": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: internal.ResourceNameValidation(
					// instance policy
					regexp.MustCompile(fmt.Sprintf("^%s%s$", internal.InstanceNamePrefix, internal.ResourceIDPattern)),
					// project policy
					regexp.MustCompile(fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern)),
				),
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

	databases, err := client.ListDatabase(ctx, parent, "")
	if err != nil {
		return diag.FromErr(err)
	}

	dbList := []interface{}{}
	for _, database := range databases {
		db := map[string]interface{}{}
		db["name"] = database.Name
		db["project"] = database.Project
		db["environment"] = database.Environment
		db["sync_state"] = database.SyncState.String()
		db["successful_sync_time"] = database.SuccessfulSyncTime.AsTime().UTC().Format(time.RFC3339)
		db["schema_version"] = database.SchemaVersion
		db["labels"] = database.Labels

		// catalog, err := client.GetDatabaseCatalog(ctx, database.Name)
		// if err != nil {
		// 	return diag.Errorf("failed to get catalog for database %s with error: %v", database.Name, err.Error())
		// }
		// db["catalog"] = flattenDatabaseCatalog(catalog)

		dbList = append(dbList, db)
	}

	if err := d.Set("databases", dbList); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}
