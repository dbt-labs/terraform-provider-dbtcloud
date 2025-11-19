package helper

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Pre-compiled regex for performance
var schemaNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_()"'{} ]*$`)

// SchemaNameValidator returns a validator that ensures schema/dataset names
// contain only allowed characters: letters, numbers, underscore, parentheses,
// quotes, curly braces, dot, and space.
func SchemaNameValidator() validator.String {
	return stringvalidator.RegexMatches(
		schemaNamePattern,
		"The schema/dataset name contains invalid characters. "+
			"Only letters, numbers, underscores, parentheses, quotes, curly braces, dots, and spaces are allowed.",
	)
}
