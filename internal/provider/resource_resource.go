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

// Ensure ResourceResource satisfies the resource.Resource interface.
var _ resource.Resource = &ResourceResource{}
var _ resource.ResourceWithImportState = &ResourceResource{}

// ResourceResource manages a mainlayer_resource Terraform resource.
type ResourceResource struct {
	client *client.Client
}

// ResourceResourceModel maps the Terraform schema to Go fields.
type ResourceResourceModel struct {
	ID          types.String  `tfsdk:"id"`
	Slug        types.String  `tfsdk:"slug"`
	Type        types.String  `tfsdk:"type"`
	PriceUSDC   types.Float64 `tfsdk:"price_usdc"`
	FeeModel    types.String  `tfsdk:"fee_model"`
	Description types.String  `tfsdk:"description"`
	CallbackURL types.String  `tfsdk:"callback_url"`
	CreatedAt   types.String  `tfsdk:"created_at"`
	UpdatedAt   types.String  `tfsdk:"updated_at"`
}

// NewResourceResource is the factory function for mainlayer_resource.
func NewResourceResource() resource.Resource {
	return &ResourceResource{}
}

// Metadata sets the resource type name.
func (r *ResourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource"
}

// Schema defines the Terraform schema for mainlayer_resource.
func (r *ResourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Mainlayer resource. A resource represents a monetised API endpoint, " +
			"tool, or service that agents and developers can call through the Mainlayer platform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the resource, assigned by Mainlayer.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "A URL-safe identifier for the resource (e.g. `my-api`). Must be unique within your account.",
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The resource type. Common values: `api`, `tool`, `model`, `dataset`.",
			},
			"price_usdc": schema.Float64Attribute{
				Required:            true,
				MarkdownDescription: "The price per call in USD (billed in real time). For example, `0.01` = $0.01 per call.",
			},
			"fee_model": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The billing model. Supported values: `pay_per_call`, `subscription`, `free`.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "A human-readable description of what the resource does.",
			},
			"callback_url": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The HTTPS URL that Mainlayer will forward requests to after processing payment.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "RFC3339 timestamp of when the resource was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "RFC3339 timestamp of when the resource was last updated.",
			},
		},
	}
}

// Configure extracts the API client from the provider configuration.
func (r *ResourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new mainlayer_resource via the API.
func (r *ResourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResource := &client.Resource{
		Slug:        plan.Slug.ValueString(),
		Type:        plan.Type.ValueString(),
		PriceUSDC:   plan.PriceUSDC.ValueFloat64(),
		FeeModel:    plan.FeeModel.ValueString(),
		Description: plan.Description.ValueString(),
		CallbackURL: plan.CallbackURL.ValueString(),
	}

	created, err := r.client.CreateResource(ctx, apiResource)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Mainlayer Resource", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.Slug = types.StringValue(created.Slug)
	plan.Type = types.StringValue(created.Type)
	plan.PriceUSDC = types.Float64Value(created.PriceUSDC)
	plan.FeeModel = types.StringValue(created.FeeModel)
	plan.Description = types.StringValue(created.Description)
	plan.CallbackURL = types.StringValue(created.CallbackURL)
	plan.CreatedAt = types.StringValue(created.CreatedAt)
	plan.UpdatedAt = types.StringValue(created.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest API data.
func (r *ResourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResource, err := r.client.GetResource(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Mainlayer Resource", err.Error())
		return
	}
	if apiResource == nil {
		// Resource has been deleted outside of Terraform.
		resp.State.RemoveResource(ctx)
		return
	}

	state.Slug = types.StringValue(apiResource.Slug)
	state.Type = types.StringValue(apiResource.Type)
	state.PriceUSDC = types.Float64Value(apiResource.PriceUSDC)
	state.FeeModel = types.StringValue(apiResource.FeeModel)
	state.Description = types.StringValue(apiResource.Description)
	state.CallbackURL = types.StringValue(apiResource.CallbackURL)
	state.CreatedAt = types.StringValue(apiResource.CreatedAt)
	state.UpdatedAt = types.StringValue(apiResource.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates an existing mainlayer_resource via the API.
func (r *ResourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ResourceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ResourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResource := &client.Resource{
		Slug:        plan.Slug.ValueString(),
		Type:        plan.Type.ValueString(),
		PriceUSDC:   plan.PriceUSDC.ValueFloat64(),
		FeeModel:    plan.FeeModel.ValueString(),
		Description: plan.Description.ValueString(),
		CallbackURL: plan.CallbackURL.ValueString(),
	}

	updated, err := r.client.UpdateResource(ctx, state.ID.ValueString(), apiResource)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Mainlayer Resource", err.Error())
		return
	}

	plan.ID = state.ID
	plan.Slug = types.StringValue(updated.Slug)
	plan.Type = types.StringValue(updated.Type)
	plan.PriceUSDC = types.Float64Value(updated.PriceUSDC)
	plan.FeeModel = types.StringValue(updated.FeeModel)
	plan.Description = types.StringValue(updated.Description)
	plan.CallbackURL = types.StringValue(updated.CallbackURL)
	plan.CreatedAt = types.StringValue(updated.CreatedAt)
	plan.UpdatedAt = types.StringValue(updated.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a mainlayer_resource via the API.
func (r *ResourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteResource(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Mainlayer Resource", err.Error())
	}
}

// ImportState enables `terraform import mainlayer_resource.example <id>`.
func (r *ResourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	apiResource, err := r.client.GetResource(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Mainlayer Resource", err.Error())
		return
	}
	if apiResource == nil {
		resp.Diagnostics.AddError("Resource Not Found", fmt.Sprintf("No resource with ID %q was found in Mainlayer.", req.ID))
		return
	}

	state := ResourceResourceModel{
		ID:          types.StringValue(apiResource.ID),
		Slug:        types.StringValue(apiResource.Slug),
		Type:        types.StringValue(apiResource.Type),
		PriceUSDC:   types.Float64Value(apiResource.PriceUSDC),
		FeeModel:    types.StringValue(apiResource.FeeModel),
		Description: types.StringValue(apiResource.Description),
		CallbackURL: types.StringValue(apiResource.CallbackURL),
		CreatedAt:   types.StringValue(apiResource.CreatedAt),
		UpdatedAt:   types.StringValue(apiResource.UpdatedAt),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
