package account_features

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &accountFeaturesResource{}
	_ resource.ResourceWithConfigure = &accountFeaturesResource{}
)

type accountFeaturesResource struct {
	client *dbt_cloud.Client
}

func AccountFeaturesResource() resource.Resource {
	return &accountFeaturesResource{}
}

func readFeatures(client *dbt_cloud.Client) (AccountFeaturesResourceModel, error) {
	features, err := client.GetAccountFeatures()
	if err != nil {
		return AccountFeaturesResourceModel{}, err
	}

	return AccountFeaturesResourceModel{
		ID:                      types.StringValue(fmt.Sprintf("%d", client.AccountID)),
		AdvancedCI:              types.BoolValue(features.AdvancedCI),
		PartialParsing:          types.BoolValue(features.PartialParsing),
		RepoCaching:             types.BoolValue(features.RepoCaching),
		AIFeatures:              types.BoolValue(features.AIFeatures),
		WarehouseCostVisibility: types.BoolValue(features.WarehouseCostVisibility),
	}, nil
}

func (r *accountFeaturesResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_account_features"
}

func (r *accountFeaturesResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	// Read current state
	var plan AccountFeaturesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update features
	if !plan.AdvancedCI.IsUnknown() {
		err := r.client.UpdateAccountFeature("advanced-ci", plan.AdvancedCI.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError("Error updating advanced-ci feature", err.Error())
			return
		}
	}

	if !plan.PartialParsing.IsUnknown() {
		err := r.client.UpdateAccountFeature("partial-parsing", plan.PartialParsing.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError("Error updating partial-parsing feature", err.Error())
			return
		}
	}

	if !plan.RepoCaching.IsUnknown() {
		err := r.client.UpdateAccountFeature("repo-caching", plan.RepoCaching.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError("Error updating repo-caching feature", err.Error())
			return
		}
	}

	if !plan.AIFeatures.IsUnknown() {
		err := r.client.UpdateAccountFeature("ai_features", plan.AIFeatures.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError("Error updating ai_features feature", err.Error())
			return
		}
	}

	if !plan.WarehouseCostVisibility.IsUnknown() {
		err := r.client.UpdateAccountFeature("warehouse_cost_visibility", plan.WarehouseCostVisibility.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError("Error updating warehouse_cost_visibility feature", err.Error())
			return
		}
	}

	features, err := readFeatures(r.client)
	if err != nil {
		resp.Diagnostics.AddError("Error reading account features", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &features)
	resp.Diagnostics.Append(diags...)
}

func (r *accountFeaturesResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {

	features, err := readFeatures(r.client)
	if err != nil {
		resp.Diagnostics.AddError("Error reading account features", err.Error())
		return
	}

	diags := resp.State.Set(ctx, &features)
	resp.Diagnostics.Append(diags...)
}

func (r *accountFeaturesResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan AccountFeaturesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AccountFeaturesResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update changed values
	if !plan.AdvancedCI.IsUnknown() && !plan.AdvancedCI.Equal(state.AdvancedCI) {
		err := r.client.UpdateAccountFeature("advanced-ci", plan.AdvancedCI.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError("Error updating advanced-ci feature", err.Error())
			return
		}
	}

	if !plan.PartialParsing.IsUnknown() && !plan.PartialParsing.Equal(state.PartialParsing) {
		err := r.client.UpdateAccountFeature("partial-parsing", plan.PartialParsing.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError("Error updating partial-parsing feature", err.Error())
			return
		}
	}

	if !plan.RepoCaching.IsUnknown() && !plan.RepoCaching.Equal(state.RepoCaching) {
		err := r.client.UpdateAccountFeature("repo-caching", plan.RepoCaching.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError("Error updating repo-caching feature", err.Error())
			return
		}
	}

	if !plan.AIFeatures.IsUnknown() && !plan.AIFeatures.Equal(state.AIFeatures) {
		err := r.client.UpdateAccountFeature("ai_features", plan.AIFeatures.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError("Error updating ai_features feature", err.Error())
			return
		}
	}

	if !plan.WarehouseCostVisibility.IsUnknown() && !plan.WarehouseCostVisibility.Equal(state.WarehouseCostVisibility) {
		err := r.client.UpdateAccountFeature("warehouse_cost_visibility", plan.WarehouseCostVisibility.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError("Error updating warehouse_cost_visibility feature", err.Error())
			return
		}
	}

	features, err := readFeatures(r.client)
	if err != nil {
		resp.Diagnostics.AddError("Error reading account features", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &features)
	resp.Diagnostics.Append(diags...)
}

func (r *accountFeaturesResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// no-op, we keep the existing values as we technically can't "delete" the settings, just turn them on and off
}

func (r *accountFeaturesResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}
