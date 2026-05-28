package job

import (
	"strings"
	"testing"
)

func TestValidateJobTypeChange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		prevType    string
		newType     string
		expectError bool
		errorSubstr string
	}{
		// No change cases - all should pass
		{
			name:        "ci to ci - no change allowed",
			prevType:    JobTypeCI,
			newType:     JobTypeCI,
			expectError: false,
		},
		{
			name:        "scheduled to scheduled - no change allowed",
			prevType:    JobTypeScheduled,
			newType:     JobTypeScheduled,
			expectError: false,
		},
		{
			name:        "other to other - no change allowed",
			prevType:    JobTypeOther,
			newType:     JobTypeOther,
			expectError: false,
		},
		{
			name:        "adaptive to adaptive - no change allowed",
			prevType:    JobTypeAdaptive,
			newType:     JobTypeAdaptive,
			expectError: false,
		},
		{
			name:        "merge to merge - no change allowed",
			prevType:    JobTypeMerge,
			newType:     JobTypeMerge,
			expectError: false,
		},

		// Empty previous type - any new type allowed
		{
			name:        "empty to ci - allowed",
			prevType:    "",
			newType:     JobTypeCI,
			expectError: false,
		},
		{
			name:        "empty to scheduled - allowed",
			prevType:    "",
			newType:     JobTypeScheduled,
			expectError: false,
		},
		{
			name:        "empty to adaptive - allowed",
			prevType:    "",
			newType:     JobTypeAdaptive,
			expectError: false,
		},

		// CI job type transitions - only CI allowed
		{
			name:        "ci to scheduled - not allowed",
			prevType:    JobTypeCI,
			newType:     JobTypeScheduled,
			expectError: true,
			errorSubstr: "can only be set to 'ci'",
		},
		{
			name:        "ci to other - not allowed",
			prevType:    JobTypeCI,
			newType:     JobTypeOther,
			expectError: true,
			errorSubstr: "can only be set to 'ci'",
		},
		{
			name:        "ci to adaptive - not allowed",
			prevType:    JobTypeCI,
			newType:     JobTypeAdaptive,
			expectError: true,
			errorSubstr: "can only be set to 'ci'",
		},
		{
			name:        "ci to merge - not allowed",
			prevType:    JobTypeCI,
			newType:     JobTypeMerge,
			expectError: true,
			errorSubstr: "can only be set to 'ci'",
		},

		// Adaptive job type transitions - only adaptive allowed
		{
			name:        "adaptive to ci - not allowed",
			prevType:    JobTypeAdaptive,
			newType:     JobTypeCI,
			expectError: true,
			errorSubstr: "can only be set to 'adaptive'",
		},
		{
			name:        "adaptive to scheduled - not allowed",
			prevType:    JobTypeAdaptive,
			newType:     JobTypeScheduled,
			expectError: true,
			errorSubstr: "can only be set to 'adaptive'",
		},
		{
			name:        "adaptive to other - not allowed",
			prevType:    JobTypeAdaptive,
			newType:     JobTypeOther,
			expectError: true,
			errorSubstr: "can only be set to 'adaptive'",
		},

		// Merge job type transitions - merge, scheduled, and other are all allowed in-place;
		// transitions to CI or Adaptive are blocked (those would require replacement at the
		// ModifyPlan level but validateJobTypeChange still guards them).
		{
			name:        "merge to ci - not allowed",
			prevType:    JobTypeMerge,
			newType:     JobTypeCI,
			expectError: true,
			errorSubstr: "cannot change job_type to 'ci'",
		},
		{
			name:        "merge to scheduled - allowed in-place",
			prevType:    JobTypeMerge,
			newType:     JobTypeScheduled,
			expectError: false,
		},
		{
			name:        "merge to other - allowed in-place",
			prevType:    JobTypeMerge,
			newType:     JobTypeOther,
			expectError: false,
		},

		// Scheduled job type transitions - scheduled, other, and merge are all allowed in-place
		{
			name:        "scheduled to other - allowed",
			prevType:    JobTypeScheduled,
			newType:     JobTypeOther,
			expectError: false,
		},
		{
			name:        "scheduled to merge - allowed in-place",
			prevType:    JobTypeScheduled,
			newType:     JobTypeMerge,
			expectError: false,
		},
		{
			name:        "scheduled to ci - not allowed",
			prevType:    JobTypeScheduled,
			newType:     JobTypeCI,
			expectError: true,
			errorSubstr: "cannot change job_type to 'ci'",
		},
		{
			name:        "scheduled to adaptive - not allowed",
			prevType:    JobTypeScheduled,
			newType:     JobTypeAdaptive,
			expectError: true,
			errorSubstr: "cannot change job_type to 'adaptive'",
		},

		// Other job type transitions - scheduled, other, and merge are all allowed in-place
		{
			name:        "other to scheduled - allowed",
			prevType:    JobTypeOther,
			newType:     JobTypeScheduled,
			expectError: false,
		},
		{
			name:        "other to merge - allowed in-place",
			prevType:    JobTypeOther,
			newType:     JobTypeMerge,
			expectError: false,
		},
		{
			name:        "other to ci - not allowed",
			prevType:    JobTypeOther,
			newType:     JobTypeCI,
			expectError: true,
			errorSubstr: "cannot change job_type to 'ci'",
		},
		{
			name:        "other to adaptive - not allowed",
			prevType:    JobTypeOther,
			newType:     JobTypeAdaptive,
			expectError: true,
			errorSubstr: "cannot change job_type to 'adaptive'",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateJobTypeChange(tc.prevType, tc.newType)

			if tc.expectError {
				if err == nil {
					t.Errorf("expected error but got none for transition %s -> %s", tc.prevType, tc.newType)
					return
				}
				if tc.errorSubstr != "" && !strings.Contains(err.Error(), tc.errorSubstr) {
					t.Errorf("expected error to contain %q, got %q", tc.errorSubstr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v for transition %s -> %s", err, tc.prevType, tc.newType)
				}
			}
		})
	}
}

