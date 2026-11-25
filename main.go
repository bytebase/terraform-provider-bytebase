// Package main is the main binary for Terraform Bytebase Provider.
package main

import (
	"context"
	"log"

	"github.com/bytebase/terraform-provider-bytebase/provider"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	server, err := provider.NewProtocol6Server(context.Background(), provider.NewProvider())
	if err != nil {
		log.Fatal(err)
	}
	plugin.Serve(&plugin.ServeOpts{
		GRPCProviderV6Func: func() tfprotov6.ProviderServer { return server },
	})
}
