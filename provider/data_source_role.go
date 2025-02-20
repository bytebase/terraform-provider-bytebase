package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceRole() *schema.Resource {
	return &schema.Resource{
		Description: "The role data source.",
		ReadContext: dataSourceRoleRead,
		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The role unique resource id.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The role full name in roles/{resource id} format.",
			},
			"title": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The role title.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The role description.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The role type.",
			},
			"permissions": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The role permissions.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	roleName := fmt.Sprintf("%s%s", internal.RoleNamePrefix, d.Get("resource_id").(string))

	role, err := c.GetRole(ctx, roleName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(role.Name)

	return setRole(d, role)
}
