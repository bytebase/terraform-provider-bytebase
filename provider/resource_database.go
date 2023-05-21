package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		Description:   "The database resource.",
		CreateContext: resourceDatabaseCreate,
		ReadContext:   dataSourceDatabaseRead,
		UpdateContext: resourceDatabaseUpdate,
		DeleteContext: resourceDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The project for a database with projects/{project} format.",
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
				Optional:    true,
				Computed:    true,
				Description: "The  deployment and policy control labels.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDatabaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	var diags diag.Diagnostics
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "Only support update the database",
	})

	databaseName := d.Get("name").(string)
	instanceID := d.Get("instance").(string)

	if _, err := c.GetDatabase(ctx, &api.DatabaseFindMessage{
		DatabaseName: d.Get("name").(string),
		InstanceID:   d.Get("instance").(string),
	}); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to find database",
			Detail:   fmt.Sprintf("cannot find instances/%s/databases/%s with error: %v", instanceID, databaseName, err),
		})
		return diags
	}

	project := fmt.Sprintf("projects/%s", d.Get("project").(string))
	labels := map[string]string{}
	for key, val := range d.Get("labels").(map[string]interface{}) {
		labels[key] = val.(string)
	}

	database, err := c.UpdateDatabase(ctx, &api.DatabasePatchMessage{
		Name:    fmt.Sprintf("instances/%s/databases/%s", instanceID, databaseName),
		Project: &project,
		Labels:  &labels,
	})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to update database",
			Detail:   fmt.Sprintf("Failed to patch instances/%s/databases/%s with error: %v", instanceID, databaseName, err),
		})
		return diags
	}

	d.SetId(database.Name)
	diag := dataSourceDatabaseRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("name") {
		return diag.Errorf("cannot change the name")
	}
	if d.HasChange("instance") {
		return diag.Errorf("cannot change the instance")
	}
	if d.HasChange("schema_version") {
		return diag.Errorf("cannot change the schema_version")
	}

	c := m.(api.Client)

	patch := &api.DatabasePatchMessage{
		Name: d.Id(),
	}
	if d.HasChange("project") {
		projectName := fmt.Sprintf("projects/%s", d.Get("project").(string))
		patch.Project = &projectName
	}
	if d.HasChange("labels") {
		labels := d.Get("labels").(map[string]string)
		patch.Labels = &labels
	}

	if _, err := c.UpdateDatabase(ctx, patch); err != nil {
		return diag.FromErr(err)
	}

	return dataSourceDatabaseRead(ctx, d, m)
}

func resourceDatabaseDelete(_ context.Context, d *schema.ResourceData, —— interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "Database is not allowed to delete",
	})

	d.SetId("")

	return diags
}
