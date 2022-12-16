// Package main is the main binary for Terraform Bytebase Provider.
package main

import (
	"github.com/bytebase/terraform-provider-bytebase/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: provider.NewProvider,
	})
}
