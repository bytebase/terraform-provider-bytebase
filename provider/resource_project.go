package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

// defaultProj is the default project name.
var defaultProj = fmt.Sprintf("%sdefault", internal.ProjectNamePrefix)

func resourceProjct() *schema.Resource {
	return &schema.Resource{
		Description:   "The project resource.",
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		UpdateContext: resourceProjectUpdate,
		DeleteContext: resourceProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The project unique resource id. Cannot change this after created.",
			},
			"title": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The project title.",
			},
			"key": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The project unique key.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The project full name in projects/{resource id} format.",
			},
			"workflow": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The project workflow.",
			},
			"databases": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The databases in the project.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "The database full name in instances/{instance id}/databases/{db name} format.",
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
							Optional:    true,
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

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	projectID := d.Get("resource_id").(string)
	projectName := fmt.Sprintf("%s%s", internal.ProjectNamePrefix, projectID)

	title := d.Get("title").(string)
	key := d.Get("key").(string)

	existedProject, err := c.GetProject(ctx, projectName)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get project %s failed with error: %v", projectName, err))
	}

	var diags diag.Diagnostics
	if existedProject != nil && err == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Project already exists",
			Detail:   fmt.Sprintf("Project %s already exists, try to exec the update operation", projectName),
		})

		if existedProject.State == api.Deleted {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Project is deleted",
				Detail:   fmt.Sprintf("Project %s already deleted, try to undelete the project", projectName),
			})
			if _, err := c.UndeleteProject(ctx, projectName); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to undelete project",
					Detail:   fmt.Sprintf("Undelete project %s failed, error: %v", projectName, err),
				})
				return diags
			}
		}

		project, err := c.UpdateProject(ctx, &api.ProjectPatchMessage{
			Name:  projectName,
			Title: &title,
			Key:   &key,
		})
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update project",
				Detail:   fmt.Sprintf("Update project %s failed, error: %v", projectName, err),
			})
			return diags
		}

		d.SetId(project.Name)
	} else {
		project, err := c.CreateProject(ctx, projectID, &api.ProjectMessage{
			Name:  projectName,
			Title: title,
			Key:   key,
		})
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(project.Name)
	}

	if diag := updateDatabasesInProject(ctx, d, c, d.Id()); diag != nil {
		diags = append(diags, diag...)
		return diags
	}

	diag := resourceProjectRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("resource_id") {
		return diag.Errorf("cannot change the resource id")
	}

	c := m.(api.Client)
	projectName := d.Id()

	existedProject, err := c.GetProject(ctx, projectName)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get project %s failed with error: %v", projectName, err))
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	if existedProject.State == api.Deleted {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Project is deleted",
			Detail:   fmt.Sprintf("Project %s already deleted, try to undelete the project", projectName),
		})
		if _, err := c.UndeleteProject(ctx, projectName); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to undelete project",
				Detail:   fmt.Sprintf("Undelete project %s failed, error: %v", projectName, err),
			})
			return diags
		}
	}

	patch := &api.ProjectPatchMessage{
		Name: projectName,
	}
	if d.HasChange("title") {
		v := d.Get("title").(string)
		patch.Title = &v
	}
	if d.HasChange("key") {
		v := d.Get("key").(string)
		patch.Key = &v
	}

	if _, err := c.UpdateProject(ctx, patch); err != nil {
		diags = append(diags, diag.FromErr(err)...)
		return diags
	}

	if d.HasChange("databases") {
		if diag := updateDatabasesInProject(ctx, d, c, d.Id()); diag != nil {
			diags = append(diags, diag...)
			return diags
		}
	}

	diag := resourceProjectRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	projectName := d.Id()

	project, err := c.GetProject(ctx, projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	filter := fmt.Sprintf(`project == "%s"`, project.Name)
	response, err := c.ListDatabase(ctx, &api.DatabaseFindMessage{
		InstanceID: "-",
		Filter:     &filter,
	})
	if err != nil {
		return diag.Errorf("failed to list database with error: %v", err)
	}

	return setProjectWithDatabases(d, project, response.Databases)
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	projectName := d.Id()

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if err := c.DeleteProject(ctx, projectName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func updateDatabasesInProject(ctx context.Context, d *schema.ResourceData, client api.Client, projectName string) diag.Diagnostics {
	filter := fmt.Sprintf(`project == "%s"`, projectName)
	listDB, err := client.ListDatabase(ctx, &api.DatabaseFindMessage{
		InstanceID: "-",
		Filter:     &filter,
	})
	if err != nil {
		return diag.Errorf("failed to list database with error: %v", err)
	}
	existedDBMap := map[string]*api.DatabaseMessage{}
	for _, db := range listDB.Databases {
		existedDBMap[db.Name] = db
	}

	rawList, ok := d.Get("databases").([]interface{})
	if !ok {
		return nil
	}
	updatedDBMap := map[string]*api.DatabasePatchMessage{}
	for _, raw := range rawList {
		obj := raw.(map[string]interface{})
		dbName := obj["name"].(string)

		labels := map[string]string{}
		for key, val := range obj["labels"].(map[string]interface{}) {
			labels[key] = val.(string)
		}

		// TODO(ed):
		patch := &api.DatabasePatchMessage{
			Name:    dbName,
			Project: &projectName,
			Labels:  &labels,
		}
		updatedDBMap[dbName] = patch
		if _, err := client.UpdateDatabase(ctx, patch); err != nil {
			return diag.Errorf("failed to update database %s with error: %v", patch.Name, err)
		}
	}

	for _, db := range existedDBMap {
		if _, ok := updatedDBMap[db.Name]; !ok {
			// move db to default project
			if _, err := client.UpdateDatabase(ctx, &api.DatabasePatchMessage{
				Name:    db.Name,
				Project: &defaultProj,
			}); err != nil {
				return diag.Errorf("failed to move database %s to project %s with error: %v", db.Name, defaultProj, err)
			}
		}
	}

	return nil
}
