package provider

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		Description: "The user data source.",
		ReadContext: dataSourceUserRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: internal.ResourceNameValidation(
					regexp.MustCompile(fmt.Sprintf("^%s", internal.UserNamePrefix)),
				),
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
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	userName := d.Get("name").(string)

	user, err := c.GetUser(ctx, userName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(user.Name)

	return setUser(d, user)
}

func setUser(d *schema.ResourceData, user *v1pb.User) diag.Diagnostics {
	if err := d.Set("title", user.Title); err != nil {
		return diag.Errorf("cannot set title for user: %s", err.Error())
	}
	if err := d.Set("email", user.Email); err != nil {
		return diag.Errorf("cannot set email for user: %s", err.Error())
	}
	if err := d.Set("phone", user.Phone); err != nil {
		return diag.Errorf("cannot set phone for user: %s", err.Error())
	}
	if err := d.Set("type", user.UserType.String()); err != nil {
		return diag.Errorf("cannot set type for user: %s", err.Error())
	}
	if user.ServiceKey != "" {
		if err := d.Set("service_key", user.ServiceKey); err != nil {
			return diag.Errorf("cannot set the service_key: %s", err.Error())
		}
	}
	if err := d.Set("mfa_enabled", user.MfaEnabled); err != nil {
		return diag.Errorf("cannot set mfa_enabled for user: %s", err.Error())
	}
	if err := d.Set("state", user.State.String()); err != nil {
		return diag.Errorf("cannot set state for user: %s", err.Error())
	}
	if p := user.Profile; p != nil {
		if err := d.Set("last_login_time", p.LastLoginTime.AsTime().UTC().Format(time.RFC3339)); err != nil {
			return diag.Errorf("cannot set last_login_time for user: %s", err.Error())
		}
		if err := d.Set("last_change_password_time", p.LastChangePasswordTime.AsTime().UTC().Format(time.RFC3339)); err != nil {
			return diag.Errorf("cannot set last_change_password_time for user: %s", err.Error())
		}
		if err := d.Set("source", p.Source); err != nil {
			return diag.Errorf("cannot set source for user: %s", err.Error())
		}
	}
	return nil
}
