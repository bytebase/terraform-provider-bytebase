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
var defaultProj = "projects/default"

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
			"workflow": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(api.ProjectWorkflowUI),
					string(api.ProjectWorkflowVCS),
				}, false),
				Description: "The project workflow.",
			},
			"visibility": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  string(api.ProjectVisibilityPublic),
				ValidateFunc: validation.StringInSlice([]string{
					string(api.ProjectVisibilityPublic),
					string(api.ProjectVisibilityPrivate),
				}, false),
				Description: "The project visibility. Cannot change this after created",
			},
			"tenant_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  string(api.ProjectTenantModeDisabled),
				ValidateFunc: validation.StringInSlice([]string{
					string(api.ProjectTenantModeDisabled),
					string(api.ProjectTenantModeEnabled),
				}, false),
				Description: "The project tenant mode.",
			},
			"db_name_template": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The project db name template.",
			},
			"schema_version": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(api.ProjectSchemaVersionTimestamp),
					string(api.ProjectSchemaVersionSemantic),
					string(api.ProjectSchemaVersionUnspecified),
				}, false),
				Description: "The project schema version type. Cannot change this after created.",
			},
			"schema_change": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(api.ProjectSchemaChangeDDL),
					string(api.ProjectSchemaChangeSDL),
				}, false),
				Description: "The project schema change type.",
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
							Description:  "The database name.",
						},
						"instance": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: internal.ResourceIDValidation,
							Description:  "The instance resource id for the database.",
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
	projectName := fmt.Sprintf("projects/%s", projectID)

	title := d.Get("title").(string)
	key := d.Get("key").(string)
	workflow := api.ProjectWorkflow(d.Get("workflow").(string))
	tenantMode := api.ProjectTenantMode(d.Get("tenant_mode").(string))
	dbNameTemplate := d.Get("db_name_template").(string)
	schemaChange := api.ProjectSchemaChange(d.Get("schema_change").(string))
	visibility := api.ProjectVisibility(d.Get("visibility").(string))
	schemaVersion := api.ProjectSchemaVersion(d.Get("schema_version").(string))

	existedProject, err := c.GetProject(ctx, projectID, true /* showDeleted */)
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

		if existedProject.Visibility != visibility {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid argument",
				Detail:   fmt.Sprintf("cannot update project %s visibility to %s", projectName, visibility),
			})
			return diags
		}
		if existedProject.SchemaVersion != schemaVersion {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid argument",
				Detail:   fmt.Sprintf("cannot update project %s schema_version to %s", projectName, schemaVersion),
			})
			return diags
		}

		if existedProject.State == api.Deleted {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Project is deleted",
				Detail:   fmt.Sprintf("Project %s already deleted, try to undelete the project", projectName),
			})
			if _, err := c.UndeleteProject(ctx, projectID); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to undelete project",
					Detail:   fmt.Sprintf("Undelete project %s failed, error: %v", projectName, err),
				})
				return diags
			}
		}

		project, err := c.UpdateProject(ctx, projectID, &api.ProjectPatchMessage{
			Title:          &title,
			Key:            &key,
			Workflow:       &workflow,
			TenantMode:     &tenantMode,
			DBNameTemplate: &dbNameTemplate,
			SchemaChange:   &schemaChange,
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
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Project not exists",
			Detail:   fmt.Sprintf("Project %s not exists, try to exec the create operation", projectName),
		})
		project, err := c.CreateProject(ctx, projectID, &api.ProjectMessage{
			Title:          title,
			Key:            key,
			Workflow:       workflow,
			Visibility:     visibility,
			TenantMode:     tenantMode,
			DBNameTemplate: dbNameTemplate,
			SchemaVersion:  schemaVersion,
			SchemaChange:   schemaChange,
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
	if d.HasChange("visibility") {
		return diag.Errorf("cannot change the visibility in project")
	}
	if d.HasChange("schema_version") {
		return diag.Errorf("cannot change the schema_version in project")
	}

	c := m.(api.Client)

	projectID, err := internal.GetProjectID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	projectName := fmt.Sprintf("projects/%s", projectID)

	existedProject, err := c.GetProject(ctx, projectID, true /* showDeleted */)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get project %s failed with error: %v", projectName, err))
	}
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	if existedProject.State == api.Deleted {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Project is deleted",
			Detail:   fmt.Sprintf("Project %s already deleted, try to undelete the project", projectName),
		})
		if _, err := c.UndeleteProject(ctx, projectID); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to undelete project",
				Detail:   fmt.Sprintf("Undelete project %s failed, error: %v", projectName, err),
			})
			return diags
		}
	}

	patch := &api.ProjectPatchMessage{}
	if d.HasChange("title") {
		v := d.Get("title").(string)
		patch.Title = &v
	}
	if d.HasChange("key") {
		v := d.Get("key").(string)
		patch.Key = &v
	}
	if d.HasChange("db_name_template") {
		v := d.Get("db_name_template").(string)
		patch.DBNameTemplate = &v
	}
	if d.HasChange("workflow") {
		v := api.ProjectWorkflow(d.Get("workflow").(string))
		patch.Workflow = &v
	}
	if d.HasChange("schema_change") {
		v := api.ProjectSchemaChange(d.Get("schema_change").(string))
		patch.SchemaChange = &v
	}
	if d.HasChange("tenant_mode") {
		v := api.ProjectTenantMode(d.Get("tenant_mode").(string))
		patch.TenantMode = &v
	}

	if _, err := c.UpdateProject(ctx, projectID, patch); err != nil {
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

	projectID, err := internal.GetProjectID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	project, err := c.GetProject(ctx, projectID, false /* showDeleted */)
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

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	projectID, err := internal.GetProjectID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := c.DeleteProject(ctx, projectID); err != nil {
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
		instance := obj["instance"].(string)

		labels := map[string]string{}
		for key, val := range obj["labels"].(map[string]interface{}) {
			labels[key] = val.(string)
		}

		name := fmt.Sprintf("instances/%s/databases/%s", instance, dbName)
		patch := &api.DatabasePatchMessage{
			Name:    name,
			Project: &projectName,
			Labels:  &labels,
		}
		updatedDBMap[name] = patch
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
