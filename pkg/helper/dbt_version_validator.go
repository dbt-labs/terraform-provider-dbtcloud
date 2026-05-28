package helper

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type DbtVersionValidator struct{}

const validDbtVersionsDescription = "`major.minor.0-latest`, `major.minor.0-pre`, `compatible`, `extended`, `versionless`, `latest`, `fallback`, `latest-fusion`, `fusion-stable`, `fusion-extended`, `fusion-nightly` or `fusion-fallback`"

func (v DbtVersionValidator) Description(ctx context.Context) string {
	return "Validates that the dbt_version is in the format " + validDbtVersionsDescription + "."
}

func (v DbtVersionValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that the `dbt_version` is in the format " + validDbtVersionsDescription + "."
}

func (v DbtVersionValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Skip validation if the value is unknown or null
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	// Get the value of dbt_version
	dbtVersion := req.ConfigValue.ValueString()

	// Define the regex pattern for valid dbt_version formats
	validVersionPattern := `^(compatible|extended|latest|fallback|versionless|latest-fusion|fusion-stable|fusion-extended|fusion-nightly|fusion-fallback|[0-9]+\.[0-9]+\.0-(latest|pre))$`
	matched, err := regexp.MatchString(validVersionPattern, dbtVersion)
	if err != nil {
		resp.Diagnostics.AddError(
			"Regex Error",
			fmt.Sprintf("An error occurred while validating the dbt_version: %s", err),
		)
		return
	}

	// If the value does not match the pattern, return an error
	if !matched {
		resp.Diagnostics.AddError(
			"Invalid dbt_version Format",
			fmt.Sprintf("The `dbt_version` must be in the format "+validDbtVersionsDescription+". Got: %s", dbtVersion),
		)
	}
}

// IsFusionVersion reports whether the supplied dbt_version string is one of
// the Fusion release tracks.
func IsFusionVersion(dbtVersion string) bool {
	switch dbtVersion {
	case "latest-fusion", "fusion-stable", "fusion-extended", "fusion-nightly", "fusion-fallback":
		return true
	}
	return false
}