func TestValidateJobTypeChange_AllTransitions(t *testing.T) {
	t.Parallel()

	// This test validates the complete transition matrix
	// true = transition allowed, false = transition not allowed
	transitionMatrix := map[string]map[string]bool{
		"": {
			JobTypeCI:        true,
			JobTypeMerge:     true,
			JobTypeScheduled: true,
			JobTypeOther:     true,
			JobTypeAdaptive:  true,
		},
		JobTypeCI: {
			JobTypeCI:        true,
			JobTypeMerge:     false,
			JobTypeScheduled: false,
			JobTypeOther:     false,
			JobTypeAdaptive:  false,
		},
		JobTypeMerge: {
			JobTypeCI:        false,
			JobTypeMerge:     true,
			JobTypeScheduled: true, // in-place transition allowed
			JobTypeOther:     true, // in-place transition allowed
			JobTypeAdaptive:  false,
		},
		JobTypeScheduled: {
			JobTypeCI:        false,
			JobTypeMerge:     true, // in-place transition allowed
			JobTypeScheduled: true,
			JobTypeOther:     true,
			JobTypeAdaptive:  false,
		},
		JobTypeOther: {
			JobTypeCI:        false,
			JobTypeMerge:     true, // in-place transition allowed
			JobTypeScheduled: true,
			JobTypeOther:     true,
			JobTypeAdaptive:  false,
		},
		JobTypeAdaptive: {
			JobTypeCI:        false,
			JobTypeMerge:     false,
			JobTypeScheduled: false,
			JobTypeOther:     false,
			JobTypeAdaptive:  true,
		},
	}

	for prevType, transitions := range transitionMatrix {
		for newType, shouldBeAllowed := range transitions {
			prevType := prevType
			newType := newType
			shouldBeAllowed := shouldBeAllowed

			testName := prevType + " -> " + newType
			if prevType == "" {
				testName = "(empty) -> " + newType
			}

			t.Run(testName, func(t *testing.T) {
				t.Parallel()

				err := validateJobTypeChange(prevType, newType)
				isAllowed := err == nil

				if isAllowed != shouldBeAllowed {
					if shouldBeAllowed {
						t.Errorf("expected transition %s -> %s to be allowed, but got error: %v", prevType, newType, err)
					} else {
						t.Errorf("expected transition %s -> %s to be blocked, but it was allowed", prevType, newType)
					}
				}
			})
		}
	}
}

