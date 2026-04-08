package provider

import (
	"context"
	"strconv"
	"time"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceIdentityProviderList() *schema.Resource {
	return &schema.Resource{
		Description: "The identity provider data source list.",
		ReadContext: dataSourceIdentityProviderListRead,
		Schema: map[string]*schema.Schema{
			"identity_providers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The identity provider unique resource id.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The identity provider full name in idps/{resource id} format.",
						},
						"title": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The identity provider display title.",
						},
						"domain": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The domain for email matching when using this identity provider.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The identity provider type. One of OAUTH2, OIDC, LDAP.",
						},
					},
				},
			},
		},
	}
}

func dataSourceIdentityProviderListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	idps, err := c.ListIdentityProvider(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	idpList := []map[string]interface{}{}
	for _, idp := range idps {
		raw, err := flattenIdentityProviderListItem(idp)
		if err != nil {
			return diag.FromErr(err)
		}
		idpList = append(idpList, raw)
	}

	if err := d.Set("identity_providers", idpList); err != nil {
		return diag.FromErr(err)
	}

	// always refresh
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}

func flattenIdentityProviderListItem(idp *v1pb.IdentityProvider) (map[string]interface{}, error) {
	idpID, err := internal.GetIDPID(idp.Name)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"resource_id": idpID,
		"name":        idp.Name,
		"title":       idp.Title,
		"domain":      idp.Domain,
		"type":        idp.Type.String(),
	}, nil
}
