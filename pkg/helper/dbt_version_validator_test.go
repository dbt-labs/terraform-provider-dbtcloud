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
		// dbt platform release tracks (non-Fusion)
		{name: "valid - latest", val: types.StringValue("latest")},
		{name: "valid - compatible", val: types.StringValue("compatible")},
		{name: "valid - extended", val: types.StringValue("extended")},
		{name: "valid - fallback", val: types.StringValue("fallback")},

		// Fusion release tracks
		{name: "valid - latest-fusion (legacy alias)", val: types.StringValue("latest-fusion")},
		{name: "valid - fusion-stable", val: types.StringValue("fusion-stable")},
		{name: "valid - fusion-extended", val: types.StringValue("fusion-extended")},
		{name: "valid - fusion-nightly", val: types.StringValue("fusion-nightly")},
		{name: "valid - fusion-fallback", val: types.StringValue("fusion-fallback")},

		// Legacy aliases
		{name: "valid - versionless (legacy)", val: types.StringValue("versionless")},

		// v1-/v2- spellings of the release tracks
		{name: "valid - v1-latest", val: types.StringValue("v1-latest")},
		{name: "valid - v1-compatible", val: types.StringValue("v1-compatible")},
		{name: "valid - v1-extended", val: types.StringValue("v1-extended")},
		{name: "valid - v2-stable", val: types.StringValue("v2-stable")},
		{name: "valid - v2-extended", val: types.StringValue("v2-extended")},
		{name: "valid - v2-nightly", val: types.StringValue("v2-nightly")},
		{name: "valid - v2-fallback", val: types.StringValue("v2-fallback")},

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
		{name: "invalid - v1-fallback is not a track", val: types.StringValue("v1-fallback"), expectError: true},
		{name: "invalid - v2-latest is not a track", val: types.StringValue("v2-latest"), expectError: true},
		{name: "invalid - v3 track", val: types.StringValue("v3-stable"), expectError: true},
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
		{val: "v2-stable", expected: true},
		{val: "v2-extended", expected: true},
		{val: "v2-nightly", expected: true},
		{val: "v2-fallback", expected: true},

		// Non-Fusion (false)
		{val: "latest", expected: false},
		{val: "compatible", expected: false},
		{val: "extended", expected: false},
		{val: "fallback", expected: false},
		{val: "versionless", expected: false},
		{val: "v1-latest", expected: false},
		{val: "v1-compatible", expected: false},
		{val: "v1-extended", expected: false},
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

func TestEveryDbtVersionAliasIsValid(t *testing.T) {
	t.Parallel()

	for alias, canonical := range dbtVersionAliases {
		for _, v := range []string{alias, canonical} {
			response := validator.StringResponse{}
			DbtVersionValidator{}.ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("dbt_version"),
				ConfigValue: types.StringValue(v),
			}, &response)
			if response.Diagnostics.HasError() {
				t.Errorf("validator rejects %q: %s", v, response.Diagnostics.Errors()[0].Detail())
			}
		}
	}
}

func TestDbtVersionFromAPI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		configured types.String
		apiVersion string
		expected   types.String
	}{
		{name: "same spelling", configured: types.StringValue("latest"), apiVersion: "latest", expected: types.StringValue("latest")},
		{name: "config latest, API v1-latest", configured: types.StringValue("latest"), apiVersion: "v1-latest", expected: types.StringValue("latest")},
		{name: "config v1-latest, API latest", configured: types.StringValue("v1-latest"), apiVersion: "latest", expected: types.StringValue("v1-latest")},
		{name: "config versionless, API latest", configured: types.StringValue("versionless"), apiVersion: "latest", expected: types.StringValue("versionless")},
		{name: "config versionless, API v1-latest", configured: types.StringValue("versionless"), apiVersion: "v1-latest", expected: types.StringValue("versionless")},
		{name: "config fusion-stable, API v2-stable", configured: types.StringValue("fusion-stable"), apiVersion: "v2-stable", expected: types.StringValue("fusion-stable")},
		{name: "config v2-fallback, API fusion-fallback", configured: types.StringValue("v2-fallback"), apiVersion: "fusion-fallback", expected: types.StringValue("v2-fallback")},
		{name: "different version", configured: types.StringValue("latest"), apiVersion: "v2-stable", expected: types.StringValue("v2-stable")},
		{name: "latest-fusion is not fusion-stable", configured: types.StringValue("latest-fusion"), apiVersion: "fusion-stable", expected: types.StringValue("fusion-stable")},
		{name: "null config after import", configured: types.StringNull(), apiVersion: "v1-latest", expected: types.StringValue("v1-latest")},
		{name: "unknown config", configured: types.StringUnknown(), apiVersion: "latest", expected: types.StringValue("latest")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := DbtVersionFromAPI(tc.configured, tc.apiVersion); !got.Equal(tc.expected) {
				t.Fatalf("DbtVersionFromAPI(%s, %q) = %s, want %s", tc.configured, tc.apiVersion, got, tc.expected)
			}
		})
	}
}
