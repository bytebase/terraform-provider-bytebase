package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func resourceVCSConnector() *schema.Resource {
	return &schema.Resource{
		Description:   "The vcs connector resource.",
		CreateContext: resourceVCSConnectorCreate,
		ReadContext:   resourceVCSConnectorRead,
		UpdateContext: resourceVCSConnectorUpdate,
		DeleteContext: resourceVCSConnectorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The vcs connector title.",
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: internal.ResourceNameValidation(regexp.MustCompile(fmt.Sprintf("^%s%s$", internal.VCSProviderNamePrefix, internal.ResourceIDPattern))),
				Description:      "The vcs provider full name in vcsProviders/{resource id} format.",
			},
			"database_group": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
				Description: "Apply changes to the database group.",
			},
			"repository_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The connected repository id in vcs provider.",
			},
			"repository_path": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The connected repository path in vcs provider.",
			},
			"repository_directory": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The connected repository directory in vcs provider.",
			},
			"repository_branch": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The connected repository branch in vcs provider.",
			},
			"repository_url": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The connected repository url in vcs provider.",
			},
		},
	}
}

func resourceVCSConnectorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	fullName := d.Id()
	connector, err := c.GetVCSConnector(ctx, fullName)
	if err != nil {
		return diag.FromErr(err)
	}

	return setVCSConnector(d, connector)
}

func resourceVCSConnectorDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	fullName := d.Id()

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if err := c.DeleteVCSConnector(ctx, fullName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func resourceVCSConnectorCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	connectorID := d.Get("resource_id").(string)
	projectName := d.Get("project").(string)
	connectorName := fmt.Sprintf("%s/%s%s", projectName, internal.VCSConnectorNamePrefix, connectorID)

	existedConnector, err := c.GetVCSConnector(ctx, connectorName)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get vcs connector %s failed with error: %v", connectorName, err))
	}

	title := d.Get("title").(string)
	vcsProviderName := d.Get("vcs_provider").(string)
	databaseGroup := d.Get("database_group").(string)
	repositoryID := d.Get("repository_id").(string)
	repositoryPath := d.Get("repository_path").(string)
	repositoryDirectory := d.Get("repository_directory").(string)
	repositoryBranch := d.Get("repository_branch").(string)
	repositoryURL := d.Get("repository_url").(string)

	var diags diag.Diagnostics
	if existedConnector != nil && err == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "VCS connector already exists",
			Detail:   fmt.Sprintf("VCS connector %s already exists, try to exec the update operation", connectorName),
		})

		updateMasks := []string{}
		if repositoryBranch != "" && repositoryBranch != existedConnector.Branch {
			updateMasks = append(updateMasks, "branch")
		}
		if repositoryDirectory != "" && repositoryDirectory != existedConnector.BaseDirectory {
			updateMasks = append(updateMasks, "base_directory")
		}
		if databaseGroup != "" && databaseGroup != existedConnector.DatabaseGroup {
			updateMasks = append(updateMasks, "database_group")
		}

		if len(updateMasks) > 0 {
			if _, err := c.UpdateVCSConnector(ctx, &v1pb.VCSConnector{
				Name:          connectorName,
				Title:         title,
				VcsProvider:   existedConnector.VcsProvider,
				DatabaseGroup: databaseGroup,
				ExternalId:    existedConnector.ExternalId,
				FullPath:      existedConnector.FullPath,
				BaseDirectory: repositoryDirectory,
				Branch:        repositoryBranch,
				WebUrl:        existedConnector.WebUrl,
			}, updateMasks); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to update vcs connector",
					Detail:   fmt.Sprintf("Update vcs connector %s failed, error: %v", connectorName, err),
				})
				return diags
			}
		}
	} else {
		if _, err := c.CreateVCSConnector(ctx, projectName, connectorID, &v1pb.VCSConnector{
			Name:          connectorName,
			Title:         title,
			VcsProvider:   vcsProviderName,
			DatabaseGroup: databaseGroup,
			ExternalId:    repositoryID,
			FullPath:      repositoryPath,
			BaseDirectory: repositoryDirectory,
			Branch:        repositoryBranch,
			WebUrl:        repositoryURL,
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(connectorName)

	diag := resourceVCSConnectorRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceVCSConnectorUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("resource_id") {
		return diag.Errorf("cannot change the resource id")
	}
	if d.HasChange("project") {
		return diag.Errorf("cannot change the project")
	}
	if d.HasChange("repository_id") {
		return diag.Errorf("cannot change the repository_id")
	}
	if d.HasChange("repository_path") {
		return diag.Errorf("cannot change the repository_path")
	}
	if d.HasChange("repository_url") {
		return diag.Errorf("cannot change the repository_url")
	}

	c := m.(api.Client)
	connectorName := d.Id()

	existedConnector, err := c.GetVCSConnector(ctx, connectorName)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get vcs connector %s failed with error: %v", connectorName, err))
		return diag.FromErr(err)
	}

	paths := []string{}
	if d.HasChange("database_group") {
		paths = append(paths, "database_group")
	}
	if d.HasChange("repository_directory") {
		paths = append(paths, "base_directory")
	}
	if d.HasChange("repository_branch") {
		paths = append(paths, "branch")
	}

	var diags diag.Diagnostics
	if len(paths) > 0 {
		if _, err := c.UpdateVCSConnector(ctx, &v1pb.VCSConnector{
			Name:          connectorName,
			Title:         existedConnector.Title,
			VcsProvider:   existedConnector.VcsProvider,
			DatabaseGroup: d.Get("database_group").(string),
			ExternalId:    existedConnector.ExternalId,
			FullPath:      existedConnector.FullPath,
			BaseDirectory: d.Get("repository_directory").(string),
			Branch:        d.Get("repository_branch").(string),
			WebUrl:        existedConnector.WebUrl,
		}, paths); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update vcs connector",
				Detail:   fmt.Sprintf("Update vcs connector %s failed, error: %v", connectorName, err),
			})
			return diags
		}
	}

	diag := resourceVCSConnectorRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}
