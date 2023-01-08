package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

var environmentTitleRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func resourceEnvironment() *schema.Resource {
	return &schema.Resource{
		Description:   "The environment resource.",
		CreateContext: resourceEnvironmentCreate,
		ReadContext:   resourceEnvironmentRead,
		UpdateContext: resourceEnvironmentUpdate,
		DeleteContext: resourceEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The environment unique resource id.",
			},
			"title": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The environment unique name.",
				ValidateFunc: validation.StringMatch(environmentTitleRegex, fmt.Sprintf("environment title must matches %v", environmentTitleRegex)),
			},
			"order": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "The environment sorting order.",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"environment_tier_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "UNPROTECTED",
				ValidateFunc: validation.StringInSlice([]string{
					"PROTECTED",
					"UNPROTECTED",
				}, false),
				Description: "If marked as PROTECTED, developers cannot execute any query on this environment's databases using SQL Editor by default.",
			},
		},
	}
}

func resourceEnvironmentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	environmentID := d.Get("resource_id").(string)
	existedEnv, err := c.GetEnvironment(ctx, environmentID)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get environment %s failed with error: %v", environmentID, err))
	}

	var diags diag.Diagnostics
	if existedEnv != nil && err == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Environment already exists",
			Detail:   fmt.Sprintf("Environment %s already exists, try to exec the update operation", environmentID),
		})

		if existedEnv.State == api.Deleted {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Environment is deleted",
				Detail:   fmt.Sprintf("Environment %s already deleted, try to undelete the environment", environmentID),
			})
			if _, err := c.UndeleteEnvironment(ctx, environmentID); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to undelete environment",
					Detail:   fmt.Sprintf("Undelete environment %s failed, error: %v", environmentID, err),
				})
				return diags
			}
		}

		title := d.Get("title").(string)
		order := d.Get("order").(int)
		tier := d.Get("environment_tier_policy").(string)
		env, err := c.UpdateEnvironment(ctx, environmentID, &api.EnvironmentPatchMessage{
			Title: &title,
			Order: &order,
			Tier:  &tier,
		})
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update environment",
				Detail:   fmt.Sprintf("Update environment %s failed, error: %v", environmentID, err),
			})
			return diags
		}

		d.SetId(env.Name)
	} else {
		env, err := c.CreateEnvironment(ctx, environmentID, &api.EnvironmentMessage{
			Title: d.Get("title").(string),
			Order: d.Get("order").(int),
			Tier:  d.Get("environment_tier_policy").(string),
		})
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(env.Name)
	}

	diag := resourceEnvironmentRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	envID, err := internal.GetEnvironmentID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	env, err := c.GetEnvironment(ctx, envID)
	if err != nil {
		return diag.FromErr(err)
	}

	return setEnvironment(d, env)
}

func resourceEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("resource_id") {
		return diag.Errorf("cannot change the resource id")
	}

	c := m.(api.Client)

	envID, err := internal.GetEnvironmentID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	existedEnv, err := c.GetEnvironment(ctx, envID)
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	if existedEnv.State == api.Deleted {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Environment is deleted",
			Detail:   fmt.Sprintf("Environment %s already deleted, try to undelete the environment", envID),
		})
		if _, err := c.UndeleteEnvironment(ctx, envID); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to undelete environment",
				Detail:   fmt.Sprintf("Undelete environment %s failed, error: %v", envID, err),
			})
			return diags
		}
	}

	patch := &api.EnvironmentPatchMessage{}
	if d.HasChange("title") {
		title, ok := d.Get("title").(string)
		if ok {
			patch.Title = &title
		}
	}

	if d.HasChange("order") {
		order, ok := d.Get("order").(int)
		if ok {
			patch.Order = &order
		}
	}

	if d.HasChange("environment_tier_policy") {
		tier, ok := d.Get("environment_tier_policy").(string)
		if ok {
			patch.Tier = &tier
		}
	}

	if _, err := c.UpdateEnvironment(ctx, envID, patch); err != nil {
		return diag.FromErr(err)
	}

	diag := resourceEnvironmentRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceEnvironmentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	envID, err := internal.GetEnvironmentID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := c.DeleteEnvironment(ctx, envID); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func setEnvironment(d *schema.ResourceData, env *api.EnvironmentMessage) diag.Diagnostics {
	if err := d.Set("title", env.Title); err != nil {
		return diag.Errorf("cannot set name for environment: %s", err.Error())
	}
	if err := d.Set("order", env.Order); err != nil {
		return diag.Errorf("cannot set order for environment: %s", err.Error())
	}
	if err := d.Set("environment_tier_policy", env.Tier); err != nil {
		return diag.Errorf("cannot set environment_tier_policy for environment: %s", err.Error())
	}

	return nil
}
