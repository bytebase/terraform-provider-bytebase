package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

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
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The environment unique name.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"order": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Default:     nil,
				Description: "The environment sorting order.",
			},
		},
	}
}

func resourceEnvironmentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name, ok := d.Get("name").(string)
	if !ok {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get the environment name",
			Detail:   "The environment name is required for creation",
		})
		return diags
	}

	create := &api.EnvironmentCreate{
		Name: name,
	}

	order, err := getEnvironmentOrder(d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse the environment order",
			Detail:   err.Error(),
		})
		return diags
	}
	create.Order = order

	env, err := c.CreateEnvironment(create)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(env.ID))

	return resourceEnvironmentRead(ctx, d, m)
}

func resourceEnvironmentRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	envID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	env, err := c.GetEnvironment(envID)
	if err != nil {
		return diag.FromErr(err)
	}

	return setEnvironment(d, env)
}

func resourceEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	var diags diag.Diagnostics

	envID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("name") || d.HasChange("order") {
		patch := &api.EnvironmentPatch{}

		name, ok := d.GetOk("name")
		if ok {
			val := name.(string)
			patch.Name = &val
		}

		order, err := getEnvironmentOrder(d)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to parse the environment order",
				Detail:   err.Error(),
			})
			return diags
		}
		patch.Order = order

		if _, err := c.UpdateEnvironment(envID, patch); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceEnvironmentRead(ctx, d, m)
}

func resourceEnvironmentDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	envID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := c.DeleteEnvironment(envID); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func setEnvironment(d *schema.ResourceData, env *api.Environment) diag.Diagnostics {
	if err := d.Set("name", env.Name); err != nil {
		return diag.Errorf("cannot set name for environment: %s", err.Error())
	}
	if err := d.Set("order", env.Order); err != nil {
		return diag.Errorf("cannot set order for environment: %s", err.Error())
	}

	return nil
}

func getEnvironmentOrder(d *schema.ResourceData) (*int, error) {
	order, ok := d.Get("order").(string)
	if ok && order != "" {
		val, err := strconv.Atoi(order)
		if err != nil {
			return nil, err
		}
		return &val, nil
	}

	return nil, nil
}
