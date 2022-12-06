package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func dataSourceEnvironment() *schema.Resource {
	return &schema.Resource{
		Description: "The environment data source.",
		ReadContext: dataSourceEnvironmentRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The environment id.",
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The environment unique name.",
			},
			"order": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The environment sorting order.",
			},
		},
	}
}

func dataSourceEnvironmentRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	name := d.Get("name").(string)

	var diags diag.Diagnostics
	environmentList, err := c.ListEnvironment(&api.EnvironmentFind{
		Name: name,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if len(environmentList) == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get the environment",
			Detail:   fmt.Sprintf("Cannot find the environment %s", name),
		})
		return diags
	}

	env := environmentList[0]
	d.SetId(strconv.Itoa(env.ID))

	return setEnvironment(d, env)
}
