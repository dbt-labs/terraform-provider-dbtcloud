package validators_test

import (
	"context"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/environment/validators"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUseUnknownWhenProfileChanges_ProfileNewlySet(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"primary_profile_id": schema.Int64Attribute{},
			"connection_id":      schema.Int64Attribute{},
		},
	}

	config := tfsdk.Config{
		Schema: testSchema,
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"primary_profile_id": tftypes.Number,
					"connection_id":      tftypes.Number,
				},
			},
			map[string]tftypes.Value{
				"primary_profile_id": tftypes.NewValue(tftypes.Number, 123),
				"connection_id":      tftypes.NewValue(tftypes.Number, 0),
			},
		),
	}

	// State has no profile (null) — this is a new profile being set
	state := tfsdk.State{
		Schema: testSchema,
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"primary_profile_id": tftypes.Number,
					"connection_id":      tftypes.Number,
				},
			},
			map[string]tftypes.Value{
				"primary_profile_id": tftypes.NewValue(tftypes.Number, nil),
				"connection_id":      tftypes.NewValue(tftypes.Number, 456),
			},
		),
	}

	req := planmodifier.Int64Request{
		ConfigValue: types.Int64Value(0),
		PlanValue:   types.Int64Value(0),
		StateValue:  types.Int64Value(456),
		Config:      config,
		State:       state,
	}

	resp := &planmodifier.Int64Response{
		PlanValue: req.PlanValue,
	}

	m := validators.UseUnknownWhenProfileChanges{}
	m.PlanModifyInt64(context.Background(), req, resp)

	if !resp.PlanValue.IsUnknown() {
		t.Errorf("expected plan value to be unknown when profile is newly set, got %v", resp.PlanValue)
	}
}

func TestUseUnknownWhenProfileChanges_ProfileChanged(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"primary_profile_id": schema.Int64Attribute{},
			"connection_id":      schema.Int64Attribute{},
		},
	}

	config := tfsdk.Config{
		Schema: testSchema,
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"primary_profile_id": tftypes.Number,
					"connection_id":      tftypes.Number,
				},
			},
			map[string]tftypes.Value{
				"primary_profile_id": tftypes.NewValue(tftypes.Number, 999),
				"connection_id":      tftypes.NewValue(tftypes.Number, 0),
			},
		),
	}

	// State has a different profile — profile is changing
	state := tfsdk.State{
		Schema: testSchema,
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"primary_profile_id": tftypes.Number,
					"connection_id":      tftypes.Number,
				},
			},
			map[string]tftypes.Value{
				"primary_profile_id": tftypes.NewValue(tftypes.Number, 123),
				"connection_id":      tftypes.NewValue(tftypes.Number, 456),
			},
		),
	}

	req := planmodifier.Int64Request{
		ConfigValue: types.Int64Value(0),
		PlanValue:   types.Int64Value(0),
		StateValue:  types.Int64Value(456),
		Config:      config,
		State:       state,
	}

	resp := &planmodifier.Int64Response{
		PlanValue: req.PlanValue,
	}

	m := validators.UseUnknownWhenProfileChanges{}
	m.PlanModifyInt64(context.Background(), req, resp)

	if !resp.PlanValue.IsUnknown() {
		t.Errorf("expected plan value to be unknown when profile changed, got %v", resp.PlanValue)
	}
}

