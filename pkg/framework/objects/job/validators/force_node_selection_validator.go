package job_validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Bool = &forceNodeSelectionValidator{}

type forceNodeSelectionValidator struct{}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v forceNodeSelectionValidator) Description(ctx context.Context) string {
	return "When dbt_version is not 'latest-fusion', force_node_selection must be set to true"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v forceNodeSelectionValidator) MarkdownDescription(ctx context.Context) string {
	return "When `dbt_version` is not `latest-fusion`, `force_node_selection` must be set to `true`"
}

// ValidateBool performs the validation.
func (v forceNodeSelectionValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	// If the value is unknown or null, skip validation
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	// Get the force_node_selection value
	forceNodeSelection := req.ConfigValue.ValueBool()

	// Get the dbt_version value from the config
	var dbtVersion types.String
	diags := req.Config.GetAttribute(ctx, path.Root("dbt_version"), &dbtVersion)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip validation if dbt_version is null or unknown
	if dbtVersion.IsNull() || dbtVersion.IsUnknown() {
		return
	}

	// Validation: if dbt_version is not "latest-fusion", force_node_selection must be true
	if dbtVersion.ValueString() != "latest-fusion" && !forceNodeSelection {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid force_node_selection Configuration",
			fmt.Sprintf(
				"When dbt_version is '%s' (not 'latest-fusion'), force_node_selection must be set to true. "+
					"Set force_node_selection = true or change dbt_version to 'latest-fusion'.",
				dbtVersion.ValueString(),
			),
		)
	}
}

// ForceNodeSelectionValidator returns a validator that ensures force_node_selection is true
// when dbt_version is not "latest-fusion".
func ForceNodeSelectionValidator() validator.Bool {
	return forceNodeSelectionValidator{}
}
