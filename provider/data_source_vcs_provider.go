package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceVCSProvider() *schema.Resource {
	return &schema.Resource{
		Description: "The vcs provider data source.",
		ReadContext: dataSourceVCSProviderRead,
		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The vcs provider unique resource id.",
			},
			"title": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The vcs provider title.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The vcs provider full name in vcsProviders/{resource id} format.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The vcs provider type.",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The vcs provider url.",
			},
		},
	}
}

func dataSourceVCSProviderRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	providerName := fmt.Sprintf("%s%s", internal.VCSProviderNamePrefix, d.Get("resource_id").(string))

	provider, err := c.GetVCSProvider(ctx, providerName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(provider.Name)

	return setVCSProvider(d, provider)
}

func setVCSProvider(d *schema.ResourceData, provider *v1pb.VCSProvider) diag.Diagnostics {
	providerID, err := internal.GetVCSProviderID(provider.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("resource_id", providerID); err != nil {
		return diag.Errorf("cannot set resource_id for vcs provider: %s", err.Error())
	}
	if err := d.Set("title", provider.Title); err != nil {
		return diag.Errorf("cannot set title for vcs provider: %s", err.Error())
	}
	if err := d.Set("name", provider.Name); err != nil {
		return diag.Errorf("cannot set name for vcs provider: %s", err.Error())
	}
	if err := d.Set("type", provider.Type.String()); err != nil {
		return diag.Errorf("cannot set type for vcs provider: %s", err.Error())
	}
	if err := d.Set("url", provider.Url); err != nil {
		return diag.Errorf("cannot set url for vcs provider: %s", err.Error())
	}

	return nil
}
