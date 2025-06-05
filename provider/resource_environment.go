package provider

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func resourceEnvironment() *schema.Resource {
	return &schema.Resource{
		Description:   "The environment resource.",
		CreateContext: resourceEnvironmentUpsert,
		ReadContext:   resourceEnvironmentRead,
		UpdateContext: resourceEnvironmentUpsert,
		DeleteContext: resourceEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The environment unique id.",
			},
			"title": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The environment display name.",
			},
			"order": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "The environment sorting order.",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The environment readonly name in environments/{id} format.",
			},
			"color": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The environment color.",
			},
			"protected": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "The environment is protected or not.",
			},
		},
	}
}

func resourceEnvironmentUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	newOrder := d.Get("order").(int)
	environmentID := d.Get("resource_id").(string)
	environmentName := fmt.Sprintf("%s%s", internal.EnvironmentNamePrefix, environmentID)
	v1Env := &v1pb.EnvironmentSetting_Environment{
		Id:    environmentID,
		Title: d.Get("title").(string),
		Color: d.Get("color").(string),
		Tags: map[string]string{
			"protected": "protected",
		},
	}
	if !d.Get("protected").(bool) {
		v1Env.Tags["protected"] = "unprotected"
	}

	existedEnv, order, enironmentList, err := findEnvironment(ctx, c, environmentName)
	if err != nil {
		if !strings.HasPrefix(err.Error(), "cannot found the environment") {
			return diag.FromErr(err)
		}
	}
	if enironmentList == nil {
		return diag.Errorf("cannot found environment setting")
	}

	if newOrder >= len(enironmentList) {
		return diag.Errorf("the new order %v out of range %v", newOrder, len(enironmentList)-1)
	}

	var diags diag.Diagnostics
	if existedEnv != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Environment already exists",
			Detail:   fmt.Sprintf("Environment %s already exists, try to exec the update operation", environmentID),
		})

		if order == newOrder {
			enironmentList[order] = v1Env
		} else {
			enironmentList = slices.Delete(enironmentList, order, order+1)
			enironmentList = slices.Insert(enironmentList, newOrder, v1Env)
		}
	} else {
		enironmentList = slices.Insert(enironmentList, newOrder, v1Env)
	}

	if err := updateEnvironmentSetting(ctx, c, enironmentList); err != nil {
		return diag.FromErr(err)
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

	env, order, _, err := findEnvironment(ctx, c, environmentName)
	if err != nil {
		return diag.FromErr(err)
	}

	return setEnvironment(d, env, order)
}

func resourceEnvironmentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	environmentName := d.Id()

	_, order, enironmentList, err := findEnvironment(ctx, c, environmentName)
	if err != nil {
		return diag.FromErr(err)
	}

	enironmentList = slices.Delete(enironmentList, order, order+1)
	if err := updateEnvironmentSetting(ctx, c, enironmentList); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func updateEnvironmentSetting(ctx context.Context, client api.Client, list []*v1pb.EnvironmentSetting_Environment) error {
	_, err := client.UpsertSetting(ctx, &v1pb.Setting{
		Name: fmt.Sprintf("%s%s", internal.SettingNamePrefix, v1pb.Setting_ENVIRONMENT.String()),
		Value: &v1pb.Value{
			Value: &v1pb.Value_EnvironmentSetting{
				EnvironmentSetting: &v1pb.EnvironmentSetting{
					Environments: list,
				},
			},
		},
	}, []string{})
	return err
}

func getEnvironmentList(ctx context.Context, client api.Client) ([]*v1pb.EnvironmentSetting_Environment, error) {
	environmentSetting, err := client.GetSetting(ctx, fmt.Sprintf("%s%s", internal.SettingNamePrefix, v1pb.Setting_ENVIRONMENT.String()))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get environment setting")
	}
	return environmentSetting.GetValue().GetEnvironmentSetting().Environments, nil
}

func findEnvironment(ctx context.Context, client api.Client, name string) (*v1pb.EnvironmentSetting_Environment, int, []*v1pb.EnvironmentSetting_Environment, error) {
	enironmentList, err := getEnvironmentList(ctx, client)
	if err != nil {
		return nil, 0, nil, err
	}

	for index, env := range enironmentList {
		if env.Name == name {
			return env, index, enironmentList, nil
		}
	}
	return nil, 0, enironmentList, errors.Errorf("cannot found the environment %v", name)
}

func setEnvironment(d *schema.ResourceData, env *v1pb.EnvironmentSetting_Environment, order int) diag.Diagnostics {
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
	if err := d.Set("order", order); err != nil {
		return diag.Errorf("cannot set order for environment: %s", err.Error())
	}
	if err := d.Set("color", env.Color); err != nil {
		return diag.Errorf("cannot set color for environment: %s", err.Error())
	}
	if err := d.Set("protected", env.Tags["protected"] == "protected"); err != nil {
		return diag.Errorf("cannot set protected for environment: %s", err.Error())
	}

	return nil
}
