package validators_test

import (
	"context"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/environment/validators"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUseStateWhenConfigSet_ConfigSetPlanUnknown(t *testing.T) {
	t.Parallel()

	// Config has a value, plan is unknown (Computed kicking in), state has a value.
	// Should behave like UseStateForUnknown: copy state into plan.
	req := planmodifier.Int64Request{
		ConfigValue: types.Int64Value(123),
		PlanValue:   types.Int64Unknown(),
		StateValue:  types.Int64Value(456),
	}

	resp := &planmodifier.Int64Response{
		PlanValue: req.PlanValue,
	}

	m := validators.UseStateWhenConfigSet{}
	m.PlanModifyInt64(context.Background(), req, resp)

	if resp.PlanValue.IsUnknown() {
		t.Errorf("expected plan value to use state, got unknown")
	}
	if resp.PlanValue.ValueInt64() != 456 {
		t.Errorf("expected plan value to be 456 (from state), got %v", resp.PlanValue.ValueInt64())
	}
}

func TestUseStateWhenConfigSet_ConfigSetPlanKnown(t *testing.T) {
	t.Parallel()

	// Config has a value, plan already has a known value. No modification needed.
	req := planmodifier.Int64Request{
		ConfigValue: types.Int64Value(123),
		PlanValue:   types.Int64Value(123),
		StateValue:  types.Int64Value(456),
	}

	resp := &planmodifier.Int64Response{
		PlanValue: req.PlanValue,
	}

	m := validators.UseStateWhenConfigSet{}
	m.PlanModifyInt64(context.Background(), req, resp)

	if resp.PlanValue.ValueInt64() != 123 {
		t.Errorf("expected plan value to remain 123, got %v", resp.PlanValue.ValueInt64())
	}
}

func TestUseStateWhenConfigSet_ConfigNullStateHasValue(t *testing.T) {
	t.Parallel()

	// Config is null (user removed the attribute), state has a value.
	// This is the key case: plan should become null, NOT carry forward state.
	req := planmodifier.Int64Request{
		ConfigValue: types.Int64Null(),
		PlanValue:   types.Int64Value(456), // UseStateForUnknown would have set this
		StateValue:  types.Int64Value(456),
	}

	resp := &planmodifier.Int64Response{
		PlanValue: req.PlanValue,
	}

	m := validators.UseStateWhenConfigSet{}
	m.PlanModifyInt64(context.Background(), req, resp)

	if !resp.PlanValue.IsNull() {
		t.Errorf("expected plan value to be null when config removes attribute, got %v", resp.PlanValue)
	}
}

func TestUseStateWhenConfigSet_ConfigNullStateNull(t *testing.T) {
	t.Parallel()

	// Config is null, state is null. Attribute was never set. Plan stays null.
	req := planmodifier.Int64Request{
		ConfigValue: types.Int64Null(),
		PlanValue:   types.Int64Null(),
		StateValue:  types.Int64Null(),
	}

	resp := &planmodifier.Int64Response{
		PlanValue: req.PlanValue,
	}

	m := validators.UseStateWhenConfigSet{}
	m.PlanModifyInt64(context.Background(), req, resp)

	if !resp.PlanValue.IsNull() {
		t.Errorf("expected plan value to remain null, got %v", resp.PlanValue)
	}
}

func TestUseStateWhenConfigSet_ConfigSetStateNull(t *testing.T) {
	t.Parallel()

	// Config has a value, state is null (first create). Plan is unknown (Computed).
	// State is null so we can't copy it — plan should stay unknown for the provider to fill.
	req := planmodifier.Int64Request{
		ConfigValue: types.Int64Value(123),
		PlanValue:   types.Int64Unknown(),
		StateValue:  types.Int64Null(),
	}

	resp := &planmodifier.Int64Response{
		PlanValue: req.PlanValue,
	}

	m := validators.UseStateWhenConfigSet{}
	m.PlanModifyInt64(context.Background(), req, resp)

	if !resp.PlanValue.IsUnknown() {
		t.Errorf("expected plan value to stay unknown on first create (no state to copy), got %v", resp.PlanValue)
	}
}
