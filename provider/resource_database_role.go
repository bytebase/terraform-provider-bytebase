package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func resourceInstanceRole() *schema.Resource {
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

	instanceID := d.Get("instance").(string)
	roleName := d.Get("name").(string)

	instance, err := c.GetInstance(ctx, &api.InstanceFindMessage{
		InstanceID: instanceID,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if instance.Engine != api.EngineTypePostgres {
		return diag.Errorf("resource_database_role only supports the instance with POSTGRES type")
	}

	upsert := &api.RoleUpsert{
		RoleName:  roleName,
		Attribute: convertRoleAttributeToString(d),
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

	existedRole, err := c.GetRole(ctx, instanceID, roleName)
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

		role, err := c.UpdateRole(ctx, instanceID, roleName, upsert)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(role.Name)
	} else {
		role, err := c.CreateRole(ctx, instanceID, upsert)
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
	if d.HasChange("instance") {
		return diag.Errorf("cannot change the instance")
	}

	c := m.(api.Client)

	instanceID, roleName, err := internal.GetInstanceRoleID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err := c.GetRole(ctx, instanceID, roleName); err != nil {
		return diag.FromErr(err)
	}

	upsert := &api.RoleUpsert{
		RoleName: roleName,
	}

	if d.HasChange("name") {
		if v := d.Get("name").(string); v != "" {
			upsert.RoleName = v
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
		upsert.Attribute = convertRoleAttributeToString(d)
	}

	if _, err := c.UpdateRole(ctx, instanceID, roleName, upsert); err != nil {
		return diag.FromErr(err)
	}

	role, err := c.GetRole(ctx, instanceID, upsert.RoleName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(role.Name)
	return setRole(d, role)
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	instanceID, roleName, err := internal.GetInstanceRoleID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	role, err := c.GetRole(ctx, instanceID, roleName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(role.Name)
	return setRole(d, role)
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	instanceID, roleName, err := internal.GetInstanceRoleID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := c.DeleteRole(ctx, instanceID, roleName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func setRole(d *schema.ResourceData, role *api.Role) diag.Diagnostics {
	if err := d.Set("name", role.RoleName); err != nil {
		return diag.Errorf("cannot set name for role: %s", err.Error())
	}
	if err := d.Set("connection_limit", role.ConnectionLimit); err != nil {
		return diag.Errorf("cannot set connection_limit for role: %s", err.Error())
	}
	if err := d.Set("valid_until", role.ValidUntil); err != nil {
		return diag.Errorf("cannot set valid_until for role: %s", err.Error())
	}

	if err := d.Set("attribute", []interface{}{convertStringToRoleAttribute(role.Attribute)}); err != nil {
		return diag.Errorf("cannot set attribute for role: %s", err.Error())
	}

	return nil
}

func convertRoleAttributeToString(d *schema.ResourceData) *string {
	rawList := d.Get("attribute").([]interface{})
	if len(rawList) < 1 {
		return nil
	}

	raw, ok := rawList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	attributes := []string{}
	if raw["super_user"].(bool) {
		attributes = append(attributes, "SUPERUSER")
	} else {
		attributes = append(attributes, "NOSUPERUSER")
	}
	if raw["no_inherit"].(bool) {
		attributes = append(attributes, "NOINHERIT")
	} else {
		attributes = append(attributes, "INHERIT")
	}
	if raw["create_role"].(bool) {
		attributes = append(attributes, "CREATEROLE")
	} else {
		attributes = append(attributes, "NOCREATEROLE")
	}
	if raw["create_db"].(bool) {
		attributes = append(attributes, "CREATEDB")
	} else {
		attributes = append(attributes, "NOCREATEDB")
	}
	if raw["can_login"].(bool) {
		attributes = append(attributes, "LOGIN")
	} else {
		attributes = append(attributes, "NOLOGIN")
	}
	if raw["replication"].(bool) {
		attributes = append(attributes, "REPLICATION")
	} else {
		attributes = append(attributes, "NOREPLICATION")
	}
	if raw["bypass_rls"].(bool) {
		attributes = append(attributes, "BYPASSRLS")
	} else {
		attributes = append(attributes, "NOBYPASSRLS")
	}

	resp := strings.Join(attributes, " ")
	return &resp
}

func convertStringToRoleAttribute(attribute *string) map[string]interface{} {
	attr := []string{}
	if attribute != nil {
		attr = strings.Split(*attribute, " ")
	}
	return map[string]interface{}{
		"super_user":  containsAttribute(attr, "SUPERUSER"),
		"no_inherit":  containsAttribute(attr, "NOINHERIT"),
		"create_role": containsAttribute(attr, "CREATEROLE"),
		"create_db":   containsAttribute(attr, "CREATEDB"),
		"can_login":   containsAttribute(attr, "LOGIN"),
		"replication": containsAttribute(attr, "REPLICATION"),
		"bypass_rls":  containsAttribute(attr, "BYPASSRLS"),
	}
}

func containsAttribute(attributes []string, target string) bool {
	for _, attr := range attributes {
		if attr == target {
			return true
		}
	}
	return false
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
