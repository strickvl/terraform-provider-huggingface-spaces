package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure HuggingFaceSpacesProvider satisfies various provider interfaces.
var _ provider.Provider = &HuggingFaceSpacesProvider{}

// HuggingFaceSpacesProvider defines the provider implementation.
type HuggingFaceSpacesProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// HuggingFaceSpacesProviderModel describes the provider data model.
type HuggingFaceSpacesProviderModel struct {
	Token types.String `tfsdk:"token"`
}

func (p *HuggingFaceSpacesProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "huggingface-spaces"
	resp.Version = p.version
}

func (p *HuggingFaceSpacesProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				MarkdownDescription: "The Hugging Face API token.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *HuggingFaceSpacesProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data HuggingFaceSpacesProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *HuggingFaceSpacesProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *HuggingFaceSpacesProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSpaceDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &HuggingFaceSpacesProvider{
			version: version,
		}
	}
}
