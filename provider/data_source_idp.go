package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func dataSourceIdentityProvider() *schema.Resource {
	return &schema.Resource{
		Description: "The identity provider data source.",
		ReadContext: dataSourceIdentityProviderRead,
		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: internal.ResourceIDValidation,
				Description:  "The identity provider unique resource id.",
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
			"oauth2_config": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The OAuth2 identity provider configuration.",
				Elem:        getDataSourceOAuth2ConfigElem(),
			},
			"oidc_config": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The OIDC identity provider configuration.",
				Elem:        getDataSourceOIDCConfigElem(),
			},
			"ldap_config": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The LDAP identity provider configuration.",
				Elem:        getDataSourceLDAPConfigElem(),
			},
		},
	}
}

func dataSourceIdentityProviderRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	idpName := fmt.Sprintf("%s%s", internal.IDPNamePrefix, d.Get("resource_id").(string))

	idp, err := c.GetIdentityProvider(ctx, idpName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(idp.Name)

	return setIdentityProvider(d, idp)
}

func getDataSourceFieldMappingElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"identifier": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The field name of the unique identifier in the IdP user info.",
			},
			"display_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The field name of the display name in the IdP user info.",
			},
			"phone": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The field name of the primary phone in the IdP user info.",
			},
			"groups": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The field name of the groups in the IdP user info.",
			},
		},
	}
}

func getDataSourceOAuth2ConfigElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"auth_url":        {Type: schema.TypeString, Computed: true, Description: "The authorization endpoint URL."},
			"token_url":       {Type: schema.TypeString, Computed: true, Description: "The token endpoint URL."},
			"user_info_url":   {Type: schema.TypeString, Computed: true, Description: "The user information endpoint URL."},
			"client_id":       {Type: schema.TypeString, Computed: true, Description: "The OAuth2 client identifier."},
			"client_secret":   {Type: schema.TypeString, Computed: true, Sensitive: true, Description: "The OAuth2 client secret."},
			"scopes":          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "The list of OAuth2 scopes."},
			"field_mapping":   {Type: schema.TypeList, Computed: true, Elem: getDataSourceFieldMappingElem(), Description: "Mapping configuration for user attributes."},
			"skip_tls_verify": {Type: schema.TypeBool, Computed: true, Description: "Whether to skip TLS certificate verification."},
			"auth_style":      {Type: schema.TypeString, Computed: true, Description: "The authentication style for client credentials."},
		},
	}
}

func getDataSourceOIDCConfigElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"issuer":          {Type: schema.TypeString, Computed: true, Description: "The OIDC issuer URL."},
			"client_id":       {Type: schema.TypeString, Computed: true, Description: "The OIDC client identifier."},
			"client_secret":   {Type: schema.TypeString, Computed: true, Sensitive: true, Description: "The OIDC client secret."},
			"scopes":          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "The OIDC scopes."},
			"field_mapping":   {Type: schema.TypeList, Computed: true, Elem: getDataSourceFieldMappingElem(), Description: "Mapping configuration for user attributes."},
			"skip_tls_verify": {Type: schema.TypeBool, Computed: true, Description: "Whether to skip TLS certificate verification."},
			"auth_style":      {Type: schema.TypeString, Computed: true, Description: "The authentication style for client credentials."},
			"auth_endpoint":   {Type: schema.TypeString, Computed: true, Description: "The authorization endpoint from OIDC well-known configuration."},
		},
	}
}

func getDataSourceLDAPConfigElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"host":              {Type: schema.TypeString, Computed: true, Description: "The hostname or IP address of the LDAP server."},
			"port":              {Type: schema.TypeInt, Computed: true, Description: "The port number of the LDAP server."},
			"skip_tls_verify":   {Type: schema.TypeBool, Computed: true, Description: "Whether to skip TLS certificate verification."},
			"bind_dn":           {Type: schema.TypeString, Computed: true, Description: "The DN of the user to bind as a service account."},
			"bind_password":     {Type: schema.TypeString, Computed: true, Sensitive: true, Description: "The password of the bind user."},
			"base_dn":           {Type: schema.TypeString, Computed: true, Description: "The base DN to search for users."},
			"user_filter":       {Type: schema.TypeString, Computed: true, Description: "The filter to search for users."},
			"security_protocol": {Type: schema.TypeString, Computed: true, Description: "The security protocol."},
			"field_mapping":     {Type: schema.TypeList, Computed: true, Elem: getDataSourceFieldMappingElem(), Description: "Mapping configuration for user attributes."},
		},
	}
}
