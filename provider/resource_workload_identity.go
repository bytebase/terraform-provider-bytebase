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

func resourceWorkloadIdentity() *schema.Resource {
	return &schema.Resource{
		Description:   "The workload identity resource.",
		ReadContext:   resourceWorkloadIdentityRead,
		DeleteContext: resourceWorkloadIdentityDelete,
		CreateContext: resourceWorkloadIdentityCreate,
		UpdateContext: resourceWorkloadIdentityUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"parent": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The parent resource. Format: projects/{project} for project-level, workspaces/- for workspace-level.",
				ValidateDiagFunc: internal.ResourceNameValidation(
					"^workspaces/-$",
					"^projects/[a-z]([a-z0-9-]{0,61}[a-z0-9])?$",
				),
			},
			"workload_identity_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The ID for the workload identity, which becomes part of the email.",
			},
			"title": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The display title of the workload identity.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The workload identity name in workloadIdentities/{email} format.",
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
				Optional:    true,
				MaxItems:    1,
				Description: "The workload identity configuration for OIDC token validation.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"provider_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								v1pb.WorkloadIdentityConfig_GITHUB.String(),
								v1pb.WorkloadIdentityConfig_GITLAB.String(),
							}, false),
							Description: "The provider type. Supported values: GITHUB, GITLAB.",
						},
						"issuer_url": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The OIDC Issuer URL. Auto-filled based on provider_type if not specified.",
						},
						"allowed_audiences": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "The allowed audiences for token validation.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"subject_pattern": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The subject pattern to match (e.g., \"repo:owner/repo:ref:refs/heads/main\").",
						},
					},
				},
			},
		},
	}
}

func resourceWorkloadIdentityRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	fullName := d.Id()
	wi, err := c.GetWorkloadIdentity(ctx, fullName)
	if err != nil {
		if internal.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Resource %s not found, removing from state", fullName))
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return setWorkloadIdentity(d, wi)
}

func resourceWorkloadIdentityCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	parent := d.Get("parent").(string)
	workloadIdentityID := d.Get("workload_identity_id").(string)
	title := d.Get("title").(string)

	// Check if the workload identity already exists by listing and matching the ID.
	existedWI, err := findWorkloadIdentityByID(ctx, c, workloadIdentityID)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("find workload identity %s failed with error: %v", workloadIdentityID, err))
	}

	var diags diag.Diagnostics
	if existedWI != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Workload identity already exists",
			Detail:   fmt.Sprintf("Workload identity %s already exists, try to exec the update operation", existedWI.Name),
		})

		if existedWI.State == v1pb.State_DELETED {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Workload identity is deleted",
				Detail:   fmt.Sprintf("Workload identity %s already deleted, try to undelete", existedWI.Name),
			})
			if _, err := c.UndeleteWorkloadIdentity(ctx, existedWI.Name); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to undelete workload identity",
					Detail:   fmt.Sprintf("Undelete workload identity %s failed, error: %v", existedWI.Name, err),
				})
				return diags
			}
		}

		updateMasks := []string{}
		if title != existedWI.Title {
			updateMasks = append(updateMasks, "title")
		}
		newConfig := expandWorkloadIdentityConfig(d)
		if newConfig != nil {
			updateMasks = append(updateMasks, "workload_identity_config")
		}
		if len(updateMasks) > 0 {
			if _, err := c.UpdateWorkloadIdentity(ctx, &v1pb.WorkloadIdentity{
				Name:                   existedWI.Name,
				Title:                  title,
				WorkloadIdentityConfig: newConfig,
			}, updateMasks); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to update workload identity",
					Detail:   fmt.Sprintf("Update workload identity %s failed, error: %v", existedWI.Name, err),
				})
				return diags
			}
		}
		d.SetId(existedWI.Name)
	} else {
		wi := &v1pb.WorkloadIdentity{
			Title:                  title,
			WorkloadIdentityConfig: expandWorkloadIdentityConfig(d),
		}

		created, err := c.CreateWorkloadIdentity(ctx, parent, workloadIdentityID, wi)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(created.Name)
	}

	diag := resourceWorkloadIdentityRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func findWorkloadIdentityByID(ctx context.Context, c api.Client, workloadIdentityID string) (*v1pb.WorkloadIdentity, error) {
	// The name is constructed as workloadIdentities/{id}@workload.bytebase.com.
	name := fmt.Sprintf("%s%s@workload.bytebase.com", internal.WorkloadIdentityNamePrefix, workloadIdentityID)
	wi, err := c.GetWorkloadIdentity(ctx, name)
	if err != nil {
		return nil, err
	}
	return wi, nil
}

func resourceWorkloadIdentityUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	name := d.Id()

	existedWI, err := c.GetWorkloadIdentity(ctx, name)
	if err != nil {
		return diag.Errorf("get workload identity %s failed with error: %v", name, err)
	}

	var diags diag.Diagnostics
	if existedWI.State == v1pb.State_DELETED {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Workload identity is deleted",
			Detail:   fmt.Sprintf("Workload identity %s already deleted, try to undelete", name),
		})
		if _, err := c.UndeleteWorkloadIdentity(ctx, name); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to undelete workload identity",
				Detail:   fmt.Sprintf("Undelete workload identity %s failed, error: %v", name, err),
			})
			return diags
		}
	}

	paths := []string{}
	if d.HasChange("title") {
		paths = append(paths, "title")
	}
	if d.HasChange("workload_identity_config") {
		paths = append(paths, "workload_identity_config")
	}

	if len(paths) > 0 {
		title := d.Get("title").(string)
		patch := &v1pb.WorkloadIdentity{
			Name:                   existedWI.Name,
			Title:                  title,
			WorkloadIdentityConfig: expandWorkloadIdentityConfig(d),
		}
		if _, err := c.UpdateWorkloadIdentity(ctx, patch, paths); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update workload identity",
				Detail:   fmt.Sprintf("Update workload identity %s failed, error: %v", name, err),
			})
			return diags
		}
	}

	diag := resourceWorkloadIdentityRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceWorkloadIdentityDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	return internal.ResourceDelete(ctx, d, c.DeleteWorkloadIdentity)
}

func expandWorkloadIdentityConfig(d *schema.ResourceData) *v1pb.WorkloadIdentityConfig {
	rawConfigs, ok := d.GetOk("workload_identity_config")
	if !ok {
		return nil
	}
	configs := rawConfigs.([]interface{})
	if len(configs) == 0 || configs[0] == nil {
		return nil
	}
	raw := configs[0].(map[string]interface{})

	config := &v1pb.WorkloadIdentityConfig{}

	if v, ok := raw["provider_type"].(string); ok && v != "" {
		providerValue, exists := v1pb.WorkloadIdentityConfig_ProviderType_value[v]
		if exists {
			config.ProviderType = v1pb.WorkloadIdentityConfig_ProviderType(providerValue)
		}
	}
	if v, ok := raw["issuer_url"].(string); ok {
		config.IssuerUrl = v
	}
	if v, ok := raw["allowed_audiences"].([]interface{}); ok {
		audiences := make([]string, 0, len(v))
		for _, a := range v {
			audiences = append(audiences, a.(string))
		}
		config.AllowedAudiences = audiences
	}
	if v, ok := raw["subject_pattern"].(string); ok {
		config.SubjectPattern = v
	}

	return config
}
