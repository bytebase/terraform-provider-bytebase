package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"google.golang.org/genproto/googleapis/type/expr"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func resourceRisk() *schema.Resource {
	return &schema.Resource{
		Description:   "The risk resource. Require ENTERPRISE subscription. Check the docs https://www.bytebase.com/docs/administration/risk-center?source=terraform for more information.",
		ReadContext:   resourceRiskRead,
		DeleteContext: resourceRiskDelete,
		CreateContext: resourceRiskCreate,
		UpdateContext: resourceRiskUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"title": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The risk title.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The risk full name in risks/{uid} format.",
			},
			"source": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					v1pb.Risk_DDL.String(),
					v1pb.Risk_DML.String(),
					v1pb.Risk_CREATE_DATABASE.String(),
					v1pb.Risk_REQUEST_QUERY.String(),
					v1pb.Risk_REQUEST_EXPORT.String(),
					v1pb.Risk_DATA_EXPORT.String(),
				}, false),
				Description: "The risk source.",
			},
			"level": {
				Type:     schema.TypeInt,
				Required: true,
				ValidateFunc: validation.IntInSlice([]int{
					300, 200, 100,
				}),
				Description: "The risk level.",
			},
			"active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "If the risk is active.",
			},
			"condition": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The risk condition.",
			},
		},
	}
}

func resourceRiskRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	fullName := d.Id()
	risk, err := c.GetRisk(ctx, fullName)
	if err != nil {
		return diag.FromErr(err)
	}

	return setRisk(d, risk)
}

func setRisk(d *schema.ResourceData, risk *v1pb.Risk) diag.Diagnostics {
	if err := d.Set("title", risk.Title); err != nil {
		return diag.Errorf("cannot set title for risk: %s", err.Error())
	}
	if err := d.Set("name", risk.Name); err != nil {
		return diag.Errorf("cannot set name for risk: %s", err.Error())
	}
	if err := d.Set("source", risk.Source.String()); err != nil {
		return diag.Errorf("cannot set source for risk: %s", err.Error())
	}
	if err := d.Set("level", int(risk.Level)); err != nil {
		return diag.Errorf("cannot set level for risk: %s", err.Error())
	}
	if err := d.Set("active", risk.Active); err != nil {
		return diag.Errorf("cannot set active for risk: %s", err.Error())
	}
	if err := d.Set("condition", risk.Condition.Expression); err != nil {
		return diag.Errorf("cannot set condition for risk: %s", err.Error())
	}

	return nil
}

func resourceRiskDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	fullName := d.Id()

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if err := c.DeleteRisk(ctx, fullName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func resourceRiskCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	title := d.Get("title").(string)
	active := d.Get("active").(bool)
	level := int32(d.Get("level").(int))
	source := v1pb.Risk_Source(v1pb.Risk_Source_value[d.Get("source").(string)])

	risk := &v1pb.Risk{
		Title:  title,
		Active: active,
		Level:  level,
		Source: source,
		Condition: &expr.Expr{
			Expression: d.Get("condition").(string),
		},
	}

	created, err := c.CreateRisk(ctx, risk)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(created.Name)

	return resourceRiskRead(ctx, d, m)
}

func resourceRiskUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	riskName := d.Id()

	existedRisk, err := c.GetRisk(ctx, riskName)
	if err != nil {
		return diag.Errorf("get risk %s failed with error: %v", riskName, err)
	}

	updateMasks := []string{}
	if d.HasChange("title") {
		updateMasks = append(updateMasks, "title")
		existedRisk.Title = d.Get("title").(string)
	}
	if d.HasChange("active") {
		updateMasks = append(updateMasks, "active")
		existedRisk.Active = d.Get("active").(bool)
	}
	if d.HasChange("level") {
		updateMasks = append(updateMasks, "level")
		existedRisk.Level = int32(d.Get("level").(int))
	}
	if d.HasChange("source") {
		updateMasks = append(updateMasks, "source")
		existedRisk.Source = v1pb.Risk_Source(v1pb.Risk_Source_value[d.Get("source").(string)])
	}
	if d.HasChange("condition") {
		updateMasks = append(updateMasks, "condition")
		existedRisk.Condition = &expr.Expr{
			Expression: d.Get("condition").(string),
		}
	}

	var diags diag.Diagnostics
	if len(updateMasks) > 0 {
		if _, err := c.UpdateRisk(ctx, existedRisk, updateMasks); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update risk",
				Detail:   fmt.Sprintf("Update risk %s failed, error: %v", riskName, err),
			})
			return diags
		}
	}

	diag := resourceRiskRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}
