package job_validators_test

import (
	"context"
	"testing"

	job_validators "github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/job/validators"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var featuresType = tftypes.Set{ElementType: tftypes.String}

var planModifierSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"force_node_selection":       schema.BoolAttribute{},
		"cost_optimization_features": schema.SetAttribute{ElementType: types.StringType},
	},
}

var objectType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"force_node_selection":       tftypes.Bool,
		"cost_optimization_features": featuresType,
	},
}

// featureSet builds a tftypes value for cost_optimization_features. A nil slice
// is the attribute being absent.
func featureSet(features []string) tftypes.Value {
	if features == nil {
		return tftypes.NewValue(featuresType, nil)
	}
	elements := make([]tftypes.Value, 0, len(features))
	for _, feature := range features {
		elements = append(elements, tftypes.NewValue(tftypes.String, feature))
	}
	return tftypes.NewValue(featuresType, elements)
}

func object(forceNodeSelection *bool, features []string) tftypes.Value {
	var fns tftypes.Value
	if forceNodeSelection == nil {
		fns = tftypes.NewValue(tftypes.Bool, nil)
	} else {
		fns = tftypes.NewValue(tftypes.Bool, *forceNodeSelection)
	}
	return tftypes.NewValue(objectType, map[string]tftypes.Value{
		"force_node_selection":       fns,
		"cost_optimization_features": featureSet(features),
	})
}

func boolPtr(value bool) *bool { return &value }

// runModifier drives the plan modifier. A nil stateObject stands for a create,
// where there is no prior state at all.
func runModifier(
	t *testing.T,
	configValue types.Bool,
	stateValue types.Bool,
	configObject tftypes.Value,
	stateObject *tftypes.Value,
) planmodifier.BoolResponse {
	t.Helper()

	state := tfsdk.State{Schema: planModifierSchema, Raw: tftypes.NewValue(objectType, nil)}
	if stateObject != nil {
		state = tfsdk.State{Schema: planModifierSchema, Raw: *stateObject}
	}

	req := planmodifier.BoolRequest{
		ConfigValue: configValue,
		PlanValue:   types.BoolUnknown(),
		StateValue:  stateValue,
		Config:      tfsdk.Config{Schema: planModifierSchema, Raw: configObject},
		State:       state,
	}
	resp := planmodifier.BoolResponse{PlanValue: req.PlanValue}

	job_validators.UseUnknownWhenCostOptimizationFeaturesChange{}.PlanModifyBool(
		context.Background(), req, &resp,
	)
	return resp
}

// TestPlanUnknownWhenFeaturesAdded is the regression for issue #751. A job with
// no cost optimization features has force_node_selection true. Adding a feature
// makes the API set it to false, so holding true in the plan breaks the apply.
func TestPlanUnknownWhenFeaturesAdded(t *testing.T) {
	t.Parallel()

	stateObject := object(boolPtr(true), nil)
	resp := runModifier(
		t,
		types.BoolNull(),
		types.BoolValue(true),
		object(nil, []string{"dbt_state"}),
		&stateObject,
	)

	if !resp.PlanValue.IsUnknown() {
		t.Errorf("expected the plan value to be unknown, got %v", resp.PlanValue)
	}
}

// TestPlanUnknownWhenFeaturesRemoved covers the other direction, where the API
// sets force_node_selection back to true.
func TestPlanUnknownWhenFeaturesRemoved(t *testing.T) {
	t.Parallel()

	stateObject := object(boolPtr(false), []string{"dbt_state"})
	resp := runModifier(
		t,
		types.BoolNull(),
		types.BoolValue(false),
		object(nil, []string{}),
		&stateObject,
	)

	if !resp.PlanValue.IsUnknown() {
		t.Errorf("expected the plan value to be unknown, got %v", resp.PlanValue)
	}
}

// TestPlanKeepsStateWhenFeaturesStable is the behaviour UseStateForUnknown gave
// before. Without it an unrelated change would show force_node_selection as
// "known after apply" on every plan.
func TestPlanKeepsStateWhenFeaturesStable(t *testing.T) {
	t.Parallel()

	stateObject := object(boolPtr(false), []string{"dbt_state"})
	resp := runModifier(
		t,
		types.BoolNull(),
		types.BoolValue(false),
		object(nil, []string{"dbt_state"}),
		&stateObject,
	)

	if resp.PlanValue.IsUnknown() || resp.PlanValue.ValueBool() {
		t.Errorf("expected the plan to keep false from state, got %v", resp.PlanValue)
	}
}

// TestPlanKeepsStateWhenFeaturesAbsentFromConfig covers a configuration that
// never mentions cost_optimization_features. The feature set is not changing,
// so the value has to settle rather than read as unknown on every plan.
func TestPlanKeepsStateWhenFeaturesAbsentFromConfig(t *testing.T) {
	t.Parallel()

	stateObject := object(boolPtr(true), nil)
	resp := runModifier(
		t,
		types.BoolNull(),
		types.BoolValue(true),
		object(nil, nil),
		&stateObject,
	)

	if resp.PlanValue.IsUnknown() || !resp.PlanValue.ValueBool() {
		t.Errorf("expected the plan to keep true from state, got %v", resp.PlanValue)
	}
}

// TestPlanKeepsConfiguredValue checks that a value the configuration sets is
// never replaced, whatever happens to the feature set.
func TestPlanKeepsConfiguredValue(t *testing.T) {
	t.Parallel()

	stateObject := object(boolPtr(true), nil)
	req := planmodifier.BoolRequest{
		ConfigValue: types.BoolValue(false),
		PlanValue:   types.BoolValue(false),
		StateValue:  types.BoolValue(true),
		Config: tfsdk.Config{
			Schema: planModifierSchema,
			Raw:    object(boolPtr(false), []string{"dbt_state"}),
		},
		State: tfsdk.State{Schema: planModifierSchema, Raw: stateObject},
	}
	resp := planmodifier.BoolResponse{PlanValue: req.PlanValue}

	job_validators.UseUnknownWhenCostOptimizationFeaturesChange{}.PlanModifyBool(
		context.Background(), req, &resp,
	)

	if resp.PlanValue.IsUnknown() || resp.PlanValue.ValueBool() {
		t.Errorf("expected the configured false to survive, got %v", resp.PlanValue)
	}
}

// TestPlanLeavesUnknownOnCreate checks that a create keeps the unknown it
// starts with, because there is no prior value to hold on to.
func TestPlanLeavesUnknownOnCreate(t *testing.T) {
	t.Parallel()

	resp := runModifier(
		t,
		types.BoolNull(),
		types.BoolNull(),
		object(nil, []string{"dbt_state"}),
		nil,
	)

	if !resp.PlanValue.IsUnknown() {
		t.Errorf("expected the plan value to stay unknown, got %v", resp.PlanValue)
	}
}
