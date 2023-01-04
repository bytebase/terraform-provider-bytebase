package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func resourceDatabaseRole() *schema.Resource {
	return &schema.Resource{
		Description:   "The role resource.",
		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The role unique name.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"environment": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The environment resource id.",
				ValidateFunc: internal.ResourceIDValidation,
			},
			"instance": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The instance resource id.",
				ValidateFunc: internal.ResourceIDValidation,
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The password.",
			},
			"connection_limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      -1,
				ValidateFunc: validation.IntAtLeast(-1),
				Description:  "Connection count limit for role",
			},
			"valid_until": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateDatetime,
				Description:  "It sets a date and time after which the role's password is no longer valid. Should be a timestamp in \"2006-01-02T15:04:05+08:00\" format.",
			},
			"attribute": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "The attribute for the role.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"super_user": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Set the `SUPERUSER` attribute for the role. Default `false`",
						},
						"no_inherit": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Set the `NOINHERIT` attribute for the role. Default `false`.",
						},
						"create_role": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Set the `CREATEROLE` attribute for the role. Default `false`.",
						},
						"create_db": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Set the `CREATEDB` attribute for the role. Default `false`.",
						},
						"can_login": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Set the `LOGIN` attribute for the role. Default `false`.",
						},
						"replication": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Set the `REPLICATION` attribute for the role. Default `false`.",
						},
						"bypass_rls": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Set the `BYPASSRLS` attribute for the role. Default `false`.",
						},
					},
				},
			},
		},
	}
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	environmentID := d.Get("environment").(string)
	instanceID := d.Get("instance").(string)
	roleName := d.Get("name").(string)

	upsert := &api.RoleUpsert{
		Title:     roleName,
		Attribute: convertRoleAttribute(d),
	}

	if v := d.Get("password").(string); v != "" {
		upsert.Password = &v
	}
	if v := d.Get("connection_limit").(int); v != -1 {
		upsert.ConnectionLimit = &v
	}
	if v := d.Get("valid_until").(string); v != "" {
		upsert.ValidUntil = &v
	}

	existedRole, err := c.GetRole(ctx, environmentID, instanceID, roleName)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get role %s failed with error: %v", roleName, err))
	}

	var diags diag.Diagnostics
	if existedRole != nil && err == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Role already exists",
			Detail:   fmt.Sprintf("Role %s already exists, try to exec the update operation", roleName),
		})

		role, err := c.UpdateRole(ctx, environmentID, instanceID, roleName, upsert)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(role.Name)
	} else {
		role, err := c.CreateRole(ctx, environmentID, instanceID, upsert)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(role.Name)
	}

	diag := resourceRoleRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	environmentID, instanceID, roleName, err := internal.GetEnvironmentInstanceRoleID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err := c.GetRole(ctx, environmentID, instanceID, roleName); err != nil {
		return diag.FromErr(err)
	}

	upsert := &api.RoleUpsert{
		Title: roleName,
	}

	if d.HasChange("name") {
		if v := d.Get("name").(string); v != "" {
			upsert.Title = v
		}
	}
	if d.HasChange("password") {
		if v := d.Get("password").(string); v != "" {
			upsert.Password = &v
		}
	}
	if d.HasChange("connection_limit") {
		if v := d.Get("connection_limit").(int); v != -1 {
			upsert.ConnectionLimit = &v
		}
	}
	if d.HasChange("valid_until") {
		if v := d.Get("valid_until").(string); v != "" {
			upsert.ValidUntil = &v
		}
	}
	if d.HasChange("attribute") {
		upsert.Attribute = convertRoleAttribute(d)
	}

	if _, err := c.UpdateRole(ctx, environmentID, instanceID, roleName, upsert); err != nil {
		return diag.FromErr(err)
	}

	role, err := c.GetRole(ctx, environmentID, instanceID, upsert.Title)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(role.Name)
	return setRole(d, role)
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	environmentID, instanceID, roleName, err := internal.GetEnvironmentInstanceRoleID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	role, err := c.GetRole(ctx, environmentID, instanceID, roleName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(role.Name)
	return setRole(d, role)
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	environmentID, instanceID, roleName, err := internal.GetEnvironmentInstanceRoleID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := c.DeleteRole(ctx, environmentID, instanceID, roleName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func setRole(d *schema.ResourceData, role *api.Role) diag.Diagnostics {
	if err := d.Set("name", role.Title); err != nil {
		return diag.Errorf("cannot set name for role: %s", err.Error())
	}
	if err := d.Set("connection_limit", role.ConnectionLimit); err != nil {
		return diag.Errorf("cannot set connection_limit for role: %s", err.Error())
	}
	if err := d.Set("valid_until", role.ValidUntil); err != nil {
		return diag.Errorf("cannot set valid_until for role: %s", err.Error())
	}

	attribute := map[string]interface{}{
		"super_user":  role.Attribute.SuperUser,
		"no_inherit":  role.Attribute.NoInherit,
		"create_role": role.Attribute.CreateRole,
		"create_db":   role.Attribute.CreateDB,
		"can_login":   role.Attribute.CanLogin,
		"replication": role.Attribute.Replication,
		"bypass_rls":  role.Attribute.ByPassRLS,
	}
	if err := d.Set("attribute", []interface{}{attribute}); err != nil {
		return diag.Errorf("cannot set attribute for role: %s", err.Error())
	}

	return nil
}

func convertRoleAttribute(d *schema.ResourceData) *api.RoleAttribute {
	rawList := d.Get("attribute").([]interface{})
	if len(rawList) < 1 {
		return nil
	}

	raw, ok := rawList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &api.RoleAttribute{
		SuperUser:   raw["super_user"].(bool),
		NoInherit:   raw["no_inherit"].(bool),
		CreateRole:  raw["create_role"].(bool),
		CreateDB:    raw["create_db"].(bool),
		CanLogin:    raw["can_login"].(bool),
		Replication: raw["replication"].(bool),
		ByPassRLS:   raw["bypass_rls"].(bool),
	}
}

func validateDatetime(val interface{}, _ string) (ws []string, es []error) {
	raw := val.(string)
	if raw == "" {
		return nil, nil
	}

	if _, err := time.Parse(time.RFC3339, raw); err != nil {
		if err.Error() == "day out of range" {
			return ws, append(es, errors.Errorf("invalid timestamp %s, %s", raw, err.Error()))
		}
		return ws, append(es, errors.Errorf(`valid_until should in "2006-01-02T15:04:05+08:00" format with timezone`))
	}
	return nil, nil
}
