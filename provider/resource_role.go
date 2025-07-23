package provider

import (
	"context"
	"fmt"
	"strings"

	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func resourceRole() *schema.Resource {
	return &schema.Resource{
		Description:   "The role resource. Require ENTERPRISE subscription. Check the docs https://www.bytebase.com/docs/administration/custom-roles/?source=terraform for more information.",
		ReadContext:   internal.ResourceRead(resourceRoleRead),
		DeleteContext: internal.ResourceDelete,
		CreateContext: resourceRoleCreate,
		UpdateContext: resourceRoleUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The role unique resource id.",
			},
			"title": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The role title.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The role description.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The role full name in roles/{resource id} format.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The role type.",
			},
			"permissions": {
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Description: "The role permissions. Permissions should start with \"bb.\" prefix. Check https://github.com/bytebase/bytebase/blob/main/backend/component/iam/permission.yaml for all permissions.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	fullName := d.Id()
	role, err := c.GetRole(ctx, fullName)
	if err != nil {
		return diag.FromErr(err)
	}

	return setRole(d, role)
}

func setRole(d *schema.ResourceData, role *v1pb.Role) diag.Diagnostics {
	roleID, err := internal.GetRoleID(role.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("resource_id", roleID); err != nil {
		return diag.Errorf("cannot set resource_id for role: %s", err.Error())
	}
	if err := d.Set("title", role.Title); err != nil {
		return diag.Errorf("cannot set title for role: %s", err.Error())
	}
	if err := d.Set("name", role.Name); err != nil {
		return diag.Errorf("cannot set name for role: %s", err.Error())
	}
	if err := d.Set("description", role.Description); err != nil {
		return diag.Errorf("cannot set description for role: %s", err.Error())
	}
	if err := d.Set("type", role.Type.String()); err != nil {
		return diag.Errorf("cannot set type for role: %s", err.Error())
	}
	if err := d.Set("permissions", role.Permissions); err != nil {
		return diag.Errorf("cannot set permissions for role: %s", err.Error())
	}

	return nil
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	roleID := d.Get("resource_id").(string)
	roleName := fmt.Sprintf("%s%s", internal.RoleNamePrefix, roleID)

	existedRole, err := c.GetRole(ctx, roleName)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get role %s failed with error: %v", roleName, err))
	}

	title := d.Get("title").(string)
	description := d.Get("description").(string)

	permissions, diagnostic := getRolePermissions(d)
	if diagnostic != nil {
		return diagnostic
	}

	role := &v1pb.Role{
		Name:        roleName,
		Title:       title,
		Description: description,
		Permissions: permissions,
	}

	var diags diag.Diagnostics
	if existedRole != nil && err == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Role already exists",
			Detail:   fmt.Sprintf("Role %s already exists, try to exec the update operation", roleName),
		})

		updateMasks := []string{"title", "permissions"}
		rawConfig := d.GetRawConfig()
		if config := rawConfig.GetAttr("description"); !config.IsNull() && role.Description != existedRole.Description {
			updateMasks = append(updateMasks, "description")
		}

		if _, err := c.UpdateRole(ctx, role, updateMasks); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update role",
				Detail:   fmt.Sprintf("Update role %s failed, error: %v", roleName, err),
			})
			return diags
		}
	} else {
		if _, err := c.CreateRole(ctx, roleID, role); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(roleName)

	diag := resourceRoleRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func getRolePermissions(d *schema.ResourceData) ([]string, diag.Diagnostics) {
	permissions := []string{}
	rawSet, ok := d.Get("permissions").(*schema.Set)
	if !ok {
		return nil, diag.Errorf("invalid role permissions")
	}
	for _, raw := range rawSet.List() {
		permission := raw.(string)
		if !strings.HasPrefix(permission, "bb.") {
			return nil, diag.Errorf("permission should start with \"bb.\" prefix.")
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("resource_id") {
		return diag.Errorf("cannot change the resource id")
	}

	c := m.(api.Client)
	roleName := d.Id()

	permissions, diagnostic := getRolePermissions(d)
	if diagnostic != nil {
		return diagnostic
	}

	existedRole, err := c.GetRole(ctx, roleName)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get role %s failed with error: %v", roleName, err))
	}
	if existedRole == nil {
		// Allow missing.
		existedRole = &v1pb.Role{
			Name:        roleName,
			Title:       d.Get("title").(string),
			Description: d.Get("description").(string),
			Permissions: permissions,
		}
	}

	updateMasks := []string{}
	if d.HasChange("title") {
		updateMasks = append(updateMasks, "title")
		existedRole.Title = d.Get("title").(string)
	}
	if d.HasChange("description") {
		updateMasks = append(updateMasks, "description")
		existedRole.Description = d.Get("description").(string)
	}
	if d.HasChange("permissions") {
		updateMasks = append(updateMasks, "permissions")
		existedRole.Permissions = permissions
	}

	var diags diag.Diagnostics
	if len(updateMasks) > 0 {
		if _, err := c.UpdateRole(ctx, existedRole, updateMasks); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update role",
				Detail:   fmt.Sprintf("Update role %s failed, error: %v", roleName, err),
			})
			return diags
		}
	}

	diag := resourceRoleRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}
