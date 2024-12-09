package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceVCSProviderList() *schema.Resource {
	return &schema.Resource{
		Description: "The vcs provider data source list.",
		ReadContext: dataSourceVCSProviderListRead,
		Schema: map[string]*schema.Schema{
			"vcs_providers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The vcs provider unique resource id.",
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
				},
			},
		},
	}
}

func dataSourceVCSProviderListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	response, err := c.ListVCSProvider(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	providers := []map[string]interface{}{}
	for _, provider := range response.VcsProviders {
		providerID, err := internal.GetVCSProviderID(provider.Name)
		if err != nil {
			return diag.FromErr(err)
		}

		rawProvider := make(map[string]interface{})
		rawProvider["resource_id"] = providerID
		rawProvider["title"] = provider.Title
		rawProvider["name"] = provider.Name
		rawProvider["type"] = provider.Type.String()
		rawProvider["url"] = provider.Url

		providers = append(providers, rawProvider)
	}

	if err := d.Set("vcs_providers", providers); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
