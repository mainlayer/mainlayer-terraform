package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mainlayer/terraform-provider-mainlayer/internal/client"
)

// Ensure MainlayerProvider satisfies the provider.Provider interface.
var _ provider.Provider = &MainlayerProvider{}
var _ provider.ProviderWithFunctions = &MainlayerProvider{}

// MainlayerProvider is the top-level provider implementation.
type MainlayerProvider struct {
	version string
}

// MainlayerProviderModel describes the provider configuration schema.
type MainlayerProviderModel struct {
	APIKey  types.String `tfsdk:"api_key"`
	BaseURL types.String `tfsdk:"base_url"`
}

// New returns a provider factory function used by the plugin framework.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MainlayerProvider{version: version}
	}
}

// Metadata returns the provider type name and version.
func (p *MainlayerProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mainlayer"
	resp.Version = p.version
}

// Schema defines the provider-level configuration attributes.
func (p *MainlayerProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Mainlayer provider allows you to manage Mainlayer resources and plans as infrastructure code. " +
			"Configure the provider with your Mainlayer API key to get started.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The Mainlayer API key used to authenticate requests. " +
					"Can also be set via the `MAINLAYER_API_KEY` environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Override the Mainlayer API base URL. Defaults to `https://api.mainlayer.fr`. " +
					"Can also be set via the `MAINLAYER_BASE_URL` environment variable.",
				Optional: true,
			},
		},
	}
}

// Configure reads the provider configuration and builds the shared API client.
func (p *MainlayerProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config MainlayerProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve api_key: config > env var.
	apiKey := os.Getenv("MAINLAYER_API_KEY")
	if !config.APIKey.IsNull() && !config.APIKey.IsUnknown() {
		apiKey = config.APIKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The Mainlayer provider requires an API key. Set the `api_key` attribute in the provider block "+
				"or export the `MAINLAYER_API_KEY` environment variable.",
		)
		return
	}

	// Resolve base_url: config > env var > default (empty = client default).
	baseURL := os.Getenv("MAINLAYER_BASE_URL")
	if !config.BaseURL.IsNull() && !config.BaseURL.IsUnknown() {
		baseURL = config.BaseURL.ValueString()
	}

	c := client.NewClient(apiKey, baseURL)

	// Make the client available to resources and data sources.
	resp.DataSourceData = c
	resp.ResourceData = c
}

// Resources returns the list of resource types supported by this provider.
func (p *MainlayerProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewResourceResource,
		NewPlanResource,
		NewVendorResource,
	}
}

// DataSources returns the list of data source types supported by this provider.
func (p *MainlayerProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewResourcesDataSource,
	}
}

// Functions returns any provider functions (none in this provider).
func (p *MainlayerProvider) Functions(_ context.Context) []func() function.Function {
	return nil
}
