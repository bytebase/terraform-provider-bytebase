package provider

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceVCSConnector() *schema.Resource {
	return &schema.Resource{
		Description: "The vcs connector data source.",
		ReadContext: dataSourceVCSConnectorRead,
		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The vcs connector unique resource id.",
			},
			"project": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: internal.ResourceNameValidation(regexp.MustCompile(fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern))),
				Description:      "The project name in projects/{resource id} format.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The vcs connector full name in projects/{project}/vcsConnector/{resource id} format.",
			},
			"title": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The vcs connector title.",
			},
			"creator": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The vcs connector creator in users/{email} format.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The vcs connector create time in YYYY-MM-DDThh:mm:ss.000Z format",
			},
			"updater": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The vcs connector updater in users/{email} format.",
			},
			"update_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The vcs connector update time in YYYY-MM-DDThh:mm:ss.000Z format",
			},
			"vcs_provider": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The vcs provider full name in vcsProviders/{resource id} format.",
			},
			"database_group": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "Apply changes to the database group.",
			},
			"repository_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The connected repository id in vcs provider.",
			},
			"repository_path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The connected repository path in vcs provider.",
			},
			"repository_directory": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The connected repository directory in vcs provider.",
			},
			"repository_branch": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The connected repository branch in vcs provider.",
			},
			"repository_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The connected repository url in vcs provider.",
			},
		},
	}
}

func dataSourceVCSConnectorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	project := d.Get("project").(string)
	connectorName := fmt.Sprintf("%s/%s%s", project, internal.VCSConnectorNamePrefix, d.Get("resource_id").(string))

	connector, err := c.GetVCSConnector(ctx, connectorName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(connector.Name)

	return setVCSConnector(d, connector)
}

func setVCSConnector(d *schema.ResourceData, connector *v1pb.VCSConnector) diag.Diagnostics {
	projectID, connectorID, err := internal.GetVCSConnectorID(connector.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("resource_id", connectorID); err != nil {
		return diag.Errorf("cannot set resource_id for vcs connector: %s", err.Error())
	}
	if err := d.Set("project", fmt.Sprintf("%s%s", internal.ProjectNamePrefix, projectID)); err != nil {
		return diag.Errorf("cannot set project for vcs connector: %s", err.Error())
	}
	if err := d.Set("title", connector.Title); err != nil {
		return diag.Errorf("cannot set title for vcs connector: %s", err.Error())
	}
	if err := d.Set("name", connector.Name); err != nil {
		return diag.Errorf("cannot set name for vcs connector: %s", err.Error())
	}
	if err := d.Set("creator", connector.Creator); err != nil {
		return diag.Errorf("cannot set creator for vcs connector: %s", err.Error())
	}
	if err := d.Set("create_time", connector.CreateTime.AsTime().UTC().Format(time.RFC3339)); err != nil {
		return diag.Errorf("cannot set create_time for vcs connector: %s", err.Error())
	}
	if err := d.Set("updater", connector.Updater); err != nil {
		return diag.Errorf("cannot set updater for vcs connector: %s", err.Error())
	}
	if err := d.Set("update_time", connector.UpdateTime.AsTime().UTC().Format(time.RFC3339)); err != nil {
		return diag.Errorf("cannot set update_time for vcs connector: %s", err.Error())
	}
	if err := d.Set("vcs_provider", connector.VcsProvider); err != nil {
		return diag.Errorf("cannot set vcs_provider for vcs connector: %s", err.Error())
	}
	if err := d.Set("database_group", connector.DatabaseGroup); err != nil {
		return diag.Errorf("cannot set database_group for vcs connector: %s", err.Error())
	}
	if err := d.Set("repository_id", connector.ExternalId); err != nil {
		return diag.Errorf("cannot set repository_id for vcs connector: %s", err.Error())
	}
	if err := d.Set("repository_path", connector.FullPath); err != nil {
		return diag.Errorf("cannot set repository_path for vcs connector: %s", err.Error())
	}
	if err := d.Set("repository_directory", connector.BaseDirectory); err != nil {
		return diag.Errorf("cannot set repository_directory for vcs connector: %s", err.Error())
	}
	if err := d.Set("repository_branch", connector.Branch); err != nil {
		return diag.Errorf("cannot set repository_branch for vcs connector: %s", err.Error())
	}
	if err := d.Set("repository_url", connector.WebUrl); err != nil {
		return diag.Errorf("cannot set repository_url for vcs connector: %s", err.Error())
	}

	return nil
}
