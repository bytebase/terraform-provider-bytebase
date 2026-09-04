// Package provider is the implement for Terraform Bytebase Provider.
package provider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/client"
)

const (
	envKeyForBytebaseURL               = "BYTEBASE_URL"
	envKeyForServiceAccount            = "BYTEBASE_SERVICE_ACCOUNT"
	envKeyForServiceKey                = "BYTEBASE_SERVICE_KEY"
	envKeyForWorkloadIdentityEmail     = "BYTEBASE_WORKLOAD_IDENTITY_EMAIL"
	envKeyForWorkloadIdentityToken     = "BYTEBASE_WORKLOAD_IDENTITY_TOKEN"
	envKeyForWorkloadIdentityTokenFile = "BYTEBASE_WORKLOAD_IDENTITY_TOKEN_FILE"

	settingKeyForURL                       = "url"
	settingKeyForServiceAccount            = "service_account"
	settingKeyForServiceKey                = "service_key"
	settingKeyForWorkloadIdentityEmail     = "workload_identity_email"
	settingKeyForWorkloadIdentityToken     = "workload_identity_token"
	settingKeyForWorkloadIdentityTokenFile = "workload_identity_token_file"
	settingKeyForCustomHeader              = "custom_header"
	settingKeyForCustomHeaderName          = "name"
	settingKeyForCustomHeaderValue         = "value"
)

var customHeaderNameRegex = regexp.MustCompile("^[!#$%&'*+.^_`|~0-9A-Za-z-]+$")

// NewProvider is the implement for Terraform Bytebase Provider.
func NewProvider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			settingKeyForURL: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyForBytebaseURL, nil),
				Description: fmt.Sprintf("The external URL for your Bytebase server. If not provided in the configuration, you must set the `%s` variable in the environment.", envKeyForBytebaseURL),
			},
			settingKeyForServiceAccount: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyForServiceAccount, nil),
				Description: fmt.Sprintf("The Bytebase service account email. If not provided in the configuration, you must set the `%s` variable in the environment.", envKeyForServiceAccount),
			},
			settingKeyForServiceKey: {
				Type:        schema.TypeString,
				Sensitive:   true,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyForServiceKey, nil),
				Description: fmt.Sprintf("The Bytebase service account key. If not provided in the configuration, you must set the `%s` variable in the environment.", envKeyForServiceKey),
			},
			settingKeyForWorkloadIdentityEmail: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyForWorkloadIdentityEmail, nil),
				Description: fmt.Sprintf("The Bytebase workload identity email. If not provided in the configuration, you must set the `%s` variable in the environment.", envKeyForWorkloadIdentityEmail),
			},
			settingKeyForWorkloadIdentityToken: {
				Type:        schema.TypeString,
				Sensitive:   true,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyForWorkloadIdentityToken, nil),
				Description: fmt.Sprintf("The external OIDC token for the Bytebase workload identity. If not provided in the configuration, you must set the `%s` variable in the environment.", envKeyForWorkloadIdentityToken),
			},
			settingKeyForWorkloadIdentityTokenFile: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyForWorkloadIdentityTokenFile, nil),
				Description: fmt.Sprintf("The path to a file containing an external OIDC token for the Bytebase workload identity. If not provided in the configuration, you must set the `%s` variable in the environment.", envKeyForWorkloadIdentityTokenFile),
			},
			settingKeyForCustomHeader: {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Custom HTTP headers to include in Bytebase API requests, for example headers required by a zero-trust gateway.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						settingKeyForCustomHeaderName: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The custom HTTP header name.",
							ValidateFunc: validation.StringMatch(
								customHeaderNameRegex,
								"must be a valid HTTP header field name",
							),
						},
						settingKeyForCustomHeaderValue: {
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
							Description: "The custom HTTP header value.",
						},
					},
				},
			},
		},
		ConfigureContextFunc: providerConfigure,
		DataSourcesMap: map[string]*schema.Resource{
			"bytebase_instance":               dataSourceInstance(),
			"bytebase_instance_list":          dataSourceInstanceList(),
			"bytebase_policy":                 dataSourcePolicy(),
			"bytebase_policy_list":            dataSourcePolicyList(),
			"bytebase_project":                dataSourceProject(),
			"bytebase_project_list":           dataSourceProjectList(),
			"bytebase_setting":                dataSourceSetting(),
			"bytebase_user":                   dataSourceUser(),
			"bytebase_user_list":              dataSourceUserList(),
			"bytebase_role":                   dataSourceRole(),
			"bytebase_role_list":              dataSourceRoleList(),
			"bytebase_group":                  dataSourceGroup(),
			"bytebase_group_list":             dataSourceGroupList(),
			"bytebase_database":               dataSourceDatabase(),
			"bytebase_database_list":          dataSourceDatabaseList(),
			"bytebase_database_group":         dataSourceDatabaseGroup(),
			"bytebase_database_group_list":    dataSourceDatabaseGroupList(),
			"bytebase_review_config":          dataSourceReviewConfig(),
			"bytebase_review_config_list":     dataSourceReviewConfigList(),
			"bytebase_iam_policy":             dataSourceIAMPolicy(),
			"bytebase_environment":            dataSourceEnvironment(),
			"bytebase_service_account":        dataSourceServiceAccount(),
			"bytebase_service_account_list":   dataSourceServiceAccountList(),
			"bytebase_workspace":              dataSourceWorkspace(),
			"bytebase_workload_identity":      dataSourceWorkloadIdentity(),
			"bytebase_workload_identity_list": dataSourceWorkloadIdentityList(),
			"bytebase_idp":                    dataSourceIdentityProvider(),
			"bytebase_idp_list":               dataSourceIdentityProviderList(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"bytebase_instance":          resourceInstance(),
			"bytebase_policy":            resourcePolicy(),
			"bytebase_project":           resourceProjct(),
			"bytebase_setting":           resourceSetting(),
			"bytebase_user":              resourceUser(),
			"bytebase_role":              resourceRole(),
			"bytebase_group":             resourceGroup(),
			"bytebase_database":          resourceDatabase(),
			"bytebase_database_group":    resourceDatabaseGroup(),
			"bytebase_review_config":     resourceReviewConfig(),
			"bytebase_iam_policy":        resourceIAMPolicy(),
			"bytebase_environment":       resourceEnvironment(),
			"bytebase_service_account":   resourceServiceAccount(),
			"bytebase_workspace":         resourceWorkspace(),
			"bytebase_workload_identity": resourceWorkloadIdentity(),
			"bytebase_idp":               resourceIdentityProvider(),
		},
	}
}

