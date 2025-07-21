package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceUserList() *schema.Resource {
	return &schema.Resource{
		Description: "The user data source list.",
		ReadContext: dataSourceUserListRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter users by name with wildcard",
			},
			"email": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter users by email with wildcard",
			},
			"project": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The project full name. Filter users by project.",
				ValidateDiagFunc: internal.ResourceNameValidation(
					fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern),
				),
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  v1pb.State_ACTIVE.String(),
				ValidateFunc: validation.StringInSlice([]string{
					v1pb.State_ACTIVE.String(),
					v1pb.State_DELETED.String(),
				}, false),
				Description: "Filter users by state. Default ACTIVE.",
			},
			"user_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						v1pb.UserType_USER.String(),
						v1pb.UserType_SERVICE_ACCOUNT.String(),
						v1pb.UserType_SYSTEM_BOT.String(),
					}, false),
				},
				Description: "Filter users by types.",
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

	filter := &api.UserFilter{
		Name:    d.Get("name").(string),
		Email:   d.Get("email").(string),
		Project: d.Get("project").(string),
	}
	stateString := d.Get("state").(string)
	stateValue, ok := v1pb.State_value[stateString]
	if ok {
		filter.State = v1pb.State(stateValue)
	}

	userTypes := d.Get("user_types").(*schema.Set)
	for _, userType := range userTypes.List() {
		userTypeString := userType.(string)
		userTypeValue, ok := v1pb.UserType_value[userTypeString]
		if ok {
			filter.UserTypes = append(filter.UserTypes, v1pb.UserType(userTypeValue))
		}
	}

	allUsers, err := c.ListUser(ctx, filter)
	if err != nil {
		return diag.FromErr(err)
	}

	users := make([]map[string]interface{}, 0)
	for _, user := range allUsers {
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
		users = append(users, raw)
	}
	if err := d.Set("users", users); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}
