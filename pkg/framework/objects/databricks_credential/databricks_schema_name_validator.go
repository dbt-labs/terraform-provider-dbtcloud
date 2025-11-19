package databricks_credential

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatabricksSchemaNameValidator struct{}

func (v DatabricksSchemaNameValidator) Description(ctx context.Context) string {
	return "Validates that the schema/dataset name contains only allowed characters: letters, numbers, underscore, parentheses, quotes, curly braces, dot, and space."
}

func (v DatabricksSchemaNameValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that the schema/dataset name contains only allowed characters: letters, numbers, underscore, parentheses, quotes, curly braces, dot, and space."
}

func (v DatabricksSchemaNameValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Skip validation if the value is unknown or null
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	// Get the value of the schema/dataset name
	schemaName := req.ConfigValue.ValueString()

	// Skip validation for empty strings
	if schemaName == "" {
		return
	}

	// Get the adapter_type to determine if periods are allowed
	// Only databricks adapter allows periods in schema names
	var adapterType types.String
	adapterTypePath := req.Path.ParentPath().AtName("adapter_type")
	diags := req.Config.GetAttribute(ctx, adapterTypePath, &adapterType)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Only spark adapter allows periods in schema names
	// Datapricks adapter does NOT allow periods
	allowPeriods := false
	if !adapterType.IsNull() && !adapterType.IsUnknown() {
		adapterTypeValue := adapterType.ValueString()
		if adapterTypeValue == "spark" {
			allowPeriods = true
		}
	}

	// Define the regex pattern for valid schema/dataset names
	// Allows: letters (case-insensitive), digits, underscore, parentheses, quotes, curly braces, and space
	var validSchemaPattern = `^[a-zA-Z0-9_()"'{} ]*$`
	if allowPeriods {
		validSchemaPattern = `^[a-zA-Z0-9_()"'{}. ]*$`
	}
	matched, err := regexp.MatchString(validSchemaPattern, schemaName)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Regex Error",
			fmt.Sprintf("An error occurred while validating the schema/dataset name: %s", err),
		)
		return
	}

	// If the value does not match the pattern, return an error
	if !matched {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Schema/Dataset Name",
			fmt.Sprintf(
				"The schema/dataset name contains invalid characters. "+
					"Only letters, numbers, underscores, parentheses, quotes, curly braces, dots, and spaces are allowed. "+
					"Got: %s",
				schemaName,
			),
		)
	}
}
