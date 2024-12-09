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

func dataSourceVCSConnectorList() *schema.Resource {
	return &schema.Resource{
		Description: "The vcs connector data source list.",
		ReadContext: dataSourceVCSConnectorListRead,
		Schema: map[string]*schema.Schema{
			"project": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: internal.ResourceNameValidation(regexp.MustCompile(fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern))),
				Description:      "The project name in projects/{resource id} format.",
			},
			"vcs_connectors": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The vcs connector unique resource id.",
						},
						"project": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The project name in projects/{resource id} format.",
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
				},
			},
		},
	}
}

func dataSourceVCSConnectorListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	project := d.Get("project").(string)

	response, err := c.ListVCSConnector(ctx, project)
	if err != nil {
		return diag.FromErr(err)
	}

	connectors := []map[string]interface{}{}
	for _, connector := range response.VcsConnectors {
		projectID, connectorID, err := internal.GetVCSConnectorID(connector.Name)
		if err != nil {
			return diag.FromErr(err)
		}

		rawConnector := make(map[string]interface{})
		rawConnector["resource_id"] = connectorID
		rawConnector["project"] = fmt.Sprintf("%s%s", internal.ProjectNamePrefix, projectID)
		rawConnector["title"] = connector.Title
		rawConnector["name"] = connector.Name
		rawConnector["creator"] = connector.Creator
		rawConnector["create_time"] = connector.CreateTime.AsTime().UTC().Format(time.RFC3339)
		rawConnector["updater"] = connector.Updater
		rawConnector["update_time"] = connector.UpdateTime.AsTime().UTC().Format(time.RFC3339)
		rawConnector["vcs_provider"] = connector.VcsProvider
		rawConnector["database_group"] = connector.DatabaseGroup
		rawConnector["repository_id"] = connector.ExternalId
		rawConnector["repository_path"] = connector.FullPath
		rawConnector["repository_directory"] = connector.BaseDirectory
		rawConnector["repository_branch"] = connector.Branch
		rawConnector["repository_url"] = connector.WebUrl

		connectors = append(connectors, rawConnector)
	}

	if err := d.Set("vcs_connectors", connectors); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}
