package job

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestReconcileJobType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		desired  types.String
		apiValue string
		want     types.String
	}{
		{
			name:     "scheduled desired, api other - preserve scheduled",
			desired:  types.StringValue(JobTypeScheduled),
			apiValue: JobTypeOther,
			want:     types.StringValue(JobTypeScheduled),
		},
		{
			name:     "other desired, api scheduled - preserve other",
			desired:  types.StringValue(JobTypeOther),
			apiValue: JobTypeScheduled,
			want:     types.StringValue(JobTypeOther),
		},
		{
			name:     "scheduled desired, api scheduled - unchanged",
			desired:  types.StringValue(JobTypeScheduled),
			apiValue: JobTypeScheduled,
			want:     types.StringValue(JobTypeScheduled),
		},
		{
			name:     "scheduled desired, api ci - api authoritative",
			desired:  types.StringValue(JobTypeScheduled),
			apiValue: JobTypeCI,
			want:     types.StringValue(JobTypeCI),
		},
		{
			name:     "ci desired, api ci - unchanged",
			desired:  types.StringValue(JobTypeCI),
			apiValue: JobTypeCI,
			want:     types.StringValue(JobTypeCI),
		},
		{
			name:     "merge desired, api merge - unchanged",
			desired:  types.StringValue(JobTypeMerge),
			apiValue: JobTypeMerge,
			want:     types.StringValue(JobTypeMerge),
		},
		{
			name:     "null desired, api other - take api value",
			desired:  types.StringNull(),
			apiValue: JobTypeOther,
			want:     types.StringValue(JobTypeOther),
		},
		{
			name:     "null desired, empty api - null",
			desired:  types.StringNull(),
			apiValue: "",
			want:     types.StringNull(),
		},
		{
			name:     "scheduled desired, empty api - take api (null)",
			desired:  types.StringValue(JobTypeScheduled),
			apiValue: "",
			want:     types.StringNull(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reconcileJobType(tt.desired, tt.apiValue)
			if !got.Equal(tt.want) {
				t.Errorf("reconcileJobType(%v, %q) = %v, want %v", tt.desired, tt.apiValue, got, tt.want)
			}
		})
	}
}
