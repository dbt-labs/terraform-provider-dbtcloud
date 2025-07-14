package helper

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type DbtVersionValidator struct{}

func (v DbtVersionValidator) Description(ctx context.Context) string {
	return "Validates that the dbt_version is in the format `major.minor.0-latest`, `major.minor.0-pre`, `versionless`, or `latest`."
}

func (v DbtVersionValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that the `dbt_version` is in the format `major.minor.0-latest`, `major.minor.0-pre`, `versionless`, or `latest`."
}

func (v DbtVersionValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Skip validation if the value is unknown or null
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	// Get the value of dbt_version
	dbtVersion := req.ConfigValue.ValueString()

	// Define the regex pattern for valid dbt_version formats
	validVersionPattern := `^(latest|versionless|latest-fusion|[0-9]+\.[0-9]+\.0-(latest|pre))$`
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
			fmt.Sprintf("The `dbt_version` must be in the format `major.minor.0-latest`, `major.minor.0-pre`, `versionless`, `latest` or `latest-fusion`. Got: %s", dbtVersion),
		)
	}
}
