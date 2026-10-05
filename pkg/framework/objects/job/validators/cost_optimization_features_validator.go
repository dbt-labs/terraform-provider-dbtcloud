package job_validators

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Valid cost_optimization_features values. The API also accepts
// efficient_testing, which the provider does not expose.
const (
	CostOptimizationFeatureStateAwareOrchestration = "state_aware_orchestration"
	CostOptimizationFeatureDbtState                = "dbt_state"
	CostOptimizationFeatureInheritEnvironment      = "inherit_environment"
)

var validCostOptimizationFeatures = []string{
	CostOptimizationFeatureStateAwareOrchestration,
	CostOptimizationFeatureDbtState,
	CostOptimizationFeatureInheritEnvironment,
}

// The API rejects these unless they are sent on their own.
var exclusiveCostOptimizationFeatures = []string{
	CostOptimizationFeatureDbtState,
	CostOptimizationFeatureInheritEnvironment,
}

var _ validator.Set = &costOptimizationFeaturesValidator{}

type costOptimizationFeaturesValidator struct{}

func (v costOptimizationFeaturesValidator) Description(ctx context.Context) string {
	return fmt.Sprintf(
		"each value must be one of %s; dbt_state and inherit_environment must each be the only feature",
		strings.Join(validCostOptimizationFeatures, ", "),
	)
}

func (v costOptimizationFeaturesValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf(
		"Each value must be one of `%s`. `dbt_state` and `inherit_environment` must each be the only feature.",
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

	exclusive := make(map[string]struct{}, len(exclusiveCostOptimizationFeatures))
	for _, f := range exclusiveCostOptimizationFeatures {
		exclusive[f] = struct{}{}
	}

	var invalid []string
	var present []string
	for _, f := range features {
		if _, ok := valid[f]; !ok {
			invalid = append(invalid, f)
		}
		if _, ok := exclusive[f]; ok {
			present = append(present, f)
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

	// The API rejects a mixed set, so reject it at plan time instead of letting
	// the apply fail.
	if len(present) > 0 && len(features) > 1 {
		sort.Strings(present)
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid cost_optimization_features combination",
			fmt.Sprintf(
				"%s takes precedence over the other features and must be the only one. "+
					"Set cost_optimization_features = [%q].",
				strings.Join(present, " and "),
				present[0],
			),
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
// cost_optimization_features only contains supported values and that an
// exclusive feature, when present, is the only one in the set.
func CostOptimizationFeaturesValidator() validator.Set {
	return costOptimizationFeaturesValidator{}
}
