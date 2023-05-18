package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceDatabase() *schema.Resource {
	return &schema.Resource{
		Description: "The database data source.",
		ReadContext: dataSourceDatabaseRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The database name.",
			},
			"instance": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The instance resource id for the database.",
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
	}
}

func dataSourceDatabaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	database, err := c.GetDatabase(ctx, &api.DatabaseFindMessage{
		DatabaseName: d.Get("name").(string),
		InstanceID:   d.Get("instance").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(database.Name)
	return setDatabaseMessage(d, database)
}

func setDatabaseMessage(d *schema.ResourceData, database *api.DatabaseMessage) diag.Diagnostics {
	instanceID, databaseName, err := internal.GetInstanceDatabaseID(database.Name)
	if err != nil {
		return diag.Errorf("cannot parse name for database: %s", err.Error())
	}

	if err := d.Set("name", databaseName); err != nil {
		return diag.Errorf("cannot set name for database: %s", err.Error())
	}
	if err := d.Set("instance", instanceID); err != nil {
		return diag.Errorf("cannot set instance for database: %s", err.Error())
	}

	project := database.Project
	if project != "" {
		projectID, err := internal.GetProjectID(project)
		if err != nil {
			return diag.Errorf("failed to parse project id %s with error: %v", database.Project, err.Error())
		}
		project = projectID
	}
	if err := d.Set("project", project); err != nil {
		return diag.Errorf("cannot set project for database: %s", err.Error())
	}

	if err := d.Set("sync_state", database.SyncState); err != nil {
		return diag.Errorf("cannot set sync_state for database: %s", err.Error())
	}
	if err := d.Set("successful_sync_time", database.SuccessfulSyncTime); err != nil {
		return diag.Errorf("cannot set successful_sync_time for database: %s", err.Error())
	}
	if err := d.Set("schema_version", database.SchemaVersion); err != nil {
		return diag.Errorf("cannot set schema_version for database: %s", err.Error())
	}
	if err := d.Set("labels", database.Labels); err != nil {
		return diag.Errorf("cannot set labels for database: %s", err.Error())
	}

	return nil
}