func TestUseUnknownWhenProfileChanges_ProfileStable(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"primary_profile_id": schema.Int64Attribute{},
			"connection_id":      schema.Int64Attribute{},
		},
	}

	config := tfsdk.Config{
		Schema: testSchema,
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"primary_profile_id": tftypes.Number,
					"connection_id":      tftypes.Number,
				},
			},
			map[string]tftypes.Value{
				"primary_profile_id": tftypes.NewValue(tftypes.Number, 123),
				"connection_id":      tftypes.NewValue(tftypes.Number, 0),
			},
		),
	}

	// State has the same profile — nothing changed
	state := tfsdk.State{
		Schema: testSchema,
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"primary_profile_id": tftypes.Number,
					"connection_id":      tftypes.Number,
				},
			},
			map[string]tftypes.Value{
				"primary_profile_id": tftypes.NewValue(tftypes.Number, 123),
				"connection_id":      tftypes.NewValue(tftypes.Number, 789),
			},
		),
	}

	req := planmodifier.Int64Request{
		ConfigValue: types.Int64Value(0),
		PlanValue:   types.Int64Value(0),
		StateValue:  types.Int64Value(789),
		Config:      config,
		State:       state,
	}

	resp := &planmodifier.Int64Response{
		PlanValue: req.PlanValue,
	}

	m := validators.UseUnknownWhenProfileChanges{}
	m.PlanModifyInt64(context.Background(), req, resp)

	if resp.PlanValue.IsUnknown() {
		t.Errorf("expected plan value to preserve state when profile is stable, got unknown")
	}
	if resp.PlanValue.ValueInt64() != 789 {
		t.Errorf("expected plan value to be 789 (from state), got %v", resp.PlanValue.ValueInt64())
	}
}

func TestUseUnknownWhenProfileChanges_ProfileUnknown(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"primary_profile_id": schema.Int64Attribute{},
			"connection_id":      schema.Int64Attribute{},
		},
	}

	config := tfsdk.Config{
		Schema: testSchema,
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"primary_profile_id": tftypes.Number,
					"connection_id":      tftypes.Number,
				},
			},
			map[string]tftypes.Value{
				"primary_profile_id": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				"connection_id":      tftypes.NewValue(tftypes.Number, 0),
			},
		),
	}

	// No prior state — this is a create where the profile reference isn't resolved yet
	state := tfsdk.State{
		Schema: testSchema,
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"primary_profile_id": tftypes.Number,
					"connection_id":      tftypes.Number,
				},
			},
			map[string]tftypes.Value{
				"primary_profile_id": tftypes.NewValue(tftypes.Number, nil),
				"connection_id":      tftypes.NewValue(tftypes.Number, nil),
			},
		),
	}

	req := planmodifier.Int64Request{
		ConfigValue: types.Int64Value(0),
		PlanValue:   types.Int64Value(0),
		StateValue:  types.Int64Null(),
		Config:      config,
		State:       state,
	}

	resp := &planmodifier.Int64Response{
		PlanValue: req.PlanValue,
	}

	m := validators.UseUnknownWhenProfileChanges{}
	m.PlanModifyInt64(context.Background(), req, resp)

	if !resp.PlanValue.IsUnknown() {
		t.Errorf("expected plan value to be unknown when profile is unknown, got %v", resp.PlanValue)
	}
}

func TestUseUnknownWhenProfileChanges_ProfileNull(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"primary_profile_id": schema.Int64Attribute{},
			"connection_id":      schema.Int64Attribute{},
		},
	}

	config := tfsdk.Config{
		Schema: testSchema,
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"primary_profile_id": tftypes.Number,
					"connection_id":      tftypes.Number,
				},
			},
			map[string]tftypes.Value{
				"primary_profile_id": tftypes.NewValue(tftypes.Number, nil),
				"connection_id":      tftypes.NewValue(tftypes.Number, 456),
			},
		),
	}

	req := planmodifier.Int64Request{
		ConfigValue: types.Int64Value(456),
		PlanValue:   types.Int64Value(456),
		StateValue:  types.Int64Value(456),
		Config:      config,
	}

	resp := &planmodifier.Int64Response{
		PlanValue: req.PlanValue,
	}

	m := validators.UseUnknownWhenProfileChanges{}
	m.PlanModifyInt64(context.Background(), req, resp)

	if resp.PlanValue.IsUnknown() {
		t.Errorf("expected plan value to remain unchanged when profile is null, got unknown")
	}
	if resp.PlanValue.ValueInt64() != 456 {
		t.Errorf("expected plan value to be 456, got %v", resp.PlanValue.ValueInt64())
	}
}
