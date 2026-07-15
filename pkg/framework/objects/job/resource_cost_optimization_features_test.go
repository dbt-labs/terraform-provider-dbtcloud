package job

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFilterCostOptimizationFeaturesForCIOrMerge(t *testing.T) {
	t.Parallel()

	forceNodeSelection := false
	tests := []struct {
		name    string
		input   []string
		want    []string
		wantNil bool
	}{
		{
			name:    "nil remains unset",
			input:   nil,
			wantNil: true,
		},
		{
			name:  "explicit empty remains a clear request",
			input: []string{},
			want:  []string{},
		},
		{
			name:    "legacy SAO is omitted",
			input:   []string{costOptimizationFeatureStateAwareOrchestration},
			wantNil: true,
		},
		{
			name:  "exclusive dbt State is sent",
			input: []string{costOptimizationFeatureDbtState},
			want:  []string{costOptimizationFeatureDbtState},
		},
		{
			name:    "invalid mixed feature set is safely omitted",
			input:   []string{costOptimizationFeatureDbtState, costOptimizationFeatureStateAwareOrchestration},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotForceNodeSelection, got := filterSAOFieldsForCIOrMerge(true, &forceNodeSelection, tt.input)
			if gotForceNodeSelection != nil {
				t.Fatalf("force_node_selection = %t, want omitted", *gotForceNodeSelection)
			}
			if (got == nil) != tt.wantNil {
				t.Fatalf("filterSAOFieldsForCIOrMerge(%v) feature nil = %t, want %t", tt.input, got == nil, tt.wantNil)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("filterSAOFieldsForCIOrMerge(%v) features = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}

func TestReconcileCostOptimizationFeaturesForCIOrMerge(t *testing.T) {
	t.Parallel()

	falseValue := false
	trueValue := true
	nullFeatures := types.SetNull(types.StringType)
	emptyFeatures := costOptimizationFeatureSet()
	legacySAOFeatures := costOptimizationFeatureSet(costOptimizationFeatureStateAwareOrchestration)
	dbtStateFeatures := costOptimizationFeatureSet(costOptimizationFeatureDbtState)

	tests := []struct {
		name                 string
		jobType              string
		desiredFeatures      types.Set
		priorFeatures        types.Set
		apiForceNode         *bool
		apiFeatures          []string
		wantAPIAuthoritative bool
		wantFeatures         types.Set
	}{
		{
			name:                 "create: dbt State response is authoritative",
			jobType:              JobTypeCI,
			desiredFeatures:      dbtStateFeatures,
			priorFeatures:        nullFeatures,
			apiForceNode:         &falseValue,
			apiFeatures:          []string{costOptimizationFeatureDbtState},
			wantAPIAuthoritative: true,
			wantFeatures:         dbtStateFeatures,
		},
		{
			name:                 "create: omitted dbt State response becomes empty",
			jobType:              JobTypeCI,
			desiredFeatures:      dbtStateFeatures,
			priorFeatures:        nullFeatures,
			apiFeatures:          nil,
			wantAPIAuthoritative: true,
			wantFeatures:         emptyFeatures,
		},
		{
			name:                 "read: omitted dbt State response clears instead of deriving SAO from force node selection",
			jobType:              JobTypeCI,
			desiredFeatures:      dbtStateFeatures,
			priorFeatures:        dbtStateFeatures,
			apiForceNode:         &falseValue,
			apiFeatures:          nil,
			wantAPIAuthoritative: true,
			wantFeatures:         emptyFeatures,
		},
		{
			name:                 "update: explicit API empty response clears dbt State",
			jobType:              JobTypeCI,
			desiredFeatures:      dbtStateFeatures,
			priorFeatures:        dbtStateFeatures,
			apiForceNode:         &trueValue,
			apiFeatures:          []string{},
			wantAPIAuthoritative: true,
			wantFeatures:         emptyFeatures,
		},
		{
			name:                 "update: prior dbt State keeps an omitted response authoritative after configuration removal",
			jobType:              JobTypeCI,
			desiredFeatures:      emptyFeatures,
			priorFeatures:        dbtStateFeatures,
			apiFeatures:          nil,
			wantAPIAuthoritative: true,
			wantFeatures:         emptyFeatures,
		},
		{
			name:                 "update: prior dbt State keeps an explicit empty response authoritative after configuration removal",
			jobType:              JobTypeCI,
			desiredFeatures:      emptyFeatures,
			priorFeatures:        dbtStateFeatures,
			apiFeatures:          []string{},
			wantAPIAuthoritative: true,
			wantFeatures:         emptyFeatures,
		},
		{
			name:                 "import: API dbt State is retained without prior state",
			jobType:              JobTypeCI,
			desiredFeatures:      nullFeatures,
			priorFeatures:        nullFeatures,
			apiForceNode:         &falseValue,
			apiFeatures:          []string{costOptimizationFeatureDbtState},
			wantAPIAuthoritative: true,
			wantFeatures:         dbtStateFeatures,
		},
		{
			name:                 "legacy SAO: preserve configured feature when API omits first-class field",
			jobType:              JobTypeCI,
			desiredFeatures:      legacySAOFeatures,
			priorFeatures:        legacySAOFeatures,
			apiFeatures:          nil,
			wantAPIAuthoritative: false,
			wantFeatures:         legacySAOFeatures,
		},
		{
			name:                 "legacy SAO: explicit API empty response preserves configured feature",
			jobType:              JobTypeCI,
			desiredFeatures:      legacySAOFeatures,
			priorFeatures:        legacySAOFeatures,
			apiFeatures:          []string{},
			wantAPIAuthoritative: false,
			wantFeatures:         legacySAOFeatures,
		},
		{
			name:                 "legacy SAO: explicit user clear wins over prior feature when API omits field",
			jobType:              JobTypeCI,
			desiredFeatures:      emptyFeatures,
			priorFeatures:        legacySAOFeatures,
			apiFeatures:          nil,
			wantAPIAuthoritative: false,
			wantFeatures:         emptyFeatures,
		},
		{
			name:                 "legacy SAO: explicit user clear wins over prior feature when API returns empty",
			jobType:              JobTypeCI,
			desiredFeatures:      emptyFeatures,
			priorFeatures:        legacySAOFeatures,
			apiFeatures:          []string{},
			wantAPIAuthoritative: false,
			wantFeatures:         emptyFeatures,
		},
		{
			name:                 "non-empty first-class dbt State response overrides legacy SAO compatibility",
			jobType:              JobTypeCI,
			desiredFeatures:      legacySAOFeatures,
			priorFeatures:        legacySAOFeatures,
			apiFeatures:          []string{costOptimizationFeatureDbtState},
			wantAPIAuthoritative: true,
			wantFeatures:         dbtStateFeatures,
		},
		{
			name:                 "CI without dbt State never derives legacy SAO from force node selection",
			jobType:              JobTypeCI,
			desiredFeatures:      nullFeatures,
			priorFeatures:        nullFeatures,
			apiForceNode:         &falseValue,
			apiFeatures:          nil,
			wantAPIAuthoritative: false,
			wantFeatures:         emptyFeatures,
		},
		{
			name:                 "non CI: legacy force node selection bridge remains available",
			jobType:              JobTypeOther,
			desiredFeatures:      nullFeatures,
			priorFeatures:        nullFeatures,
			apiForceNode:         &falseValue,
			apiFeatures:          nil,
			wantAPIAuthoritative: false,
			wantFeatures:         legacySAOFeatures,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAPIAuthoritative := shouldUseAPICostOptimizationFeaturesForCIOrMerge(
				tt.desiredFeatures,
				tt.priorFeatures,
				tt.apiFeatures,
			)
			if gotAPIAuthoritative != tt.wantAPIAuthoritative {
				t.Errorf("API-authoritative decision = %t, want %t", gotAPIAuthoritative, tt.wantAPIAuthoritative)
			}

			gotFeatures := reconcileCostOptimizationFeatures(
				tt.jobType,
				tt.desiredFeatures,
				tt.priorFeatures,
				tt.apiForceNode,
				tt.apiFeatures,
			)
			if !gotFeatures.Equal(tt.wantFeatures) {
				t.Errorf("cost_optimization_features = %s, want %s", gotFeatures.String(), tt.wantFeatures.String())
			}
		})
	}
}

func TestUpdateSAOFilteringUsesEffectiveJobType(t *testing.T) {
	t.Parallel()

	falseValue := false
	legacySAO := []string{costOptimizationFeatureStateAwareOrchestration}
	tests := []struct {
		name            string
		explicitJobType string
		triggers        *JobTriggers
		features        []string
		wantFeatures    []string
		wantFeaturesNil bool
	}{
		{
			name:            "explicit CI",
			explicitJobType: JobTypeCI,
			features:        legacySAO,
			wantFeaturesNil: true,
		},
		{
			name:            "explicit Merge",
			explicitJobType: JobTypeMerge,
			features:        []string{costOptimizationFeatureDbtState},
			wantFeatures:    []string{costOptimizationFeatureDbtState},
		},
		{
			name: "inferred CI from github webhook",
			triggers: &JobTriggers{
				GithubWebhook: types.BoolValue(true),
			},
			features:        legacySAO,
			wantFeaturesNil: true,
		},
		{
			name: "inferred CI from git provider webhook",
			triggers: &JobTriggers{
				GitProviderWebhook: types.BoolValue(true),
			},
			features:     []string{costOptimizationFeatureDbtState},
			wantFeatures: []string{costOptimizationFeatureDbtState},
		},
		{
			name: "inferred Merge from on merge",
			triggers: &JobTriggers{
				OnMerge: types.BoolValue(true),
			},
			features:        legacySAO,
			wantFeaturesNil: true,
		},
		{
			name: "inferred Merge preserves an explicit empty clear",
			triggers: &JobTriggers{
				OnMerge: types.BoolValue(true),
			},
			features:     []string{},
			wantFeatures: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isCIOrMerge := isCIOrMergeJobAtCreate(tt.explicitJobType, tt.triggers)
			if !isCIOrMerge {
				t.Fatal("effective job type was not classified as CI/Merge")
			}

			gotForceNodeSelection, gotFeatures := filterSAOFieldsForCIOrMerge(
				isCIOrMerge,
				&falseValue,
				tt.features,
			)
			if gotForceNodeSelection != nil {
				t.Fatalf("force_node_selection = %t, want omitted", *gotForceNodeSelection)
			}
			if (gotFeatures == nil) != tt.wantFeaturesNil {
				t.Fatalf("cost_optimization_features nil = %t, want %t", gotFeatures == nil, tt.wantFeaturesNil)
			}
			if !reflect.DeepEqual(gotFeatures, tt.wantFeatures) {
				t.Fatalf("cost_optimization_features = %#v, want %#v", gotFeatures, tt.wantFeatures)
			}
		})
	}
}

func costOptimizationFeatureSet(features ...string) types.Set {
	values := make([]attr.Value, 0, len(features))
	for _, feature := range features {
		values = append(values, types.StringValue(feature))
	}
	return types.SetValueMust(types.StringType, values)
}
