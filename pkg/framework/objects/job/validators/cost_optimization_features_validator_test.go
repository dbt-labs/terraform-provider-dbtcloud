package job_validators_test

import (
	"context"
	"testing"

	job_validators "github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/job/validators"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func setOf(values ...string) types.Set {
	elems := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elems = append(elems, types.StringValue(v))
	}
	return types.SetValueMust(types.StringType, elems)
}

func TestCostOptimizationFeaturesValidator(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value       types.Set
		expectError bool
	}{
		"null is allowed":                 {value: types.SetNull(types.StringType), expectError: false},
		"unknown is allowed":              {value: types.SetUnknown(types.StringType), expectError: false},
		"empty set is allowed":            {value: setOf(), expectError: false},
		"state_aware_orchestration alone": {value: setOf("state_aware_orchestration"), expectError: false},
		"dbt_state alone":                 {value: setOf("dbt_state"), expectError: false},
		"dbt_state + sao is rejected":     {value: setOf("dbt_state", "state_aware_orchestration"), expectError: true},
		"efficient_testing rejected":      {value: setOf("efficient_testing"), expectError: true},
		"invalid value rejected":          {value: setOf("bogus_feature"), expectError: true},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := validator.SetRequest{
				Path:        path.Root("cost_optimization_features"),
				ConfigValue: tc.value,
			}
			resp := &validator.SetResponse{}

			job_validators.CostOptimizationFeaturesValidator().ValidateSet(context.Background(), req, resp)

			if tc.expectError && !resp.Diagnostics.HasError() {
				t.Fatalf("expected a validation error but got none")
			}
			if !tc.expectError && resp.Diagnostics.HasError() {
				t.Fatalf("expected no validation error but got: %s", resp.Diagnostics.Errors())
			}
		})
	}
}
