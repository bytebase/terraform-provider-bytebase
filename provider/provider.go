// Package provider is the implement for Bytebase Terraform provider.
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/client"
)

const (
	openAPIVersion = "v1"

	envKeyForBytebaseURL    = "BYTEBASE_URL"
	envKeyForServiceAccount = "BYTEBASE_SERVICE_ACCOUNT"
	envKeyForServiceKey     = "BYTEBASE_SERVICE_KEY"

	settingKeyForURL            = "url"
	settingKeyForServiceAccount = "service_account"
	settingKeyForServiceKey     = "service_key"
)

// NewProvider is the implement for Bytebase Terraform provider.
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
		},
		ConfigureContextFunc: providerConfigure,
		DataSourcesMap: map[string]*schema.Resource{
			"bytebase_instances":    dataSourceInstanceList(),
			"bytebase_environments": dataSourceEnvironmentList(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"bytebase_environment": resourceEnvironment(),
			"bytebase_instance":    resourceInstance(),
		},
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	email := d.Get(settingKeyForServiceAccount).(string)
	key := d.Get(settingKeyForServiceKey).(string)
	bytebaseURL := d.Get(settingKeyForURL).(string)

	if email == "" || key == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create the Bytebase client",
			Detail:   fmt.Sprintf("%s or %s cannot be empty", envKeyForServiceAccount, envKeyForServiceKey),
		})

		return nil, diags
	}

	if bytebaseURL == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create the Bytebase client",
			Detail:   fmt.Sprintf("%s cannot be empty", envKeyForBytebaseURL),
		})

		return nil, diags
	}

	c, err := client.NewClient(fmt.Sprintf("%s/%s", bytebaseURL, openAPIVersion), email, key)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Bytebase client",
			Detail:   err.Error(),
		})

		return nil, diags
	}

	return c, diags
}
