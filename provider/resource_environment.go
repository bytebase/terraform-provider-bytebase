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
				Description:  "The environment title.",
				ValidateFunc: validation.StringMatch(environmentTitleRegex, fmt.Sprintf("environment title must matches %v", environmentTitleRegex)),
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The environment full name in environments/{resource id} format.",
			},
			"order": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "The environment sorting order.",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"environment_tier_policy": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					v1pb.EnvironmentTier_PROTECTED.String(),
					v1pb.EnvironmentTier_UNPROTECTED.String(),
				}, false),
				Description: "If marked as PROTECTED, developers cannot execute any query on this environment's databases using SQL Editor by default.",
			},
		},
	}
}

func resourceEnvironmentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	environmentID := d.Get("resource_id").(string)
	environmentName := fmt.Sprintf("%s%s", internal.EnvironmentNamePrefix, environmentID)

	existedEnv, err := c.GetEnvironment(ctx, environmentName)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get environment %s failed with error: %v", environmentName, err))
	}

	title := d.Get("title").(string)
	order := d.Get("order").(int)
	tier := v1pb.EnvironmentTier(v1pb.EnvironmentTier_value[d.Get("environment_tier_policy").(string)])

	var diags diag.Diagnostics
	if existedEnv != nil && err == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Environment already exists",
			Detail:   fmt.Sprintf("Environment %s already exists, try to exec the update operation", environmentID),
		})

		if existedEnv.State == v1pb.State_DELETED {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Environment is deleted",
				Detail:   fmt.Sprintf("Environment %s already deleted, try to undelete the environment", environmentID),
			})
			if _, err := c.UndeleteEnvironment(ctx, environmentName); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to undelete environment",
					Detail:   fmt.Sprintf("Undelete environment %s failed, error: %v", environmentName, err),
				})
				return diags
			}
		}

		updateMasks := []string{}
		if title != "" && title != existedEnv.Title {
			updateMasks = append(updateMasks, "title")
		}
		if order != int(existedEnv.Order) {
			updateMasks = append(updateMasks, "order")
		}
		if tier != existedEnv.Tier {
			updateMasks = append(updateMasks, "tier")
		}

		if len(updateMasks) > 0 {
			if _, err := c.UpdateEnvironment(ctx, &v1pb.Environment{
				Name:  environmentName,
				Title: title,
				Order: int32(order),
				Tier:  tier,
				State: v1pb.State_ACTIVE,
			}, updateMasks); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to update environment",
					Detail:   fmt.Sprintf("Update environment %s failed, error: %v", environmentName, err),
				})
				return diags
			}
		}
	} else {
		if _, err := c.CreateEnvironment(ctx, environmentID, &v1pb.Environment{
			Name:  environmentName,
			Title: title,
			Order: int32(order),
			Tier:  tier,
			State: v1pb.State_ACTIVE,
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(environmentName)

	diag := resourceEnvironmentRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	environmentName := d.Id()

	env, err := c.GetEnvironment(ctx, environmentName)
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
	environmentName := d.Id()

	existedEnv, err := c.GetEnvironment(ctx, environmentName)
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	if existedEnv.State == v1pb.State_DELETED {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Environment is deleted",
			Detail:   fmt.Sprintf("Environment %s already deleted, try to undelete the environment", environmentName),
		})
		if _, err := c.UndeleteEnvironment(ctx, environmentName); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to undelete environment",
				Detail:   fmt.Sprintf("Undelete environment %s failed, error: %v", environmentName, err),
			})
			return diags
		}
	}

	paths := []string{}
	if d.HasChange("title") {
		paths = append(paths, "title")
	}

	if d.HasChange("order") {
		paths = append(paths, "order")
	}

	if d.HasChange("environment_tier_policy") {
		paths = append(paths, "tier")
	}

	if len(paths) > 0 {
		title := d.Get("title").(string)
		order := d.Get("order").(int)
		tier := v1pb.EnvironmentTier(v1pb.EnvironmentTier_value[d.Get("environment_tier_policy").(string)])

		if _, err := c.UpdateEnvironment(ctx, &v1pb.Environment{
			Name:  environmentName,
			Title: title,
			Order: int32(order),
			Tier:  tier,
			State: v1pb.State_ACTIVE,
		}, paths); err != nil {
			return diag.FromErr(err)
		}
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
	environmentName := d.Id()

	if err := c.DeleteEnvironment(ctx, environmentName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func setEnvironment(d *schema.ResourceData, env *v1pb.Environment) diag.Diagnostics {
	environmentID, err := internal.GetEnvironmentID(env.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("resource_id", environmentID); err != nil {
		return diag.Errorf("cannot set resource_id for environment: %s", err.Error())
	}
	if err := d.Set("title", env.Title); err != nil {
		return diag.Errorf("cannot set title for environment: %s", err.Error())
	}
	if err := d.Set("name", env.Name); err != nil {
		return diag.Errorf("cannot set name for environment: %s", err.Error())
	}
	if err := d.Set("order", env.Order); err != nil {
		return diag.Errorf("cannot set order for environment: %s", err.Error())
	}
	if err := d.Set("environment_tier_policy", env.Tier.String()); err != nil {
		return diag.Errorf("cannot set environment_tier_policy for environment: %s", err.Error())
	}

	return nil
}
