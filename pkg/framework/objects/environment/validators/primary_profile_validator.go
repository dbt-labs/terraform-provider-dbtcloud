package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PrimaryProfileValidator warns when primary_profile_id is set alongside
// connection_id, credential_id, or extended_attributes_id. When both are
// configured, dbt Cloud's profile mirroring service may propagate the
// environment's direct connection, credentials, and extended attributes
// onto the assigned profile, overwriting the profile's own values. This
// can cause failures in other environments that share the same profile.
type PrimaryProfileValidator struct{}

func (v PrimaryProfileValidator) Description(ctx context.Context) string {
	return "Warns when primary_profile_id is set alongside connection_id, credential_id, or extended_attributes_id."
}

func (v PrimaryProfileValidator) MarkdownDescription(ctx context.Context) string {
	return "Warns when `primary_profile_id` is set alongside `connection_id`, `credential_id`, or `extended_attributes_id`."
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
		resp.Diagnostics.AddAttributeWarning(
			req.Path,
			"Profile mirroring may overwrite profile attributes",
			"Setting primary_profile_id alongside connection_id, credential_id, or extended_attributes_id "+
				"may cause unexpected behavior. When both are configured, dbt Cloud may propagate the "+
				"environment's direct connection, credentials, and extended attributes onto the assigned "+
				"profile, overwriting the profile's own values. This can cause failures in other "+
				"environments that share the same profile. "+
				"Consider managing connection, credentials, and extended attributes through the profile resource instead.",
		)
	}
}
