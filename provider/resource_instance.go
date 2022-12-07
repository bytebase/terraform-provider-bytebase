package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func resourceInstance() *schema.Resource {
	return &schema.Resource{
		Description:   "The instance resource.",
		CreateContext: resourceInstanceCreate,
		ReadContext:   resourceInstanceRead,
		UpdateContext: resourceInstanceUpdate,
		DeleteContext: resourceInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The instance name.",
			},
			"engine": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"MYSQL",
					"POSTGRES",
					"TIDB",
					"SNOWFLAKE",
					"CLICKHOUSE",
				}, false),
				Description: "The instance engine. Support MYSQL, POSTGRES, TIDB, SNOWFLAKE, CLICKHOUSE.",
			},
			"engine_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version for instance engine.",
			},
			"external_link": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The external console URL managing this instance (e.g. AWS RDS console, your in-house DB instance console)",
			},
			"host": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "Host or socket for your instance, or the account name if the instance type is Snowflake.",
			},
			"port": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The port for your instance.",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The connection user name used by Bytebase to perform DDL and DML operations.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    false,
				Sensitive:   true,
				Description: "The connection user password used by Bytebase to perform DDL and DML operations.",
			},
			"ssl_ca": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The CA certificate. Optional, you can set this if the engine type is CLICKHOUSE.",
			},
			"ssl_cert": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The client certificate. Optional, you can set this if the engine type is CLICKHOUSE.",
			},
			"ssl_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The client key. Optional, you can set this if the engine type is CLICKHOUSE.",
			},
			"environment": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The environment name for your instance.",
			},
		},
	}
}

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	instance, err := c.CreateInstance(ctx, &api.InstanceCreate{
		Environment:  d.Get("environment").(string),
		Name:         d.Get("name").(string),
		Engine:       d.Get("engine").(string),
		ExternalLink: d.Get("external_link").(string),
		Host:         d.Get("host").(string),
		Port:         d.Get("port").(string),
		Username:     d.Get("username").(string),
		Password:     d.Get("password").(string),
		SslCa:        d.Get("ssl_ca").(string),
		SslCert:      d.Get("ssl_cert").(string),
		SslKey:       d.Get("ssl_key").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(instance.ID))

	return resourceInstanceRead(ctx, d, m)
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	instanceID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	instance, err := c.GetInstance(ctx, instanceID)
	if err != nil {
		return diag.FromErr(err)
	}

	return setInstance(d, instance)
}

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	instanceID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	patch := &api.InstancePatch{}
	if d.HasChange("name") {
		if name, ok := d.GetOk("name"); ok {
			val := name.(string)
			patch.Name = &val
		}
	}
	if d.HasChange("external_link") {
		if link, ok := d.GetOk("external_link"); ok {
			val := link.(string)
			patch.ExternalLink = &val
		}
	}
	if d.HasChange("host") {
		if host, ok := d.GetOk("host"); ok {
			val := host.(string)
			patch.Host = &val
		}
	}
	if d.HasChange("port") {
		if port, ok := d.GetOk("port"); ok {
			val := port.(string)
			patch.Port = &val
		}
	}

	if patch.HasChange() {
		if _, err := c.UpdateInstance(ctx, instanceID, patch); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceInstanceRead(ctx, d, m)
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	instanceID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := c.DeleteInstance(ctx, instanceID); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func setInstance(d *schema.ResourceData, instance *api.Instance) diag.Diagnostics {
	if err := d.Set("name", instance.Name); err != nil {
		return diag.Errorf("cannot set name for instance: %s", err.Error())
	}
	if err := d.Set("engine", instance.Engine); err != nil {
		return diag.Errorf("cannot set engine for instance: %s", err.Error())
	}
	if err := d.Set("engine_version", instance.EngineVersion); err != nil {
		return diag.Errorf("cannot set engine_version for instance: %s", err.Error())
	}
	if err := d.Set("external_link", instance.ExternalLink); err != nil {
		return diag.Errorf("cannot set external_link for instance: %s", err.Error())
	}
	if err := d.Set("host", instance.Host); err != nil {
		return diag.Errorf("cannot set host for instance: %s", err.Error())
	}
	if err := d.Set("port", instance.Port); err != nil {
		return diag.Errorf("cannot set port for instance: %s", err.Error())
	}
	if err := d.Set("environment", instance.Environment); err != nil {
		return diag.Errorf("cannot set environment for instance: %s", err.Error())
	}
	if err := d.Set("username", instance.Username); err != nil {
		return diag.Errorf("cannot set username for instance: %s", err.Error())
	}

	return nil
}
