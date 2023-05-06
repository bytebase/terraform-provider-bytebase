package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
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
			"visibility": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The project visibility.",
			},
			"tenant_mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The project tenant mode.",
			},
			"db_name_template": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The project db name template.",
			},
			"schema_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The project schema version.",
			},
			"schema_change": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The project schema change type.",
			},
		},
	}
}

func dataSourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	project, err := c.GetProject(ctx, d.Get("resource_id").(string), false /* showDeleted */)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(project.Name)
	return setProject(d, project)
}

func setProject(d *schema.ResourceData, project *api.ProjectMessage) diag.Diagnostics {
	projectID, err := internal.GetProjectID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resource_id", projectID); err != nil {
		return diag.Errorf("cannot set resource_id for project: %s", err.Error())
	}
	if err := d.Set("title", project.Title); err != nil {
		return diag.Errorf("cannot set title for project: %s", err.Error())
	}
	if err := d.Set("key", project.Title); err != nil {
		return diag.Errorf("cannot set key for project: %s", err.Error())
	}
	if err := d.Set("title", project.Title); err != nil {
		return diag.Errorf("cannot set title for project: %s", err.Error())
	}
	if err := d.Set("workflow", project.Workflow); err != nil {
		return diag.Errorf("cannot set workflow for project: %s", err.Error())
	}
	if err := d.Set("visibility", project.Visibility); err != nil {
		return diag.Errorf("cannot set visibility for project: %s", err.Error())
	}
	if err := d.Set("tenant_mode", project.TenantMode); err != nil {
		return diag.Errorf("cannot set tenant_mode for project: %s", err.Error())
	}
	if err := d.Set("db_name_template", project.DBNameTemplate); err != nil {
		return diag.Errorf("cannot set db_name_template for project: %s", err.Error())
	}
	if err := d.Set("schema_version", project.SchemaVersion); err != nil {
		return diag.Errorf("cannot set schema_version for project: %s", err.Error())
	}
	if err := d.Set("schema_change", project.SchemaChange); err != nil {
		return diag.Errorf("cannot set schema_change for project: %s", err.Error())
	}

	return nil
}
