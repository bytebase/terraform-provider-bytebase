package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func dataSourceWorkspace() *schema.Resource {
	return &schema.Resource{
		Description: "The workspace data source.",
		ReadContext: dataSourceWorkspaceRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The workspace full name in workspaces/{id} format.",
			},
			"title": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The workspace title.",
			},
			"logo": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The branding logo as a data URI (e.g. data:image/png;base64,...).",
			},
			"subscription": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The current subscription of the workspace.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"plan": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The current plan. One of FREE, TEAM, ENTERPRISE.",
						},
						"seats": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of licensed seats.",
						},
						"instances": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of licensed instances.",
						},
						"expires_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The expiration time of the subscription.",
						},
						"trialing": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the subscription is in trial.",
						},
					},
				},
			},
		},
	}
}

func dataSourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	workspace, err := c.GetWorkspace(ctx, c.GetWorkspaceName())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(workspace.Name)

	if diags := setWorkspace(d, workspace); diags.HasError() {
		return diags
	}

	subscription, err := c.GetSubscription(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return setSubscription(d, subscription)
}