func TestJobTypeConstants(t *testing.T) {
	t.Parallel()

	// Verify the constants have the expected values
	tests := []struct {
		constant string
		expected string
	}{
		{JobTypeCI, "ci"},
		{JobTypeMerge, "merge"},
		{JobTypeScheduled, "scheduled"},
		{JobTypeOther, "other"},
		{JobTypeAdaptive, "adaptive"},
	}

	for _, tc := range tests {
		if tc.constant != tc.expected {
			t.Errorf("expected constant to be %q, got %q", tc.expected, tc.constant)
		}
	}
}

func TestRequiresJobTypeReplacement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		oldType  string
		newType  string
		expected bool
	}{
		// Same type — never requires replacement
		{name: "ci to ci", oldType: JobTypeCI, newType: JobTypeCI, expected: false},
		{name: "scheduled to scheduled", oldType: JobTypeScheduled, newType: JobTypeScheduled, expected: false},
		{name: "other to other", oldType: JobTypeOther, newType: JobTypeOther, expected: false},
		{name: "merge to merge", oldType: JobTypeMerge, newType: JobTypeMerge, expected: false},
		{name: "adaptive to adaptive", oldType: JobTypeAdaptive, newType: JobTypeAdaptive, expected: false},

		// CI transitions — always require replacement
		{name: "ci to scheduled", oldType: JobTypeCI, newType: JobTypeScheduled, expected: true},
		{name: "ci to other", oldType: JobTypeCI, newType: JobTypeOther, expected: true},
		{name: "ci to merge", oldType: JobTypeCI, newType: JobTypeMerge, expected: true},
		{name: "ci to adaptive", oldType: JobTypeCI, newType: JobTypeAdaptive, expected: true},
		{name: "scheduled to ci", oldType: JobTypeScheduled, newType: JobTypeCI, expected: true},
		{name: "other to ci", oldType: JobTypeOther, newType: JobTypeCI, expected: true},
		{name: "merge to ci", oldType: JobTypeMerge, newType: JobTypeCI, expected: true},

		// Adaptive transitions — always require replacement
		{name: "adaptive to scheduled", oldType: JobTypeAdaptive, newType: JobTypeScheduled, expected: true},
		{name: "adaptive to other", oldType: JobTypeAdaptive, newType: JobTypeOther, expected: true},
		{name: "adaptive to merge", oldType: JobTypeAdaptive, newType: JobTypeMerge, expected: true},
		{name: "scheduled to adaptive", oldType: JobTypeScheduled, newType: JobTypeAdaptive, expected: true},
		{name: "other to adaptive", oldType: JobTypeOther, newType: JobTypeAdaptive, expected: true},
		{name: "merge to adaptive", oldType: JobTypeMerge, newType: JobTypeAdaptive, expected: true},

		// In-place transitions (scheduled / other / merge)
		{name: "scheduled to other", oldType: JobTypeScheduled, newType: JobTypeOther, expected: false},
		{name: "scheduled to merge", oldType: JobTypeScheduled, newType: JobTypeMerge, expected: false},
		{name: "other to scheduled", oldType: JobTypeOther, newType: JobTypeScheduled, expected: false},
		{name: "other to merge", oldType: JobTypeOther, newType: JobTypeMerge, expected: false},
		{name: "merge to scheduled", oldType: JobTypeMerge, newType: JobTypeScheduled, expected: false},
		{name: "merge to other", oldType: JobTypeMerge, newType: JobTypeOther, expected: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := requiresJobTypeReplacement(tc.oldType, tc.newType)
			if got != tc.expected {
				t.Errorf("requiresJobTypeReplacement(%q, %q) = %v, want %v", tc.oldType, tc.newType, got, tc.expected)
			}
		})
	}
}