type resolvedAuthenticationConfig struct {
	ServiceAccount   *resolvedServiceAccountAuthentication
	WorkloadIdentity *resolvedWorkloadIdentityAuthentication
}

type resolvedServiceAccountAuthentication struct {
	Email string
	Key   string
}

type resolvedWorkloadIdentityAuthentication struct {
	Email     string
	Token     string
	TokenFile string
}

func resolveAuthenticationConfig(d *schema.ResourceData) (resolvedAuthenticationConfig, error) {
	serviceAccount := d.Get(settingKeyForServiceAccount).(string)
	serviceKey := d.Get(settingKeyForServiceKey).(string)
	workloadEmail := d.Get(settingKeyForWorkloadIdentityEmail).(string)
	workloadToken := d.Get(settingKeyForWorkloadIdentityToken).(string)
	workloadTokenFile := d.Get(settingKeyForWorkloadIdentityTokenFile).(string)

	hasServiceMode := serviceAccount != "" || serviceKey != ""
	hasWorkloadMode := workloadEmail != "" || workloadToken != "" || workloadTokenFile != ""
	if hasServiceMode && hasWorkloadMode {
		return resolvedAuthenticationConfig{}, errors.New("service account and workload identity authentication cannot be configured together")
	}
	if hasServiceMode {
		if serviceAccount == "" || serviceKey == "" {
			return resolvedAuthenticationConfig{}, errors.Errorf("%s and %s must be configured together", settingKeyForServiceAccount, settingKeyForServiceKey)
		}
		return resolvedAuthenticationConfig{ServiceAccount: &resolvedServiceAccountAuthentication{
			Email: serviceAccount,
			Key:   serviceKey,
		}}, nil
	}
	if hasWorkloadMode {
		if workloadEmail == "" {
			return resolvedAuthenticationConfig{}, errors.Errorf("%s is required for workload identity authentication", settingKeyForWorkloadIdentityEmail)
		}
		if (workloadToken == "") == (workloadTokenFile == "") {
			return resolvedAuthenticationConfig{}, errors.Errorf("exactly one of %s or %s must be configured", settingKeyForWorkloadIdentityToken, settingKeyForWorkloadIdentityTokenFile)
		}
		return resolvedAuthenticationConfig{WorkloadIdentity: &resolvedWorkloadIdentityAuthentication{
			Email:     workloadEmail,
			Token:     workloadToken,
			TokenFile: workloadTokenFile,
		}}, nil
	}

	return resolvedAuthenticationConfig{}, errors.New("one authentication mode must be configured")
}

func getCustomHeaders(d *schema.ResourceData) map[string]string {
	headers := map[string]string{}
	for _, item := range d.Get(settingKeyForCustomHeader).([]interface{}) {
		header := item.(map[string]interface{})
		name := http.CanonicalHeaderKey(header[settingKeyForCustomHeaderName].(string))
		value := header[settingKeyForCustomHeaderValue].(string)
		headers[name] = value
	}
	return headers
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	bytebaseURL := d.Get(settingKeyForURL).(string)
	if bytebaseURL == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create the Bytebase client",
			Detail:   fmt.Sprintf("%s cannot be empty", envKeyForBytebaseURL),
		})

		return nil, diags
	}

	resolved, err := resolveAuthenticationConfig(d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create the Bytebase client",
			Detail:   err.Error(),
		})
		return nil, diags
	}

	authentication := client.AuthenticationConfig{}
	if service := resolved.ServiceAccount; service != nil {
		authentication.ServiceAccount = &client.ServiceAccountAuthentication{
			Email: service.Email,
			Key:   service.Key,
		}
	}
	if workload := resolved.WorkloadIdentity; workload != nil {
		authentication.WorkloadIdentity = &client.WorkloadIdentityAuthentication{
			Email:     workload.Email,
			Token:     workload.Token,
			TokenFile: workload.TokenFile,
		}
	}

	c, err := client.NewClientWithAuthentication(bytebaseURL, authentication, client.WithCustomHeaders(getCustomHeaders(d)))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create the Bytebase client",
			Detail:   fmt.Sprintf("failed to authenticate to Bytebase: %v", err),
		})

		return nil, diags
	}

	return c, diags
}
