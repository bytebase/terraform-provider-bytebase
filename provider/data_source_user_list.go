package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func dataSourceUserList() *schema.Resource {
	return &schema.Resource{
		Description: "The user data source list.",
		ReadContext: dataSourceUserListRead,
		Schema: map[string]*schema.Schema{
			"show_deleted": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Including removed users in the response.",
			},
			"users": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user name in users/{user id or email} format.",
						},
						"title": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user title.",
						},
						"email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user email.",
						},
						"phone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user phone.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user type.",
						},
						"roles": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "The user's roles in the workspace level",
						},
						"mfa_enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "The mfa_enabled flag means if the user has enabled MFA.",
						},
						"state": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user is deleted or not.",
						},
						"last_login_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user last login time.",
						},
						"last_change_password_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user last change password time.",
						},
						"source": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Source means where the user comes from. For now we support Entra ID SCIM sync, so the source could be Entra ID.",
						},
					},
				},
			},
		},
	}
}

func dataSourceUserListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	response, err := c.ListUser(ctx, d.Get("show_deleted").(bool))
	if err != nil {
		return diag.FromErr(err)
	}

	workspaceIAM, err := c.GetWorkspaceIAMPolicy(ctx)
	if err != nil {
		return diag.Errorf("cannot get workspace IAM with error: %s", err.Error())
	}

	users := make([]map[string]interface{}, 0)
	for _, user := range response.Users {
		raw := make(map[string]interface{})
		raw["name"] = user.Name
		raw["email"] = user.Email
		raw["title"] = user.Title
		raw["phone"] = user.Phone
		raw["type"] = user.UserType.String()
		raw["mfa_enabled"] = user.MfaEnabled
		raw["state"] = user.State.String()
		if p := user.Profile; p != nil {
			raw["source"] = p.Source
			raw["last_login_time"] = p.LastLoginTime.AsTime().UTC().Format(time.RFC3339)
			raw["last_change_password_time"] = p.LastChangePasswordTime.AsTime().UTC().Format(time.RFC3339)
		}
		raw["roles"] = getUserRoles(workspaceIAM, user.Email)
		users = append(users, raw)
	}
	if err := d.Set("users", users); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}
