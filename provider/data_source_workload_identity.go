package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceWorkloadIdentity() *schema.Resource {
	return &schema.Resource{
		Description: "The workload identity data source.",
		ReadContext: dataSourceWorkloadIdentityRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: internal.ResourceNameValidation(
					fmt.Sprintf("^%s", internal.WorkloadIdentityNamePrefix),
				),
				Description: "The workload identity name in workloadIdentities/{email} format.",
			},
			"title": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The display title of the workload identity.",
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The workload identity email.",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The workload identity state.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The timestamp when the workload identity was created.",
			},
			"workload_identity_config": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The workload identity configuration for OIDC token validation.",
				Elem: &schema.Resource{
					Schema: getWorkloadIdentityConfigSchema(),
				},
			},
		},
	}
}

func getWorkloadIdentityConfigSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"provider_type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The provider type. Supported values: GITHUB, GITLAB.",
		},
		"issuer_url": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The OIDC Issuer URL.",
		},
		"allowed_audiences": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "The allowed audiences for token validation.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"subject_pattern": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The subject pattern to match.",
		},
	}
}

func dataSourceWorkloadIdentityRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	name := d.Get("name").(string)

	wi, err := c.GetWorkloadIdentity(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(wi.Name)

	return setWorkloadIdentity(d, wi)
}

func setWorkloadIdentity(d *schema.ResourceData, wi *v1pb.WorkloadIdentity) diag.Diagnostics {
	if err := d.Set("name", wi.Name); err != nil {
		return diag.Errorf("cannot set name for workload identity: %s", err.Error())
	}
	if err := d.Set("title", wi.Title); err != nil {
		return diag.Errorf("cannot set title for workload identity: %s", err.Error())
	}
	if err := d.Set("email", wi.Email); err != nil {
		return diag.Errorf("cannot set email for workload identity: %s", err.Error())
	}
	if err := d.Set("state", wi.State.String()); err != nil {
		return diag.Errorf("cannot set state for workload identity: %s", err.Error())
	}
	if wi.CreateTime != nil {
		if err := d.Set("create_time", wi.CreateTime.AsTime().UTC().Format(time.RFC3339)); err != nil {
			return diag.Errorf("cannot set create_time for workload identity: %s", err.Error())
		}
	}
	if err := d.Set("workload_identity_config", flattenWorkloadIdentityConfig(wi.WorkloadIdentityConfig)); err != nil {
		return diag.Errorf("cannot set workload_identity_config for workload identity: %s", err.Error())
	}
	return nil
}

func flattenWorkloadIdentityConfig(config *v1pb.WorkloadIdentityConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"provider_type":     config.ProviderType.String(),
			"issuer_url":        config.IssuerUrl,
			"allowed_audiences": config.AllowedAudiences,
			"subject_pattern":   config.SubjectPattern,
		},
	}
}
