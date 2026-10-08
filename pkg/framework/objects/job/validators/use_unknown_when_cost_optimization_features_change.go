package job_validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UseUnknownWhenCostOptimizationFeaturesChange is a plan modifier for
// force_node_selection. The dbt platform API keeps the two fields consistent and
// gives cost_optimization_features precedence, so adding a feature to a job that
// had none rewrites force_node_selection to false. Holding the old value in the
// plan makes that apply fail with "Provider produced inconsistent result after
// apply", so the value is unknown whenever the feature set changes.
//
// When the feature set is stable, the value in state is kept, which is what
// UseStateForUnknown did before, so a plan with no change to the features still
// settles. A value the configuration sets is never touched.
type UseUnknownWhenCostOptimizationFeaturesChange struct{}

func (m UseUnknownWhenCostOptimizationFeaturesChange) Description(_ context.Context) string {
	return "Sets the value to unknown when cost_optimization_features changes, since the API derives force_node_selection from it."
}

func (m UseUnknownWhenCostOptimizationFeaturesChange) MarkdownDescription(_ context.Context) string {
	return "Sets the value to unknown when `cost_optimization_features` changes, since the API derives `force_node_selection` from it."
}

func (m UseUnknownWhenCostOptimizationFeaturesChange) PlanModifyBool(
	ctx context.Context,
	req planmodifier.BoolRequest,
	resp *planmodifier.BoolResponse,
) {
	// The configuration sets the value, so the plan already holds it.
	if !req.ConfigValue.IsNull() {
		return
	}

	// There is no prior value to keep while the job is being created, and the
	// API decides the value.
	if req.State.Raw.IsNull() {
		return
	}

	var configFeatures types.Set
	if diags := req.Config.GetAttribute(
		ctx,
		path.Root("cost_optimization_features"),
		&configFeatures,
	); diags.HasError() {
		return
	}

	var stateFeatures types.Set
	if diags := req.State.GetAttribute(
		ctx,
		path.Root("cost_optimization_features"),
		&stateFeatures,
	); diags.HasError() {
		return
	}

	// The configuration does not set the feature set, so it is not changing it.
	if configFeatures.IsNull() {
		resp.PlanValue = req.StateValue
		return
	}

	// The feature set is known and the same as before, so force_node_selection
	// keeps the value it has.
	if !configFeatures.IsUnknown() && configFeatures.Equal(stateFeatures) {
		resp.PlanValue = req.StateValue
		return
	}

	// The feature set is changing, or is not known yet. Either way the API
	// decides what force_node_selection becomes.
	resp.PlanValue = types.BoolUnknown()
}
