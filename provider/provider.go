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
	envKeyForBytebaseURL   = "BYTEBASE_URL"
	envKeyForyUserEmail    = "BYTEBASE_USER_EMAIL"
	envKeyForyUserPassword = "BYTEBASE_USER_PASSWORD"
)

// NewProvider is the implement for Bytebase Terraform provider.
func NewProvider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"bytebase_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyForBytebaseURL, nil),
				Description: fmt.Sprintf("The OpenAPI URL for your Bytebase server. If not provided in the configuration, you must set the `%s` variable in the environment.", envKeyForBytebaseURL),
			},
			"email": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyForyUserEmail, nil),
				Description: fmt.Sprintf("The Bytebase user account email. If not provided in the configuration, you must set the `%s` variable in the environment.", envKeyForyUserEmail),
			},
			"password": {
				Type:        schema.TypeString,
				Sensitive:   true,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(envKeyForyUserPassword, nil),
				Description: fmt.Sprintf("The Bytebase user account password. If not provided in the configuration, you must set the `%s` variable in the environment.", envKeyForyUserPassword),
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

	email := d.Get("email").(string)
	password := d.Get("password").(string)
	bytebaseURL := d.Get("bytebase_url").(string)

	if email == "" || password == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create the Bytebase client",
			Detail:   fmt.Sprintf("%s or %s cannot be empty", envKeyForyUserEmail, envKeyForyUserPassword),
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

	c, err := client.NewClient(bytebaseURL, email, password)
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
