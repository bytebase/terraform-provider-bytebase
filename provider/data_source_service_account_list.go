package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceServiceAccountList() *schema.Resource {
	return &schema.Resource{
		Description: "The service account data source list.",
		ReadContext: dataSourceServiceAccountListRead,
		Schema: map[string]*schema.Schema{
			"parent": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The parent resource. Format: projects/{project} for project-level, workspaces/{workspace id} for workspace-level. Defaults to the workspace if not specified.",
				ValidateDiagFunc: internal.ResourceNameValidation(
					fmt.Sprintf("^%s%s$", internal.WorkspaceNamePrefix, internal.ResourceIDPattern),
					fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern),
				),
			},
			"show_deleted": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Show deleted service accounts if specified.",
			},
			"service_accounts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The service account name in serviceAccounts/{email} format.",
						},
						"title": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The display title of the service account.",
						},
						"email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The service account email.",
						},
						"state": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The service account state.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The timestamp when the service account was created.",
						},
					},
				},
			},
		},
	}
}

func dataSourceServiceAccountListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	parent := internal.ResolveWorkspaceParent(d.Get("parent").(string), c.GetWorkspaceName())
	showDeleted := d.Get("show_deleted").(bool)

	allServiceAccounts, err := c.ListServiceAccount(ctx, parent, showDeleted)
	if err != nil {
		return diag.FromErr(err)
	}

	serviceAccounts := make([]map[string]interface{}, 0)
	for _, sa := range allServiceAccounts {
		raw := make(map[string]interface{})
		raw["name"] = sa.Name
		raw["email"] = sa.Email
		raw["title"] = sa.Title
		raw["state"] = sa.State.String()
		if sa.CreateTime != nil {
			raw["create_time"] = sa.CreateTime.AsTime().UTC().Format(time.RFC3339)
		}
		serviceAccounts = append(serviceAccounts, raw)
	}
	if err := d.Set("service_accounts", serviceAccounts); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}
