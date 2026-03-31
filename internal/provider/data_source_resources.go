package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mainlayer/terraform-provider-mainlayer/internal/client"
)

// Ensure ResourcesDataSource satisfies the datasource.DataSource interface.
var _ datasource.DataSource = &ResourcesDataSource{}

// ResourcesDataSource reads all Mainlayer resources visible to the API key.
type ResourcesDataSource struct {
	client *client.Client
}

// ResourcesDataSourceModel is the Terraform model for the data source.
type ResourcesDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Resources types.List   `tfsdk:"resources"`
}

// resourceItemAttrTypes defines the object shape for each item in Resources.
var resourceItemAttrTypes = map[string]attr.Type{
	"id":           types.StringType,
	"slug":         types.StringType,
	"type":         types.StringType,
	"price_usdc":   types.Float64Type,
	"fee_model":    types.StringType,
	"description":  types.StringType,
	"callback_url": types.StringType,
	"created_at":   types.StringType,
	"updated_at":   types.StringType,
}

// NewResourcesDataSource is the factory function for the data source.
func NewResourcesDataSource() datasource.DataSource {
	return &ResourcesDataSource{}
}

// Metadata sets the data source type name.
func (d *ResourcesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resources"
}

// Schema defines the Terraform schema for the data source.
func (d *ResourcesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Mainlayer resources associated with the authenticated API key. " +
			"Use this data source to reference existing resources without managing them in Terraform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A placeholder identifier for the data source (always `mainlayer_resources`).",
			},
			"resources": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of resources returned by the API.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the resource.",
						},
						"slug": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The URL-safe identifier for the resource.",
						},
						"type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The resource type (e.g. `api`, `tool`, `model`).",
						},
						"price_usdc": schema.Float64Attribute{
							Computed:            true,
							MarkdownDescription: "The price per call in USD.",
						},
						"fee_model": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The billing model (e.g. `pay_per_call`, `subscription`).",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "A human-readable description of the resource.",
						},
						"callback_url": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The callback URL Mainlayer forwards requests to.",
						},
						"created_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "RFC3339 timestamp of when the resource was created.",
						},
						"updated_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "RFC3339 timestamp of when the resource was last updated.",
						},
					},
				},
			},
		},
	}
}

// Configure extracts the API client from the provider configuration.
func (d *ResourcesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue.", req.ProviderData),
		)
		return
	}
	d.client = c
}

// Read fetches all resources from the Mainlayer API.
func (d *ResourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ResourcesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resources, err := d.client.ListResources(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Listing Mainlayer Resources", err.Error())
		return
	}

	resourceObjects := make([]attr.Value, 0, len(resources))
	for _, r := range resources {
		obj, diags := types.ObjectValue(
			resourceItemAttrTypes,
			map[string]attr.Value{
				"id":           types.StringValue(r.ID),
				"slug":         types.StringValue(r.Slug),
				"type":         types.StringValue(r.Type),
				"price_usdc":   types.Float64Value(r.PriceUSDC),
				"fee_model":    types.StringValue(r.FeeModel),
				"description":  types.StringValue(r.Description),
				"callback_url": types.StringValue(r.CallbackURL),
				"created_at":   types.StringValue(r.CreatedAt),
				"updated_at":   types.StringValue(r.UpdatedAt),
			},
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		resourceObjects = append(resourceObjects, obj)
	}

	listVal, diags := types.ListValue(
		types.ObjectType{AttrTypes: resourceItemAttrTypes},
		resourceObjects,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue("mainlayer_resources")
	state.Resources = listVal

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
