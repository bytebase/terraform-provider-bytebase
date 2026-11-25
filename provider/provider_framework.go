package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	frameworkschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// NewProtocol6Server combines SDK resources with the keyed IAM Framework resource.
func NewProtocol6Server(ctx context.Context, legacy *schema.Provider) (tfprotov6.ProviderServer, error) {
	if legacy == nil {
		return nil, fmt.Errorf("legacy provider must not be nil")
	}
	legacyCopy := *legacy
	legacyCopy.ResourcesMap = make(map[string]*schema.Resource, len(legacy.ResourcesMap))
	for name, implementation := range legacy.ResourcesMap {
		if name != "bytebase_iam_policy" {
			legacyCopy.ResourcesMap[name] = implementation
		}
	}
	legacyServer, err := tf5to6server.UpgradeServer(ctx, legacyCopy.GRPCProvider)
	if err != nil {
		return nil, err
	}
	framework := &iamFrameworkProvider{legacy: &legacyCopy}
	// Configuration runs in this order so both implementations share one SDK client.
	mux, err := tf6muxserver.NewMuxServer(ctx,
		func() tfprotov6.ProviderServer { return legacyServer },
		providerserver.NewProtocol6(framework),
	)
	if err != nil {
		return nil, err
	}
	return mux.ProviderServer(), nil
}

type iamFrameworkProvider struct {
	legacy *schema.Provider
}

var _ frameworkprovider.Provider = (*iamFrameworkProvider)(nil)

func (*iamFrameworkProvider) Metadata(_ context.Context, _ frameworkprovider.MetadataRequest, resp *frameworkprovider.MetadataResponse) {
	resp.TypeName = "bytebase"
}

func (p *iamFrameworkProvider) Schema(_ context.Context, _ frameworkprovider.SchemaRequest, resp *frameworkprovider.SchemaResponse) {
	attributes := make(map[string]frameworkschema.Attribute, len(p.legacy.Schema)-1)
	for name, field := range p.legacy.Schema {
		if name != settingKeyForCustomHeader {
			attributes[name] = frameworkProviderStringAttribute(field)
		}
	}
	customHeader := p.legacy.Schema[settingKeyForCustomHeader]
	customHeaderFields := customHeader.Elem.(*schema.Resource).Schema
	resp.Schema = frameworkschema.Schema{
		Attributes: attributes,
		Blocks: map[string]frameworkschema.Block{
			settingKeyForCustomHeader: frameworkschema.ListNestedBlock{
				Description: customHeader.Description,
				NestedObject: frameworkschema.NestedBlockObject{
					Attributes: map[string]frameworkschema.Attribute{
						settingKeyForCustomHeaderName:  frameworkProviderStringAttribute(customHeaderFields[settingKeyForCustomHeaderName]),
						settingKeyForCustomHeaderValue: frameworkProviderStringAttribute(customHeaderFields[settingKeyForCustomHeaderValue]),
					},
				},
			},
		},
	}
}

func frameworkProviderStringAttribute(field *schema.Schema) frameworkschema.StringAttribute {
	return frameworkschema.StringAttribute{
		Description: field.Description,
		Required:    field.Required,
		Optional:    field.Optional,
		Sensitive:   field.Sensitive,
	}
}

func (p *iamFrameworkProvider) Configure(_ context.Context, _ frameworkprovider.ConfigureRequest, resp *frameworkprovider.ConfigureResponse) {
	resp.ResourceData = p.legacy.Meta()
}

func (*iamFrameworkProvider) DataSources(context.Context) []func() datasource.DataSource {
	return nil
}

func (*iamFrameworkProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{newIAMPolicyFrameworkResource}
}
