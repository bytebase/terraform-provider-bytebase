package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceRoleList() *schema.Resource {
	return &schema.Resource{
		Description: "The role data source list.",
		ReadContext: dataSourceRoleListRead,
		Schema: map[string]*schema.Schema{
			"roles": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The role unique resource id.",
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
				},
			},
		},
	}
}

func dataSourceRoleListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	response, err := c.ListRole(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	roles := []map[string]interface{}{}
	for _, role := range response.Roles {
		roleID, err := internal.GetRoleID(role.Name)
		if err != nil {
			return diag.FromErr(err)
		}

		raw := make(map[string]interface{})
		raw["resource_id"] = roleID
		raw["name"] = role.Name
		raw["title"] = role.Title
		raw["description"] = role.Description
		raw["type"] = role.Type.String()
		raw["permissions"] = role.Permissions

		roles = append(roles, raw)
	}

	if err := d.Set("roles", roles); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
