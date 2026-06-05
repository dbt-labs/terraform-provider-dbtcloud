package job_validators

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Valid cost_optimization_features values. The dbt Cloud API enum
// (sinter/common/constants/jobs.py::CostOptimizationFeature) also defines
// efficient_testing, but it is intentionally not exposed by the provider.
const (
	CostOptimizationFeatureStateAwareOrchestration = "state_aware_orchestration"
	CostOptimizationFeatureDbtState                = "dbt_state"
)

var validCostOptimizationFeatures = []string{
	CostOptimizationFeatureStateAwareOrchestration,
	CostOptimizationFeatureDbtState,
}

var _ validator.Set = &costOptimizationFeaturesValidator{}

type costOptimizationFeaturesValidator struct{}

func (v costOptimizationFeaturesValidator) Description(ctx context.Context) string {
	return fmt.Sprintf(
		"each value must be one of %s; when dbt_state is present it must be the only feature",
		strings.Join(validCostOptimizationFeatures, ", "),
	)
}

func (v costOptimizationFeaturesValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf(
		"Each value must be one of `%s`. When `dbt_state` is present it must be the only feature.",
		strings.Join(validCostOptimizationFeatures, "`, `"),
	)
}

func (v costOptimizationFeaturesValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var features []string
	resp.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &features, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	valid := make(map[string]struct{}, len(validCostOptimizationFeatures))
	for _, f := range validCostOptimizationFeatures {
		valid[f] = struct{}{}
	}

	var invalid []string
	hasDbtState := false
	for _, f := range features {
		if _, ok := valid[f]; !ok {
			invalid = append(invalid, f)
		}
		if f == CostOptimizationFeatureDbtState {
			hasDbtState = true
		}
	}

	if len(invalid) > 0 {
		sort.Strings(invalid)
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid cost_optimization_features value",
			fmt.Sprintf(
				"%s %s not valid. Valid values are: %s.",
				strings.Join(invalid, ", "),
				pluralIsAre(len(invalid)),
				strings.Join(validCostOptimizationFeatures, ", "),
			),
		)
		return
	}

	// The API collapses any set containing dbt_state down to ["dbt_state"]
	// (dropping state_aware_orchestration / efficient_testing). Reject the mixed
	// configuration at plan time so the applied state matches the configuration
	// instead of silently diverging.
	if hasDbtState && len(features) > 1 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid cost_optimization_features combination",
			"When dbt_state is enabled it must be the only cost optimization feature. "+
				"dbt State takes precedence over the other features, so combining it with "+
				"state_aware_orchestration or efficient_testing is not supported. "+
				"Set cost_optimization_features = [\"dbt_state\"].",
		)
	}
}

func pluralIsAre(n int) string {
	if n == 1 {
		return "is"
	}
	return "are"
}

// CostOptimizationFeaturesValidator returns a validator that ensures
// cost_optimization_features only contains supported values and that dbt_state,
// when present, is the only feature in the set.
func CostOptimizationFeaturesValidator() validator.Set {
	return costOptimizationFeaturesValidator{}
}
