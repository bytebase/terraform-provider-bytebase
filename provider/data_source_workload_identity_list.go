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

func dataSourceWorkloadIdentityList() *schema.Resource {
	return &schema.Resource{
		Description: "The workload identity data source list.",
		ReadContext: dataSourceWorkloadIdentityListRead,
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
				Description: "Show deleted workload identities if specified.",
			},
			"workload_identities": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The workload identity name in workloadIdentities/{email} format.",
						},
						"title": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The display title of the workload identity.",
						},
						"email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The workload identity email.",
						},
						"state": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The workload identity state.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The timestamp when the workload identity was created.",
						},
						"workload_identity_config": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The workload identity configuration for OIDC token validation.",
							Elem: &schema.Resource{
								Schema: getWorkloadIdentityConfigSchema(),
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceWorkloadIdentityListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	parent := internal.ResolveWorkspaceParent(d.Get("parent").(string), c.GetWorkspaceName())
	showDeleted := d.Get("show_deleted").(bool)

	allWorkloadIdentities, err := c.ListWorkloadIdentity(ctx, parent, showDeleted)
	if err != nil {
		return diag.FromErr(err)
	}

	workloadIdentities := make([]map[string]interface{}, 0)
	for _, wi := range allWorkloadIdentities {
		raw := make(map[string]interface{})
		raw["name"] = wi.Name
		raw["email"] = wi.Email
		raw["title"] = wi.Title
		raw["state"] = wi.State.String()
		if wi.CreateTime != nil {
			raw["create_time"] = wi.CreateTime.AsTime().UTC().Format(time.RFC3339)
		}
		raw["workload_identity_config"] = flattenWorkloadIdentityConfig(wi.WorkloadIdentityConfig)
		workloadIdentities = append(workloadIdentities, raw)
	}
	if err := d.Set("workload_identities", workloadIdentities); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}
