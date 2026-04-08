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

func resourceIdentityProvider() *schema.Resource {
	return &schema.Resource{
		Description:   "The identity provider resource.",
		ReadContext:   resourceIdentityProviderRead,
		CreateContext: resourceIdentityProviderCreate,
		UpdateContext: resourceIdentityProviderUpdate,
		DeleteContext: resourceIdentityProviderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The identity provider display title.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The domain for email matching when using this identity provider.",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The identity provider type. One of OAUTH2, OIDC, LDAP.",
				ValidateFunc: validation.StringInSlice([]string{"OAUTH2", "OIDC", "LDAP"}, false),
			},
			"oauth2_config": getOAuth2ConfigSchema(),
			"oidc_config":   getOIDCConfigSchema(),
			"ldap_config":   getLDAPConfigSchema(),
		},
	}
}

func getFieldMappingSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"identifier": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The field name of the unique identifier in the IdP user info.",
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The field name of the display name in the IdP user info.",
			},
			"phone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The field name of the primary phone in the IdP user info.",
			},
			"groups": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The field name of the groups in the IdP user info.",
			},
		},
	}
}

func getOAuth2ConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "The OAuth2 identity provider configuration.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"auth_url": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The authorization endpoint URL.",
				},
				"token_url": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The token endpoint URL.",
				},
				"user_info_url": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The user information endpoint URL.",
				},
				"client_id": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The OAuth2 client identifier.",
				},
				"client_secret": {
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
					Description: "The OAuth2 client secret.",
				},
				"scopes": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "The list of OAuth2 scopes to request.",
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"field_mapping": {
					Type:        schema.TypeList,
					Required:    true,
					MaxItems:    1,
					Description: "Mapping configuration for user attributes.",
					Elem:        getFieldMappingSchema(),
				},
				"skip_tls_verify": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Whether to skip TLS certificate verification.",
				},
				"auth_style": {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "IN_PARAMS",
					Description:  "The authentication style for client credentials. One of IN_PARAMS, IN_HEADER.",
					ValidateFunc: validation.StringInSlice([]string{"IN_PARAMS", "IN_HEADER"}, false),
				},
			},
		},
	}
}

func getOIDCConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "The OIDC identity provider configuration.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"issuer": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The OIDC issuer URL.",
				},
				"client_id": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The OIDC client identifier.",
				},
				"client_secret": {
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
					Description: "The OIDC client secret.",
				},
				"scopes": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "The OIDC scopes to request.",
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"field_mapping": {
					Type:        schema.TypeList,
					Required:    true,
					MaxItems:    1,
					Description: "Mapping configuration for user attributes from OIDC claims.",
					Elem:        getFieldMappingSchema(),
				},
				"skip_tls_verify": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Whether to skip TLS certificate verification.",
				},
				"auth_style": {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "IN_PARAMS",
					Description:  "The authentication style for client credentials. One of IN_PARAMS, IN_HEADER.",
					ValidateFunc: validation.StringInSlice([]string{"IN_PARAMS", "IN_HEADER"}, false),
				},
				"auth_endpoint": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The authorization endpoint from the OIDC well-known configuration.",
				},
			},
		},
	}
}

func getLDAPConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "The LDAP identity provider configuration.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"host": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The hostname or IP address of the LDAP server.",
				},
				"port": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "The port number of the LDAP server. Default port depends on security protocol.",
				},
				"skip_tls_verify": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Whether to skip TLS certificate verification.",
				},
				"bind_dn": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The DN of the user to bind as a service account.",
				},
				"bind_password": {
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
					Description: "The password of the bind user.",
				},
				"base_dn": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The base DN to search for users.",
				},
				"user_filter": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The filter to search for users, e.g. (uid=%s).",
				},
				"security_protocol": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "The security protocol. One of START_TLS, LDAPS.",
					ValidateFunc: validation.StringInSlice([]string{"START_TLS", "LDAPS"}, false),
				},
				"field_mapping": {
					Type:        schema.TypeList,
					Required:    true,
					MaxItems:    1,
					Description: "Mapping configuration for user attributes from LDAP response.",
					Elem:        getFieldMappingSchema(),
				},
			},
		},
	}
}

func resourceIdentityProviderRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	fullName := d.Id()
	idp, err := c.GetIdentityProvider(ctx, fullName)
	if err != nil {
		if internal.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Resource %s not found, removing from state", fullName))
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return setIdentityProvider(d, idp)
}

func resourceIdentityProviderCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)

	idpID := d.Get("resource_id").(string)
	idpName := fmt.Sprintf("%s%s", internal.IDPNamePrefix, idpID)

	existedIDP, err := c.GetIdentityProvider(ctx, idpName)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("get identity provider %s failed with error: %v", idpName, err))
	}

	idp, diags := getIdentityProviderFromSchema(d, idpName)
	if diags.HasError() {
		return diags
	}

	if existedIDP != nil && err == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Identity provider already exists",
			Detail:   fmt.Sprintf("Identity provider %s already exists, try to exec the update operation", idpName),
		})

		updateMasks := []string{"title", "domain", "config"}
		if _, err := c.UpdateIdentityProvider(ctx, idp, updateMasks); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update identity provider",
				Detail:   fmt.Sprintf("Update identity provider %s failed, error: %v", idpName, err),
			})
			return diags
		}
	} else {
		if _, err := c.CreateIdentityProvider(ctx, idpID, idp); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(idpName)

	diag := resourceIdentityProviderRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceIdentityProviderUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("resource_id") {
		return diag.Errorf("cannot change the resource id")
	}

	c := m.(api.Client)
	idpName := d.Id()

	idp, diags := getIdentityProviderFromSchema(d, idpName)
	if diags.HasError() {
		return diags
	}

	var updateMasks []string
	if d.HasChange("title") {
		updateMasks = append(updateMasks, "title")
	}
	if d.HasChange("domain") {
		updateMasks = append(updateMasks, "domain")
	}
	if d.HasChange("oauth2_config") || d.HasChange("oidc_config") || d.HasChange("ldap_config") {
		updateMasks = append(updateMasks, "config")
	}

	if len(updateMasks) > 0 {
		if _, err := c.UpdateIdentityProvider(ctx, idp, updateMasks); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update identity provider",
				Detail:   fmt.Sprintf("Update identity provider %s failed, error: %v", idpName, err),
			})
			return diags
		}
	}

	diag := resourceIdentityProviderRead(ctx, d, m)
	if diag != nil {
		diags = append(diags, diag...)
	}

	return diags
}

func resourceIdentityProviderDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	return internal.ResourceDelete(ctx, d, c.DeleteIdentityProvider)
}

func getIdentityProviderFromSchema(d *schema.ResourceData, idpName string) (*v1pb.IdentityProvider, diag.Diagnostics) {
	idpType := d.Get("type").(string)

	idp := &v1pb.IdentityProvider{
		Name:   idpName,
		Title:  d.Get("title").(string),
		Domain: d.Get("domain").(string),
		Type:   v1pb.IdentityProviderType(v1pb.IdentityProviderType_value[idpType]),
	}

	config := &v1pb.IdentityProviderConfig{}

	switch idpType {
	case "OAUTH2":
		rawList := d.Get("oauth2_config").([]interface{})
		if len(rawList) == 0 {
			return nil, diag.Errorf("oauth2_config is required when type is OAUTH2")
		}
		raw := rawList[0].(map[string]interface{})
		oauth2Config := &v1pb.OAuth2IdentityProviderConfig{
			AuthUrl:       raw["auth_url"].(string),
			TokenUrl:      raw["token_url"].(string),
			UserInfoUrl:   raw["user_info_url"].(string),
			ClientId:      raw["client_id"].(string),
			ClientSecret:  raw["client_secret"].(string),
			SkipTlsVerify: raw["skip_tls_verify"].(bool),
			AuthStyle:     v1pb.OAuth2AuthStyle(v1pb.OAuth2AuthStyle_value[raw["auth_style"].(string)]),
		}
		for _, s := range raw["scopes"].([]interface{}) {
			oauth2Config.Scopes = append(oauth2Config.Scopes, s.(string))
		}
		oauth2Config.FieldMapping = getFieldMappingFromSchema(raw["field_mapping"].([]interface{}))
		config.Config = &v1pb.IdentityProviderConfig_Oauth2Config{Oauth2Config: oauth2Config}

	case "OIDC":
		rawList := d.Get("oidc_config").([]interface{})
		if len(rawList) == 0 {
			return nil, diag.Errorf("oidc_config is required when type is OIDC")
		}
		raw := rawList[0].(map[string]interface{})
		oidcConfig := &v1pb.OIDCIdentityProviderConfig{
			Issuer:        raw["issuer"].(string),
			ClientId:      raw["client_id"].(string),
			ClientSecret:  raw["client_secret"].(string),
			SkipTlsVerify: raw["skip_tls_verify"].(bool),
			AuthStyle:     v1pb.OAuth2AuthStyle(v1pb.OAuth2AuthStyle_value[raw["auth_style"].(string)]),
		}
		for _, s := range raw["scopes"].([]interface{}) {
			oidcConfig.Scopes = append(oidcConfig.Scopes, s.(string))
		}
		oidcConfig.FieldMapping = getFieldMappingFromSchema(raw["field_mapping"].([]interface{}))
		config.Config = &v1pb.IdentityProviderConfig_OidcConfig{OidcConfig: oidcConfig}

	case "LDAP":
		rawList := d.Get("ldap_config").([]interface{})
		if len(rawList) == 0 {
			return nil, diag.Errorf("ldap_config is required when type is LDAP")
		}
		raw := rawList[0].(map[string]interface{})
		ldapConfig := &v1pb.LDAPIdentityProviderConfig{
			Host:          raw["host"].(string),
			Port:          int32(raw["port"].(int)),
			SkipTlsVerify: raw["skip_tls_verify"].(bool),
			BindDn:        raw["bind_dn"].(string),
			BindPassword:  raw["bind_password"].(string),
			BaseDn:        raw["base_dn"].(string),
			UserFilter:    raw["user_filter"].(string),
		}
		if v := raw["security_protocol"].(string); v != "" {
			ldapConfig.SecurityProtocol = v1pb.LDAPIdentityProviderConfig_SecurityProtocol(
				v1pb.LDAPIdentityProviderConfig_SecurityProtocol_value[v],
			)
		}
		ldapConfig.FieldMapping = getFieldMappingFromSchema(raw["field_mapping"].([]interface{}))
		config.Config = &v1pb.IdentityProviderConfig_LdapConfig{LdapConfig: ldapConfig}

	default:
		return nil, diag.Errorf("unsupported identity provider type: %s", idpType)
	}

	idp.Config = config
	return idp, nil
}

