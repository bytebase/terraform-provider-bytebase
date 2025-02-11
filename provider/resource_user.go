package provider

import (
	"context"
	"fmt"
	"slices"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

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
				Description: "The user phone.",
			},
			"password": {
				Type:        schema.TypeString,
				Sensitive:   true,
				Optional:    true,
				Description: "The user login password.",
			},
			"roles": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The user's roles in the workspace level",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user type.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user name in users/{user id or email} format.",
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
		return diag.FromErr(err)
	}

	return setUser(ctx, c, d, user)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	fullName := d.Id()

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if err := c.DeleteUser(ctx, fullName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	email := d.Get("email").(string)
	userName := fmt.Sprintf("%s%s", internal.UserNamePrefix, email)

	existedUser, err := c.GetUser(ctx, userName)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get user %s failed with error: %v", userName, err))
	}

	title := d.Get("title").(string)
	phone := d.Get("phone").(string)
	password := d.Get("password").(string)

	var diags diag.Diagnostics
	if existedUser != nil && err == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "User already exists",
			Detail:   fmt.Sprintf("User %s already exists, try to exec the update operation", userName),
		})

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

		updateMasks := []string{}
		if email != "" && email != existedUser.Email {
			updateMasks = append(updateMasks, "email")
		}
		if title != "" && title != existedUser.Title {
			updateMasks = append(updateMasks, "title")
		}
		if password != "" {
			updateMasks = append(updateMasks, "password")
		}
		if phone != "" && phone != existedUser.Phone {
			updateMasks = append(updateMasks, "phone")
		}
		if len(updateMasks) > 0 {
			if _, err := c.UpdateUser(ctx, &v1pb.User{
				Name:     existedUser.Name,
				Title:    title,
				Password: password,
				Phone:    phone,
				Email:    email,
				UserType: existedUser.UserType,
				State:    v1pb.State_ACTIVE,
			}, updateMasks); err != nil {
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
		user, err := c.CreateUser(ctx, &v1pb.User{
			Name:     userName,
			Title:    title,
			Password: password,
			Phone:    phone,
			Email:    email,
			UserType: v1pb.UserType_USER,
			State:    v1pb.State_ACTIVE,
		})
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(user.Name)
	}

	if err := patchWorkspaceIAMPolicy(ctx, c, email, getRoles(d)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to patch user roles",
			Detail:   fmt.Sprintf("Update roles for user %s failed, error: %v", userName, err),
		})
		return diags
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

	if d.HasChange("roles") {
		if err := patchWorkspaceIAMPolicy(ctx, c, email, getRoles(d)); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to patch user roles",
				Detail:   fmt.Sprintf("Update roles for user %s failed, error: %v", userName, err),
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

func getRoles(d *schema.ResourceData) []string {
	rawRoles := d.Get("roles").(*schema.Set)
	roleList := []string{}

	for _, rawRole := range rawRoles.List() {
		roleList = append(roleList, rawRole.(string))
	}
	return roleList
}

func patchWorkspaceIAMPolicy(ctx context.Context, client api.Client, email string, roles []string) error {
	workspaceIamPolicy, err := client.GetWorkspaceIAMPolicy(ctx)
	if err != nil {
		return errors.Errorf("cannot get workspace IAM with error: %s", err.Error())
	}
	patchMember := fmt.Sprintf("user:%s", email)
	roleMap := map[string]bool{}
	for _, role := range roles {
		roleMap[role] = true
	}

	for _, binding := range workspaceIamPolicy.Bindings {
		index := slices.Index(binding.Members, patchMember)
		if !roleMap[binding.Role] {
			if index >= 0 {
				binding.Members = slices.Delete(binding.Members, index, index+1)
			}
		} else {
			if index < 0 {
				binding.Members = append(binding.Members, patchMember)
			}
		}

		delete(roleMap, binding.Role)
	}

	for role := range roleMap {
		workspaceIamPolicy.Bindings = append(workspaceIamPolicy.Bindings, &v1pb.Binding{
			Role: role,
			Members: []string{
				patchMember,
			},
		})
	}

	if _, err := client.SetWorkspaceIAMPolicy(ctx, &v1pb.SetIamPolicyRequest{
		Policy: workspaceIamPolicy,
		Etag:   workspaceIamPolicy.Etag,
	}); err != nil {
		return err
	}

	return nil
}
