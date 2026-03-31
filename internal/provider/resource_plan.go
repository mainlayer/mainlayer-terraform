package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mainlayer/terraform-provider-mainlayer/internal/client"
)

// Ensure PlanResource satisfies the resource.Resource interface.
var _ resource.Resource = &PlanResource{}
var _ resource.ResourceWithImportState = &PlanResource{}

// PlanResource manages a mainlayer_plan Terraform resource.
type PlanResource struct {
	client *client.Client
}

// PlanResourceModel maps the Terraform schema to Go fields.
type PlanResourceModel struct {
	ID          types.String  `tfsdk:"id"`
	ResourceID  types.String  `tfsdk:"resource_id"`
	Name        types.String  `tfsdk:"name"`
	Description types.String  `tfsdk:"description"`
	PriceUSDC   types.Float64 `tfsdk:"price_usdc"`
	CallLimit   types.Int64   `tfsdk:"call_limit"`
	Period      types.String  `tfsdk:"period"`
	CreatedAt   types.String  `tfsdk:"created_at"`
	UpdatedAt   types.String  `tfsdk:"updated_at"`
}

// NewPlanResource is the factory function for mainlayer_plan.
func NewPlanResource() resource.Resource {
	return &PlanResource{}
}

// Metadata sets the resource type name.
func (r *PlanResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plan"
}

// Schema defines the Terraform schema for mainlayer_plan.
func (r *PlanResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Mainlayer plan. Plans define subscription tiers for a resource, " +
			"allowing you to offer different usage levels and pricing to consumers.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the plan, assigned by Mainlayer.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the `mainlayer_resource` this plan belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The display name of the plan (e.g. `Starter`, `Pro`, `Enterprise`).",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "A human-readable description of what the plan includes.",
			},
			"price_usdc": schema.Float64Attribute{
				Required:            true,
				MarkdownDescription: "The recurring price for this plan in USD per billing period.",
			},
			"call_limit": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The maximum number of API calls allowed per billing period. `0` means unlimited.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"period": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The billing period. Supported values: `monthly`, `yearly`. Defaults to `monthly`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "RFC3339 timestamp of when the plan was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "RFC3339 timestamp of when the plan was last updated.",
			},
		},
	}
}

// Configure extracts the API client from the provider configuration.
func (r *PlanResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new mainlayer_plan via the API.
func (r *PlanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PlanResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPlan := &client.Plan{
		ResourceID:  plan.ResourceID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		PriceUSDC:   plan.PriceUSDC.ValueFloat64(),
		CallLimit:   plan.CallLimit.ValueInt64(),
		Period:      plan.Period.ValueString(),
	}

	created, err := r.client.CreatePlan(ctx, apiPlan)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Mainlayer Plan", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.ResourceID = types.StringValue(created.ResourceID)
	plan.Name = types.StringValue(created.Name)
	plan.Description = types.StringValue(created.Description)
	plan.PriceUSDC = types.Float64Value(created.PriceUSDC)
	plan.CallLimit = types.Int64Value(created.CallLimit)
	plan.Period = types.StringValue(created.Period)
	plan.CreatedAt = types.StringValue(created.CreatedAt)
	plan.UpdatedAt = types.StringValue(created.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest plan data.
func (r *PlanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PlanResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPlan, err := r.client.GetPlan(ctx, state.ResourceID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Mainlayer Plan", err.Error())
		return
	}
	if apiPlan == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(apiPlan.Name)
	state.Description = types.StringValue(apiPlan.Description)
	state.PriceUSDC = types.Float64Value(apiPlan.PriceUSDC)
	state.CallLimit = types.Int64Value(apiPlan.CallLimit)
	state.Period = types.StringValue(apiPlan.Period)
	state.CreatedAt = types.StringValue(apiPlan.CreatedAt)
	state.UpdatedAt = types.StringValue(apiPlan.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates an existing mainlayer_plan via the API.
func (r *PlanResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PlanResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PlanResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPlan := &client.Plan{
		ResourceID:  state.ResourceID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		PriceUSDC:   plan.PriceUSDC.ValueFloat64(),
		CallLimit:   plan.CallLimit.ValueInt64(),
		Period:      plan.Period.ValueString(),
	}

	updated, err := r.client.UpdatePlan(ctx, state.ResourceID.ValueString(), state.ID.ValueString(), apiPlan)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Mainlayer Plan", err.Error())
		return
	}

	plan.ID = state.ID
	plan.ResourceID = state.ResourceID
	plan.Name = types.StringValue(updated.Name)
	plan.Description = types.StringValue(updated.Description)
	plan.PriceUSDC = types.Float64Value(updated.PriceUSDC)
	plan.CallLimit = types.Int64Value(updated.CallLimit)
	plan.Period = types.StringValue(updated.Period)
	plan.CreatedAt = types.StringValue(updated.CreatedAt)
	plan.UpdatedAt = types.StringValue(updated.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a mainlayer_plan via the API.
func (r *PlanResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PlanResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeletePlan(ctx, state.ResourceID.ValueString(), state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Mainlayer Plan", err.Error())
	}
}

// ImportState enables `terraform import mainlayer_plan.example <resource_id>/<plan_id>`.
func (r *PlanResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expected format: <resource_id>/<plan_id>
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in the format `<resource_id>/<plan_id>`, e.g. `res_abc123/plan_xyz456`.",
		)
		return
	}
	resourceID := parts[0]
	planID := parts[1]

	apiPlan, err := r.client.GetPlan(ctx, resourceID, planID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Mainlayer Plan", err.Error())
		return
	}
	if apiPlan == nil {
		resp.Diagnostics.AddError("Plan Not Found", fmt.Sprintf("No plan with ID %q under resource %q was found.", planID, resourceID))
		return
	}

	state := PlanResourceModel{
		ID:          types.StringValue(apiPlan.ID),
		ResourceID:  types.StringValue(resourceID),
		Name:        types.StringValue(apiPlan.Name),
		Description: types.StringValue(apiPlan.Description),
		PriceUSDC:   types.Float64Value(apiPlan.PriceUSDC),
		CallLimit:   types.Int64Value(apiPlan.CallLimit),
		Period:      types.StringValue(apiPlan.Period),
		CreatedAt:   types.StringValue(apiPlan.CreatedAt),
		UpdatedAt:   types.StringValue(apiPlan.UpdatedAt),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
