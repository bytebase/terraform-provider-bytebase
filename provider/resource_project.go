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

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
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

		if existedProject.State == v1pb.State_DELETED {
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

		updateMasks := []string{}
		if title != "" && title != existedProject.Title {
			updateMasks = append(updateMasks, "title")
		}
		if key != "" && key != existedProject.Key {
			updateMasks = append(updateMasks, "key")
		}

		if len(updateMasks) > 0 {
			if _, err := c.UpdateProject(ctx, &v1pb.Project{
				Name:     projectName,
				Title:    title,
				Key:      key,
				State:    v1pb.State_ACTIVE,
				Workflow: existedProject.Workflow,
			}, updateMasks); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to update project",
					Detail:   fmt.Sprintf("Update project %s failed, error: %v", projectName, err),
				})
				return diags
			}
		}
	} else {
		if _, err := c.CreateProject(ctx, projectID, &v1pb.Project{
			Name:     projectName,
			Title:    title,
			Key:      key,
			State:    v1pb.State_ACTIVE,
			Workflow: v1pb.Workflow_UI,
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(projectName)

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
	if existedProject.State == v1pb.State_DELETED {
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

	paths := []string{}
	if d.HasChange("title") {
		paths = append(paths, "title")
	}
	if d.HasChange("key") {
		paths = append(paths, "key")
	}

	if len(paths) > 0 {
		if _, err := c.UpdateProject(ctx, &v1pb.Project{
			Name:     projectName,
			Title:    d.Get("title").(string),
			Key:      d.Get("key").(string),
			State:    v1pb.State_ACTIVE,
			Workflow: existedProject.Workflow,
		}, paths); err != nil {
			diags = append(diags, diag.FromErr(err)...)
			return diags
		}
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
	response, err := c.ListDatabase(ctx, "-", filter)
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
	listDB, err := client.ListDatabase(ctx, "-", filter)
	if err != nil {
		return diag.Errorf("failed to list database with error: %v", err)
	}
	existedDBMap := map[string]*v1pb.Database{}
	for _, db := range listDB.Databases {
		existedDBMap[db.Name] = db
	}

	rawList, ok := d.Get("databases").([]interface{})
	if !ok {
		return nil
	}
	updatedDBMap := map[string]*v1pb.Database{}
	for _, raw := range rawList {
		obj := raw.(map[string]interface{})
		dbName := obj["name"].(string)

		labels := map[string]string{}
		for key, val := range obj["labels"].(map[string]interface{}) {
			labels[key] = val.(string)
		}

		updatedDBMap[dbName] = &v1pb.Database{
			Name:    dbName,
			Project: projectName,
			Labels:  labels,
		}
		if _, err := client.UpdateDatabase(ctx, updatedDBMap[dbName], []string{"project", "label"}); err != nil {
			return diag.Errorf("failed to update database %s with error: %v", dbName, err)
		}
	}

	for _, db := range existedDBMap {
		if _, ok := updatedDBMap[db.Name]; !ok {
			// move db to default project
			if _, err := client.UpdateDatabase(ctx, &v1pb.Database{
				Name:    db.Name,
				Project: projectName,
			}, []string{"project"}); err != nil {
				return diag.Errorf("failed to move database %s to project %s with error: %v", db.Name, defaultProj, err)
			}
		}
	}

	return nil
}
