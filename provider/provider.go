// Package provider is the implement for Bytebase Terraform provider.
package provider

import (
	"context"

	"github.com/bytebase/terraform-provider-bytebase/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// NewProvider is the implement for Bytebase Terraform provider.
func NewProvider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"bytebase_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("BYTEBASE_URL", nil),
			},
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("BYTEBASE_USER_EMAIL", nil),
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("BYTEBASE_USER_PASSWORD", nil),
			},
		},
		ConfigureContextFunc: providerConfigure,
		DataSourcesMap:       map[string]*schema.Resource{},
		ResourcesMap: map[string]*schema.Resource{
			"bytebase_environment": resourceEnvironment(),
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
			Summary:  "Unable to create HashiCups client",
			Detail:   "BYTEBASE_USER_EMAIL or BYTEBASE_USER_PASSWORD cannot be empty",
		})

		return nil, diags
	}

	if bytebaseURL == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create HashiCups client",
			Detail:   "BYTEBASE_URL cannot be empty",
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
