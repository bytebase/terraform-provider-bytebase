package provider

import (
	"context"
	"fmt"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Description:   "The user resource.",
		ReadContext:   resourceUserRead,
		DeleteContext: resourceUserDelete,
		CreateContext: resourceUserCreate,
		UpdateContext: resourceUserUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"title": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The user title.",
			},
			"email": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The user email.",
			},
			"phone": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The user phone.",
			},
			"password": {
				Type:        schema.TypeString,
				Sensitive:   true,
				Optional:    true,
				Description: "The user login password.",
				DiffSuppressFunc: func(_, oldValue, newValue string, d *schema.ResourceData) bool {
					// During creation, never suppress
					if d.Id() == "" {
						return false
					}

					// Get the raw config value to see if password is set
					rawConfig := d.GetRawConfig()
					sensitiveConfig := rawConfig.GetAttr("password")

					// If password is not in config, suppress the diff
					if sensitiveConfig.IsNull() {
						return true
					}

					// If password is in config, we need to hash it and compare
					// The 'new' value here is already hashed by StateFunc
					// The 'old' value is the hash from state

					// If both are hashes and they're equal, suppress (no change)
					if oldValue != "" && newValue != "" && oldValue == newValue {
						return true
					}

					// Otherwise allow the diff (password changed)
					return false
				},
			},
			"service_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The service key for service account.",
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The user type, should be USER or SERVICE_ACCOUNT. Cannot change after creation.",
				Default:     v1pb.UserType_USER.String(),
				ValidateFunc: validation.StringInSlice([]string{
					v1pb.UserType_SERVICE_ACCOUNT.String(),
					v1pb.UserType_USER.String(),
				}, false),
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user name in users/{email} format.",
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

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	fullName := d.Id()
	user, err := c.GetUser(ctx, fullName)
	if err != nil {
		// Check if the resource was deleted outside of Terraform
		if internal.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Resource %s not found, removing from state", fullName))
			// Remove from state to trigger recreation on next apply
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return setUser(d, user)
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	email := d.Get("email").(string)
	userName := fmt.Sprintf("%s%s", internal.UserNamePrefix, email)

	existedUser, err := c.GetUser(ctx, userName)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get user %s failed with error: %v", userName, err))
	}

	user := &v1pb.User{
		Name:     userName,
		Title:    d.Get("title").(string),
		Password: d.Get("password").(string),
		Phone:    d.Get("phone").(string),
		Email:    email,
		UserType: v1pb.UserType(v1pb.UserType_value[d.Get("type").(string)]),
		State:    v1pb.State_ACTIVE,
	}

	var diags diag.Diagnostics
	if existedUser != nil && err == nil {
		user.Name = existedUser.Name
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "User already exists",
			Detail:   fmt.Sprintf("User %s already exists, try to exec the update operation", userName),
		})

		if user.UserType != existedUser.UserType {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Cannot change the user type",
				Detail:   fmt.Sprintf("User %s should be %v type", userName, existedUser.UserType.String()),
			})
			return diags
		}

		if existedUser.State == v1pb.State_DELETED {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "User is deleted",
				Detail:   fmt.Sprintf("User %s already deleted, try to undelete the user", userName),
			})
			if _, err := c.UndeleteUser(ctx, existedUser.Name); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to undelete user",
					Detail:   fmt.Sprintf("Undelete user %s failed, error: %v", existedUser.Name, err),
				})
				return diags
			}
		}

		updateMasks := []string{}
		if email != "" && email != existedUser.Email {
			updateMasks = append(updateMasks, "email")
		}
		if user.Title != existedUser.Title {
			updateMasks = append(updateMasks, "title")
		}
		if user.Password != "" {
			updateMasks = append(updateMasks, "password")
		}
		rawConfig := d.GetRawConfig()
		if config := rawConfig.GetAttr("phone"); !config.IsNull() && user.Phone != existedUser.Phone {
			updateMasks = append(updateMasks, "phone")
		}
		if len(updateMasks) > 0 {
			if _, err := c.UpdateUser(ctx, user, updateMasks); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to update user",
					Detail:   fmt.Sprintf("Update vcs user %s failed, error: %v", userName, err),
				})
				return diags
			}
		}
		d.SetId(existedUser.Name)
	} else {
		user, err := c.CreateUser(ctx, user)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(user.Name)
	}

	diag := resourceUserRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	userName := d.Id()

	if d.HasChange("type") {
		return diag.Errorf("cannot change the user type")
	}

	title := d.Get("title").(string)
	phone := d.Get("phone").(string)
	email := d.Get("email").(string)
	password := d.Get("password").(string)

	existedUser, err := c.GetUser(ctx, userName)
	if err != nil {
		return diag.Errorf("get user %s failed with error: %v", userName, err)
	}

	var diags diag.Diagnostics
	if existedUser.State == v1pb.State_DELETED {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "User is deleted",
			Detail:   fmt.Sprintf("User %s already deleted, try to undelete the user", userName),
		})
		if _, err := c.UndeleteUser(ctx, userName); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to undelete user",
				Detail:   fmt.Sprintf("Undelete user %s failed, error: %v", userName, err),
			})
			return diags
		}
	}

	paths := []string{}
	if d.HasChange("title") && title != "" {
		paths = append(paths, "title")
	}
	if d.HasChange("email") && email != "" {
		paths = append(paths, "email")
	}
	if d.HasChange("phone") {
		paths = append(paths, "phone")
	}
	if d.HasChange("password") && password != "" {
		paths = append(paths, "password")
	}

	if len(paths) > 0 {
		if _, err := c.UpdateUser(ctx, &v1pb.User{
			Name:     existedUser.Name,
			Title:    title,
			Email:    email,
			Phone:    phone,
			Password: password,
			UserType: existedUser.UserType,
			State:    v1pb.State_ACTIVE,
		}, paths); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update user",
				Detail:   fmt.Sprintf("Update user %s failed, error: %v", userName, err),
			})
			return diags
		}
	}

	diag := resourceUserRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	return internal.ResourceDelete(ctx, d, c.DeleteUser)
}
