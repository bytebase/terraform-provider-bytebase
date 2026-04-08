package provider

import (
	"context"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func resourceWorkspace() *schema.Resource {
	return &schema.Resource{
		Description:   "The workspace resource.",
		CreateContext: resourceWorkspaceUpdate,
		ReadContext:   resourceWorkspaceRead,
		UpdateContext: resourceWorkspaceUpdate,
		DeleteContext: resourceWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The workspace full name in workspaces/{id} format.",
			},
			"title": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The workspace title.",
			},
			"logo": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The branding logo as a data URI (e.g. data:image/png;base64,...).",
			},
			"license": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The license key for the workspace. Upload to activate a subscription plan.",
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

func resourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	workspaceName := d.Id()
	workspace, err := c.GetWorkspace(ctx, workspaceName)
	if err != nil {
		return diag.FromErr(err)
	}

	if diags := setWorkspace(d, workspace); diags.HasError() {
		return diags
	}

	subscription, err := c.GetSubscription(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return setSubscription(d, subscription)
}

func resourceWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	workspaceName := c.GetWorkspaceName()
	if d.Id() != "" {
		workspaceName = d.Id()
	}

	patch := &v1pb.Workspace{
		Name: workspaceName,
	}
	var updateMasks []string

	if d.HasChange("title") {
		patch.Title = d.Get("title").(string)
		updateMasks = append(updateMasks, "title")
	}
	if d.HasChange("logo") {
		patch.Logo = d.Get("logo").(string)
		updateMasks = append(updateMasks, "logo")
	}

	if len(updateMasks) > 0 {
		workspace, err := c.UpdateWorkspace(ctx, patch, updateMasks)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(workspace.Name)
		if diags := setWorkspace(d, workspace); diags.HasError() {
			return diags
		}
	} else {
		d.SetId(workspaceName)
		workspace, err := c.GetWorkspace(ctx, workspaceName)
		if err != nil {
			return diag.FromErr(err)
		}
		if diags := setWorkspace(d, workspace); diags.HasError() {
			return diags
		}
	}

	if d.HasChange("license") {
		license := d.Get("license").(string)
		if license != "" {
			if _, err := c.UploadLicense(ctx, license); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	subscription, err := c.GetSubscription(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return setSubscription(d, subscription)
}

func resourceWorkspaceDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Workspace cannot be deleted, just remove from state.
	d.SetId("")
	return nil
}

func setWorkspace(d *schema.ResourceData, workspace *v1pb.Workspace) diag.Diagnostics {
	if err := d.Set("name", workspace.Name); err != nil {
		return diag.Errorf("cannot set name: %s", err.Error())
	}
	if err := d.Set("title", workspace.Title); err != nil {
		return diag.Errorf("cannot set title: %s", err.Error())
	}
	if err := d.Set("logo", workspace.Logo); err != nil {
		return diag.Errorf("cannot set logo: %s", err.Error())
	}
	return nil
}

func setSubscription(d *schema.ResourceData, subscription *v1pb.Subscription) diag.Diagnostics {
	sub := map[string]interface{}{
		"plan":      subscription.GetPlan().String(),
		"seats":     int(subscription.GetSeats()),
		"instances": int(subscription.GetInstances()),
		"trialing":  subscription.GetTrialing(),
	}
	if v := subscription.GetExpiresTime(); v != nil {
		sub["expires_time"] = v.AsTime().Format("2006-01-02T15:04:05Z")
	} else {
		sub["expires_time"] = ""
	}

	if err := d.Set("subscription", []interface{}{sub}); err != nil {
		return diag.Errorf("cannot set subscription: %s", err.Error())
	}
	return nil
}
