package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PrimaryProfileValidator returns an error when primary_profile_id is set
// alongside connection_id, credential_id, or extended_attributes_id. When
// a profile is assigned, the API mirrors the profile's values onto the
// environment, so direct field values would conflict with the API-determined
// values and cause inconsistent Terraform state.
type PrimaryProfileValidator struct{}

func (v PrimaryProfileValidator) Description(ctx context.Context) string {
	return "Errors when primary_profile_id is set alongside connection_id, credential_id, or extended_attributes_id."
}

func (v PrimaryProfileValidator) MarkdownDescription(ctx context.Context) string {
	return "Errors when `primary_profile_id` is set alongside `connection_id`, `credential_id`, or `extended_attributes_id`."
}

func (v PrimaryProfileValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	conflictingFields := []struct {
		path string
		name string
	}{
		{"connection_id", "connection_id"},
		{"credential_id", "credential_id"},
		{"extended_attributes_id", "extended_attributes_id"},
	}

	var setFields []string
	for _, field := range conflictingFields {
		var val types.Int64
		diags := req.Config.GetAttribute(ctx, path.Root(field.path), &val)
		if diags.HasError() {
			continue
		}
		if !val.IsNull() && !val.IsUnknown() && val.ValueInt64() != 0 {
			setFields = append(setFields, field.name)
		}
	}

	if len(setFields) > 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Conflicting profile and direct field configuration",
			"When primary_profile_id is set, the API determines connection_id, credential_id, and "+
				"extended_attributes_id from the profile. Remove the direct field(s) and manage them "+
				"through the dbtcloud_profile resource instead.",
		)
	}
}
