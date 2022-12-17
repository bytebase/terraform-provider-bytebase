package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

const roleIdentifierSeparator = "__"

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
			"instance": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The instance unique name.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The password.",
			},
			"connection_limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     -1,
				Description: "Connection count limit for role",
			},
			"valid_until": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "It sets a date and time after which the role's password is no longer valid.",
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

	instanceName := d.Get("instance").(string)
	ins, diags := findInstanceByName(ctx, c, instanceName)
	if diags != nil {
		return diags
	}

	upsert := &api.RoleUpsert{
		Name:      d.Get("name").(string),
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

	role, err := c.CreateRole(ctx, ins.ID, upsert)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(getRoleIdentifier(role))
	return resourceRoleRead(ctx, d, m)
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	instanceID, name, err := parseRoleIdentifier(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	upsert := &api.RoleUpsert{
		Name: name,
	}

	if d.HasChange("name") {
		if v := d.Get("name").(string); v != "" {
			upsert.Name = v
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

	if _, err := c.UpdateRole(ctx, instanceID, name, upsert); err != nil {
		return diag.FromErr(err)
	}

	role, err := c.GetRole(ctx, instanceID, upsert.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(getRoleIdentifier(role))
	return setRole(d, role)
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	instanceID, name, err := parseRoleIdentifier(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	role, err := c.GetRole(ctx, instanceID, name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(getRoleIdentifier(role))
	return setRole(d, role)
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	instanceID, name, err := parseRoleIdentifier(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := c.DeleteRole(ctx, instanceID, name); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func setRole(d *schema.ResourceData, role *api.Role) diag.Diagnostics {
	if err := d.Set("name", role.Name); err != nil {
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

func getRoleIdentifier(r *api.Role) string {
	return fmt.Sprintf("%d%s%s", r.InstanceID, roleIdentifierSeparator, r.Name)
}

func parseRoleIdentifier(identifier string) (int, string, error) {
	slice := strings.Split(identifier, roleIdentifierSeparator)
	if len(slice) < 2 {
		return 0, "", errors.Errorf("invalid role identifier: %s", identifier)
	}

	instanceID, err := strconv.Atoi(slice[0])
	if err != nil {
		return 0, "", errors.Errorf("failed to parse the instance id with error: %v", err)
	}

	return instanceID, strings.Join(slice[1:], roleIdentifierSeparator), nil
}
