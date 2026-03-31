package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mainlayer/terraform-provider-mainlayer/internal/client"
)

// Ensure VendorResource satisfies the resource.Resource interface.
var _ resource.Resource = &VendorResource{}
var _ resource.ResourceWithImportState = &VendorResource{}

// VendorResource manages a mainlayer_vendor Terraform resource.
type VendorResource struct {
	client *client.Client
}

// VendorResourceModel maps the Terraform schema to Go fields.
type VendorResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Email       types.String `tfsdk:"email"`
	Website     types.String `tfsdk:"website"`
	Country     types.String `tfsdk:"country"`
	Description types.String `tfsdk:"description"`
	APIKey      types.String `tfsdk:"api_key"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// NewVendorResource is the factory function for mainlayer_vendor.
func NewVendorResource() resource.Resource {
	return &VendorResource{}
}

// Metadata sets the resource type name.
func (r *VendorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vendor"
}

// Schema defines the Terraform schema for mainlayer_vendor.
func (r *VendorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Mainlayer vendor account. Vendors can create and monetize resources " +
			"(APIs, tools, datasets, agents) on the Mainlayer platform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique vendor ID assigned by Mainlayer, prefixed with `vnd_`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the vendor or company.",
			},
			"email": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The primary email address for vendor communications and invoices.",
			},
			"website": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The vendor's website URL.",
			},
			"country": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "ISO 3166-1 alpha-2 country code (e.g., `US`, `FR`, `GB`).",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "A brief description of the vendor's products or services.",
			},
			"api_key": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The initial API key generated for this vendor. Store this securely.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "RFC3339 timestamp of when the vendor was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "RFC3339 timestamp of when the vendor was last updated.",
			},
		},
	}
}

// Configure extracts the API client from the provider configuration.
func (r *VendorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue.", req.ProviderData),
		)
		return
	}
	r.client = c
}

// Create creates a new mainlayer_vendor via the API.
func (r *VendorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VendorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiVendor := &client.Vendor{
		Name:        plan.Name.ValueString(),
		Email:       plan.Email.ValueString(),
		Website:     plan.Website.ValueString(),
		Country:     plan.Country.ValueString(),
		Description: plan.Description.ValueString(),
	}

	created, err := r.client.CreateVendor(ctx, apiVendor)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Mainlayer Vendor", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.Name = types.StringValue(created.Name)
	plan.Email = types.StringValue(created.Email)
	plan.Website = types.StringValue(created.Website)
	plan.Country = types.StringValue(created.Country)
	plan.Description = types.StringValue(created.Description)
	plan.APIKey = types.StringValue(created.APIKey)
	plan.CreatedAt = types.StringValue(created.CreatedAt)
	plan.UpdatedAt = types.StringValue(created.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest API data.
func (r *VendorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VendorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiVendor, err := r.client.GetVendor(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Mainlayer Vendor", err.Error())
		return
	}
	if apiVendor == nil {
		// Vendor has been deleted outside of Terraform.
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(apiVendor.Name)
	state.Email = types.StringValue(apiVendor.Email)
	state.Website = types.StringValue(apiVendor.Website)
	state.Country = types.StringValue(apiVendor.Country)
	state.Description = types.StringValue(apiVendor.Description)
	state.CreatedAt = types.StringValue(apiVendor.CreatedAt)
	state.UpdatedAt = types.StringValue(apiVendor.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates an existing mainlayer_vendor via the API.
func (r *VendorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VendorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state VendorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiVendor := &client.Vendor{
		Name:        plan.Name.ValueString(),
		Email:       plan.Email.ValueString(),
		Website:     plan.Website.ValueString(),
		Country:     plan.Country.ValueString(),
		Description: plan.Description.ValueString(),
	}

	updated, err := r.client.UpdateVendor(ctx, state.ID.ValueString(), apiVendor)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Mainlayer Vendor", err.Error())
		return
	}

	plan.ID = state.ID
	plan.Name = types.StringValue(updated.Name)
	plan.Email = types.StringValue(updated.Email)
	plan.Website = types.StringValue(updated.Website)
	plan.Country = types.StringValue(updated.Country)
	plan.Description = types.StringValue(updated.Description)
	plan.CreatedAt = types.StringValue(updated.CreatedAt)
	plan.UpdatedAt = types.StringValue(updated.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a mainlayer_vendor via the API.
func (r *VendorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VendorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteVendor(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Mainlayer Vendor", err.Error())
	}
}

// ImportState enables `terraform import mainlayer_vendor.example <id>`.
func (r *VendorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	apiVendor, err := r.client.GetVendor(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Mainlayer Vendor", err.Error())
		return
	}
	if apiVendor == nil {
		resp.Diagnostics.AddError("Vendor Not Found", fmt.Sprintf("No vendor with ID %q was found in Mainlayer.", req.ID))
		return
	}

	state := VendorResourceModel{
		ID:          types.StringValue(apiVendor.ID),
		Name:        types.StringValue(apiVendor.Name),
		Email:       types.StringValue(apiVendor.Email),
		Website:     types.StringValue(apiVendor.Website),
		Country:     types.StringValue(apiVendor.Country),
		Description: types.StringValue(apiVendor.Description),
		APIKey:      types.StringValue(apiVendor.APIKey),
		CreatedAt:   types.StringValue(apiVendor.CreatedAt),
		UpdatedAt:   types.StringValue(apiVendor.UpdatedAt),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
