package snowflake_credential

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ConflictValidator struct {
	ConflictingFields []string
}

func (v ConflictValidator) Description(ctx context.Context) string {
	return "Ensures that only one of the conflicting fields is set."
}

func (v ConflictValidator) MarkdownDescription(ctx context.Context) string {
	return "Ensures that only one of the conflicting fields is set."
}

func (v ConflictValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {

	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return // Skip validation for unknown or null values
	}

	for _, field := range v.ConflictingFields {
		// Get the path of the conflicting field
		conflictingPath := req.Path.ParentPath().AtName(field)

		// Define a variable to hold the value of the conflicting field
		var conflictingValue types.String

		// Retrieve the value of the conflicting field
		diags := req.Config.GetAttribute(ctx, conflictingPath, &conflictingValue)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		// Check if the conflicting field is known and not null
		if !conflictingValue.IsNull() {
			resp.Diagnostics.AddError(
				"Conflicting Fields",
				fmt.Sprintf(
					"Only one of [%s] can be set. Both `%s` and `%s` are set.",
					strings.Join(v.ConflictingFields, ", "),
					req.Path.String(),
					conflictingPath.String(),
				),
			)
			return
		}
	}
}
