package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func resourcePostgresqlRole() *schema.Resource {
	return &schema.Resource{
		Description:   "The Postgresql role resource.",
		CreateContext: resourcePGRoleCreate,
		ReadContext:   resourcePGRoleRead,
		UpdateContext: resourcePGRoleUpdate,
		DeleteContext: resourcePGRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The Postgresql role unique name.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"instance_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The instance id.",
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
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"attribute": {
				Type:     schema.TypeSet,
				MinItems: 0,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"super_user": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"no_inherit": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"create_role": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"create_db": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"can_login": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"replication": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"bypass_rls": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
		},
	}
}

func resourcePGRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	instanceID := d.Get("instance_id").(int)

	upsert := &api.PGRoleUpsert{
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

	role, err := c.CreatePGRole(ctx, instanceID, upsert)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(role.Name)
	return resourcePGRoleRead(ctx, d, m)
}

func resourcePGRoleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	instanceID := d.Get("instance_id").(int)

	upsert := &api.PGRoleUpsert{
		Name: d.Id(),
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

	role, err := c.UpdatePGRole(ctx, instanceID, upsert)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(role.Name)
	return resourcePGRoleRead(ctx, d, m)
}

func resourcePGRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	name := d.Id()
	instanceID := d.Get("instance_id").(int)

	role, err := c.GetPGRole(ctx, instanceID, name)
	if err != nil {
		return diag.FromErr(err)
	}

	return setPGRole(d, role)
}

func resourcePGRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	name := d.Id()
	instanceID := d.Get("instance_id").(int)
	var diags diag.Diagnostics

	if err := c.DeletePGRole(ctx, instanceID, name); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func setPGRole(d *schema.ResourceData, role *api.PGRole) diag.Diagnostics {
	if err := d.Set("name", role.Name); err != nil {
		return diag.Errorf("cannot set name for role: %s", err.Error())
	}
	if err := d.Set("instance_id", role.InstanceID); err != nil {
		return diag.Errorf("cannot set instance_id for role: %s", err.Error())
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

func convertRoleAttribute(d *schema.ResourceData) *api.PGRoleAttribute {
	rawList := d.Get("attribute").(*schema.Set).List()
	if len(rawList) < 1 {
		return nil
	}

	raw, ok := rawList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &api.PGRoleAttribute{
		SuperUser:   raw["super_user"].(bool),
		NoInherit:   raw["no_inherit"].(bool),
		CreateRole:  raw["create_role"].(bool),
		CreateDB:    raw["create_db"].(bool),
		CanLogin:    raw["can_login"].(bool),
		Replication: raw["replication"].(bool),
		ByPassRLS:   raw["bypass_rls"].(bool),
	}
}
