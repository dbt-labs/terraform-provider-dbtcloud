package helper

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// PreferWriteOnlyAttributeValidator emits a warning recommending the write-only
// alternative when a non-write-only attribute has a value set.
//
// It exists because stringvalidator.PreferWriteOnlyAttribute (from
// terraform-plugin-framework-validators) takes an absolute path.Expression and
// evaluates it against the resource root without merging it with the path of the
// attribute being validated. That works for a stand-alone resource, but breaks
// when the same schema is reused inside a nested attribute (for example, the
// credential schemas reused under `credential` for the Semantic Layer
// resources), where the write-only sibling no longer lives at the resource root.
//
// This validator resolves the write-only sibling relative to the attribute being
// validated via req.Path.ParentPath().AtName(...), mirroring the approach already
// used by the conflict validators, so it works whether the schema is used at the
// resource root or nested inside another attribute.
type PreferWriteOnlyAttributeValidator struct {
	// WriteOnlyAttributeName is the name of the sibling write-only attribute that
	// should be preferred over the attribute this validator is applied to.
	WriteOnlyAttributeName string
}

func (v PreferWriteOnlyAttributeValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v PreferWriteOnlyAttributeValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf(
		"The write-only attribute `%s` should be preferred over this attribute",
		v.WriteOnlyAttributeName,
	)
}

func (v PreferWriteOnlyAttributeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Only emit the warning when the Terraform client supports write-only attributes.
	if !req.ClientCapabilities.WriteOnlyAttributesAllowed {
		return
	}

	// No warning is needed when this attribute has no usable value set.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	writeOnlyPath := req.Path.ParentPath().AtName(v.WriteOnlyAttributeName)

	resp.Diagnostics.AddAttributeWarning(
		req.Path,
		"Available Write-Only Attribute Alternative",
		fmt.Sprintf(
			"This attribute has a WriteOnly version %s available. "+
				"Use the WriteOnly version of the attribute when possible.",
			writeOnlyPath.String(),
		),
	)
}
