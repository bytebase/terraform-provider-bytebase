package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func dataSourceDatabaseRole() *schema.Resource {
	return &schema.Resource{
		Description: "The database role data source.",
		ReadContext: dataSourceRoleRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The database role id.",
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The role unique name.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"instance": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The instance unique name.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"connection_limit": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Connection count limit for role",
			},
			"valid_until": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "It sets a date and time after which the role's password is no longer valid.",
			},
			"attribute": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The attribute for the role.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"super_user": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Set the `SUPERUSER` attribute for the role. Default `false`",
						},
						"no_inherit": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Set the `NOINHERIT` attribute for the role. Default `false`.",
						},
						"create_role": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Set the `CREATEROLE` attribute for the role. Default `false`.",
						},
						"create_db": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Set the `CREATEDB` attribute for the role. Default `false`.",
						},
						"can_login": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Set the `LOGIN` attribute for the role. Default `false`.",
						},
						"replication": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Set the `REPLICATION` attribute for the role. Default `false`.",
						},
						"bypass_rls": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Set the `BYPASSRLS` attribute for the role. Default `false`.",
						},
					},
				},
			},
		},
	}
}

func dataSourceRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	instanceName := d.Get("instance").(string)
	roleName := d.Get("name").(string)

	ins, diags := findInstanceByName(ctx, c, instanceName)
	if diags != nil {
		return diags
	}

	role, err := c.GetRole(ctx, ins.ID, roleName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(getRoleIdentifier(role))
	return setRole(d, role)
}
