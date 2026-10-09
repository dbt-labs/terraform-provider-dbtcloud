package postgres_credential

import (
	"context"
	"testing"

	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestIDAttribute_UseStateForUnknown is a regression test for #719: the "id"
// attribute had no PlanModifiers, so it was always recomputed as unknown
// whenever the resource was touched for an update (e.g. one triggered by
// bumping password_wo_version), even though id is deterministically derived
// from project_id and credential_id and never actually changes. That surfaced
// as a spurious "id = ... -> (known after apply)" in-place update on every
// terraform plan.
func TestIDAttribute_UseStateForUnknown(t *testing.T) {
	attr, ok := PostgresResourceSchema.Attributes["id"].(resource_schema.StringAttribute)
	if !ok {
		t.Fatalf("id attribute is not a resource_schema.StringAttribute")
	}
	if len(attr.PlanModifiers) == 0 {
		t.Fatalf("regression: id attribute has no plan modifiers, so terraform will always recompute it as unknown")
	}

	// Build a real prior state matching the full resource schema: the plan
	// modifier bails out early if req.State.Raw is null, so a bare
	// StringRequest{StateValue: ...} isn't enough to exercise it faithfully.
	priorState := &PostgresCredentialResourceModel{
		ID:                      types.StringValue("5:10"),
		ProjectID:               types.Int64Value(5),
		CredentialID:            types.Int64Value(10),
		IsActive:                types.BoolValue(true),
		DefaultSchema:           types.StringValue("default_schema"),
		Username:                types.StringValue("user"),
		NumThreads:              types.Int64Value(0),
		Type:                    types.StringValue("postgres"),
		TargetName:              types.StringValue("default"),
		Password:                types.StringNull(),
		PasswordWo:              types.StringNull(),
		PasswordWoVersion:       types.Int64Value(1),
		SemanticLayerCredential: types.BoolValue(false),
	}

	tfState := tfsdk.State{Schema: PostgresResourceSchema}
	if diags := tfState.Set(context.Background(), priorState); diags.HasError() {
		t.Fatalf("failed to build state fixture: %v", diags)
	}

	req := planmodifier.StringRequest{
		State:       tfState,
		StateValue:  types.StringValue("5:10"),
		PlanValue:   types.StringUnknown(),
		ConfigValue: types.StringNull(),
	}

	for _, pm := range attr.PlanModifiers {
		resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
		pm.PlanModifyString(context.Background(), req, resp)
		req.PlanValue = resp.PlanValue
	}

	if req.PlanValue.IsUnknown() {
		t.Fatalf("regression: id is still unknown after plan modification; terraform plan would show a spurious in-place update")
	}
	if req.PlanValue.ValueString() != "5:10" {
		t.Fatalf("expected id to retain the prior state value %q, got %q", "5:10", req.PlanValue.ValueString())
	}
}