func getFieldMappingFromSchema(rawList []interface{}) *v1pb.FieldMapping {
	if len(rawList) == 0 {
		return nil
	}
	raw := rawList[0].(map[string]interface{})
	return &v1pb.FieldMapping{
		Identifier:  raw["identifier"].(string),
		DisplayName: raw["display_name"].(string),
		Phone:       raw["phone"].(string),
		Groups:      raw["groups"].(string),
	}
}

func setIdentityProvider(d *schema.ResourceData, idp *v1pb.IdentityProvider) diag.Diagnostics {
	idpID, err := internal.GetIDPID(idp.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("resource_id", idpID); err != nil {
		return diag.Errorf("cannot set resource_id: %s", err.Error())
	}
	if err := d.Set("name", idp.Name); err != nil {
		return diag.Errorf("cannot set name: %s", err.Error())
	}
	if err := d.Set("title", idp.Title); err != nil {
		return diag.Errorf("cannot set title: %s", err.Error())
	}
	if err := d.Set("domain", idp.Domain); err != nil {
		return diag.Errorf("cannot set domain: %s", err.Error())
	}
	if err := d.Set("type", idp.Type.String()); err != nil {
		return diag.Errorf("cannot set type: %s", err.Error())
	}

	if config := idp.GetConfig(); config != nil {
		switch {
		case config.GetOauth2Config() != nil:
			if err := d.Set("oauth2_config", flattenOAuth2Config(config.GetOauth2Config())); err != nil {
				return diag.Errorf("cannot set oauth2_config: %s", err.Error())
			}
		case config.GetOidcConfig() != nil:
			if err := d.Set("oidc_config", flattenOIDCConfig(config.GetOidcConfig())); err != nil {
				return diag.Errorf("cannot set oidc_config: %s", err.Error())
			}
		case config.GetLdapConfig() != nil:
			if err := d.Set("ldap_config", flattenLDAPConfig(config.GetLdapConfig())); err != nil {
				return diag.Errorf("cannot set ldap_config: %s", err.Error())
			}
		}
	}

	return nil
}

func flattenFieldMapping(fm *v1pb.FieldMapping) []interface{} {
	if fm == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"identifier":   fm.Identifier,
			"display_name": fm.DisplayName,
			"phone":        fm.Phone,
			"groups":       fm.Groups,
		},
	}
}

func flattenOAuth2Config(c *v1pb.OAuth2IdentityProviderConfig) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"auth_url":        c.AuthUrl,
			"token_url":       c.TokenUrl,
			"user_info_url":   c.UserInfoUrl,
			"client_id":       c.ClientId,
			"client_secret":   c.ClientSecret,
			"scopes":          c.Scopes,
			"field_mapping":   flattenFieldMapping(c.FieldMapping),
			"skip_tls_verify": c.SkipTlsVerify,
			"auth_style":      c.AuthStyle.String(),
		},
	}
}

func flattenOIDCConfig(c *v1pb.OIDCIdentityProviderConfig) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"issuer":          c.Issuer,
			"client_id":       c.ClientId,
			"client_secret":   c.ClientSecret,
			"scopes":          c.Scopes,
			"field_mapping":   flattenFieldMapping(c.FieldMapping),
			"skip_tls_verify": c.SkipTlsVerify,
			"auth_style":      c.AuthStyle.String(),
			"auth_endpoint":   c.AuthEndpoint,
		},
	}
}

func flattenLDAPConfig(c *v1pb.LDAPIdentityProviderConfig) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"host":              c.Host,
			"port":              int(c.Port),
			"skip_tls_verify":   c.SkipTlsVerify,
			"bind_dn":           c.BindDn,
			"bind_password":     c.BindPassword,
			"base_dn":           c.BaseDn,
			"user_filter":       c.UserFilter,
			"security_protocol": c.SecurityProtocol.String(),
			"field_mapping":     flattenFieldMapping(c.FieldMapping),
		},
	}
}
