package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func resourceVCSProvider() *schema.Resource {
	return &schema.Resource{
		Description:   "The vcs provider resource.",
		CreateContext: resourceVCSProviderCreate,
		ReadContext:   resourceVCSProviderRead,
		UpdateContext: resourceVCSProviderUpdate,
		DeleteContext: resourceVCSProviderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The vcs provider unique resource id.",
			},
			"title": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The vcs provider title.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The vcs provider full name in vcsProviders/{resource id} format.",
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The vcs provider url. You need to provide the url if you're using the self-host GitLab or self-host GitHub.",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The vcs provider type.",
				ValidateFunc: validation.StringInSlice([]string{
					v1pb.VCSType_GITHUB.String(),
					v1pb.VCSType_GITLAB.String(),
					v1pb.VCSType_BITBUCKET.String(),
					v1pb.VCSType_AZURE_DEVOPS.String(),
				}, false),
			},
			"access_token": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The vcs provider token. Check the docs https://bytebase.cc/docs/vcs-integration/add-git-provider for details.",
			},
		},
	}
}

func resourceVCSProviderRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	fullName := d.Id()
	provider, err := c.GetVCSProvider(ctx, fullName)
	if err != nil {
		return diag.FromErr(err)
	}

	return setVCSProvider(d, provider)
}

func resourceVCSProviderDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	fullName := d.Id()

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if err := c.DeleteVCSProvider(ctx, fullName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func resourceVCSProviderCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	providerID := d.Get("resource_id").(string)
	providerName := fmt.Sprintf("%s%s", internal.VCSProviderNamePrefix, providerID)

	existedProvider, err := c.GetVCSProvider(ctx, providerName)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get vcs provider %s failed with error: %v", providerName, err))
	}

	title := d.Get("title").(string)
	url := d.Get("url").(string)
	vcsType := v1pb.VCSType(v1pb.VCSType_value[d.Get("type").(string)])
	token := d.Get("access_token").(string)

	var diags diag.Diagnostics
	if existedProvider != nil && err == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "VCS provider already exists",
			Detail:   fmt.Sprintf("VCS provider %s already exists, try to exec the update operation", providerName),
		})

		updateMasks := []string{}
		if token != "" {
			updateMasks = append(updateMasks, "access_token")
		}
		if title != "" && title != existedProvider.Title {
			updateMasks = append(updateMasks, "title")
		}

		if len(updateMasks) > 0 {
			if _, err := c.UpdateVCSProvider(ctx, &v1pb.VCSProvider{
				Name:        providerName,
				Title:       title,
				Url:         existedProvider.Url,
				Type:        existedProvider.Type,
				AccessToken: token,
			}, updateMasks); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to update vcs provider",
					Detail:   fmt.Sprintf("Update vcs provider %s failed, error: %v", providerName, err),
				})
				return diags
			}
		}
	} else {
		switch vcsType {
		case v1pb.VCSType_GITHUB:
			if url == "" {
				url = "https://github.com"
			}
		case v1pb.VCSType_AZURE_DEVOPS:
			url = "https://dev.azure.com"
		case v1pb.VCSType_GITLAB:
			if url == "" {
				url = "https://gitlab.com"
			}
		case v1pb.VCSType_BITBUCKET:
			url = "https://bitbucket.org"
		}
		if url == "" {
			return diag.Errorf("missing url")
		}
		if _, err := c.CreateVCSProvider(ctx, providerID, &v1pb.VCSProvider{
			Name:        providerName,
			Title:       title,
			Url:         url,
			Type:        vcsType,
			AccessToken: token,
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(providerName)

	diag := resourceVCSProviderRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceVCSProviderUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("resource_id") {
		return diag.Errorf("cannot change the resource id")
	}

	c := m.(api.Client)
	providerName := d.Id()

	existedProvider, err := c.GetVCSProvider(ctx, providerName)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get vcs provider %s failed with error: %v", providerName, err))
		return diag.FromErr(err)
	}

	paths := []string{}
	if d.HasChange("type") {
		return diag.Errorf("cannot change the vcs provider type")
	}
	if d.HasChange("title") {
		paths = append(paths, "title")
	}
	if d.HasChange("access_token") {
		paths = append(paths, "access_token")
	}

	var diags diag.Diagnostics
	if len(paths) > 0 {
		if _, err := c.UpdateVCSProvider(ctx, &v1pb.VCSProvider{
			Name:        providerName,
			Title:       d.Get("title").(string),
			Url:         existedProvider.Url,
			Type:        existedProvider.Type,
			AccessToken: d.Get("access_token").(string),
		}, paths); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update vcs provider",
				Detail:   fmt.Sprintf("Update vcs provider %s failed, error: %v", providerName, err),
			})
			return diags
		}
	}

	diag := resourceVCSProviderRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}
