package helper

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDbtVersionValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		val         types.String
		expectError bool
	}

	testCases := []testCase{
		// Mantle (dbt Core) release tracks
		{name: "valid - latest", val: types.StringValue("latest")},
		{name: "valid - compatible", val: types.StringValue("compatible")},
		{name: "valid - extended", val: types.StringValue("extended")},
		{name: "valid - fallback (Mantle)", val: types.StringValue("fallback")},

		// Fusion release tracks
		{name: "valid - latest-fusion (legacy alias)", val: types.StringValue("latest-fusion")},
		{name: "valid - fusion-stable", val: types.StringValue("fusion-stable")},
		{name: "valid - fusion-extended", val: types.StringValue("fusion-extended")},
		{name: "valid - fusion-nightly", val: types.StringValue("fusion-nightly")},
		{name: "valid - fusion-fallback", val: types.StringValue("fusion-fallback")},

		// Legacy aliases
		{name: "valid - versionless (legacy)", val: types.StringValue("versionless")},

		// Pinned version formats
		{name: "valid - 1.5.0-latest", val: types.StringValue("1.5.0-latest")},
		{name: "valid - 1.10.0-latest", val: types.StringValue("1.10.0-latest")},
		{name: "valid - 1.7.0-pre", val: types.StringValue("1.7.0-pre")},

		// Null / unknown handling
		{name: "null value - no error", val: types.StringNull()},
		{name: "unknown value - no error", val: types.StringUnknown()},

		// Invalid values
		{name: "invalid - empty string", val: types.StringValue(""), expectError: true},
		{name: "invalid - arbitrary text", val: types.StringValue("garbage"), expectError: true},
		{name: "invalid - unknown fusion variant", val: types.StringValue("fusion-experimental"), expectError: true},
		{name: "invalid - bare semver", val: types.StringValue("1.5.0"), expectError: true},
		{name: "invalid - patch suffix not allowed", val: types.StringValue("1.5.1-latest"), expectError: true},
		{name: "invalid - missing suffix", val: types.StringValue("1.5.0-"), expectError: true},
		{name: "invalid - wrong suffix", val: types.StringValue("1.5.0-stable"), expectError: true},
		{name: "invalid - uppercase track", val: types.StringValue("LATEST"), expectError: true},
		{name: "invalid - leading whitespace", val: types.StringValue(" latest"), expectError: true},
		{name: "invalid - trailing whitespace", val: types.StringValue("latest "), expectError: true},
		{name: "invalid - similar word", val: types.StringValue("latestest"), expectError: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			request := validator.StringRequest{
				Path:           path.Root("dbt_version"),
				PathExpression: path.MatchRoot("dbt_version"),
				ConfigValue:    tc.val,
			}
			response := validator.StringResponse{}

			DbtVersionValidator{}.ValidateString(context.Background(), request, &response)

			if tc.expectError && !response.Diagnostics.HasError() {
				t.Fatalf("expected error for value %q, but got none", tc.val.ValueString())
			}
			if !tc.expectError && response.Diagnostics.HasError() {
				t.Fatalf("unexpected error for value %q: %s", tc.val.ValueString(), response.Diagnostics.Errors()[0].Summary())
			}
		})
	}
}

func TestIsFusionVersion(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val      string
		expected bool
	}

	testCases := []testCase{
		// Fusion tracks (true)
		{val: "latest-fusion", expected: true},
		{val: "fusion-stable", expected: true},
		{val: "fusion-extended", expected: true},
		{val: "fusion-nightly", expected: true},
		{val: "fusion-fallback", expected: true},

		// Non-Fusion (false)
		{val: "latest", expected: false},
		{val: "compatible", expected: false},
		{val: "extended", expected: false},
		{val: "fallback", expected: false},
		{val: "versionless", expected: false},
		{val: "1.5.0-latest", expected: false},
		{val: "1.7.0-pre", expected: false},
		{val: "", expected: false},
		{val: "fusion-experimental", expected: false},
		{val: "LATEST-FUSION", expected: false},
	}

	for _, tc := range testCases {
		t.Run(tc.val, func(t *testing.T) {
			t.Parallel()
			if got := IsFusionVersion(tc.val); got != tc.expected {
				t.Fatalf("IsFusionVersion(%q) = %v, want %v", tc.val, got, tc.expected)
			}
		})
	}
}
