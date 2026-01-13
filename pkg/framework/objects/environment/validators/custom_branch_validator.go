package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CustomBranchValidator validates that custom_branch is set when use_custom_branch is true.
// This validator should be attached to the custom_branch field.
type CustomBranchValidator struct{}

func (v CustomBranchValidator) Description(ctx context.Context) string {
	return "Validates that custom_branch must be set when use_custom_branch is true, and use_custom_branch must be true when custom_branch is set."
}

func (v CustomBranchValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that `custom_branch` must be set when `use_custom_branch` is true, and `use_custom_branch` must be true when `custom_branch` is set."
}

func (v CustomBranchValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Skip validation if custom_branch is unknown (will be validated when value is known)
	if req.ConfigValue.IsUnknown() {
		return
	}

	// Get the value of use_custom_branch
	var useCustomBranch types.Bool
	useCustomBranchPath := path.Root("use_custom_branch")
	diags := req.Config.GetAttribute(ctx, useCustomBranchPath, &useCustomBranch)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	customBranchIsSet := !req.ConfigValue.IsNull() && req.ConfigValue.ValueString() != ""

	// If custom_branch is set but use_custom_branch is false
	if customBranchIsSet && !useCustomBranch.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Inconsistent custom branch configuration",
			"When custom_branch is specified, use_custom_branch must be set to true. "+
				"Either set use_custom_branch to true or remove the custom_branch attribute.",
		)
	}
}

// UseCustomBranchValidator validates that use_custom_branch requires custom_branch to be set.
// This validator should be attached to the use_custom_branch field.
type UseCustomBranchValidator struct{}

func (v UseCustomBranchValidator) Description(ctx context.Context) string {
	return "Validates that custom_branch must be set when use_custom_branch is true."
}

func (v UseCustomBranchValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that `custom_branch` must be set when `use_custom_branch` is true."
}

func (v UseCustomBranchValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	// Skip validation if use_custom_branch is unknown or null
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	// If use_custom_branch is false, no need to check for custom_branch
	if !req.ConfigValue.ValueBool() {
		return
	}

	// Get the value of custom_branch
	var customBranch types.String
	customBranchPath := path.Root("custom_branch")
	diags := req.Config.GetAttribute(ctx, customBranchPath, &customBranch)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Skip validation if custom_branch is unknown (will be validated when value is known)
	if customBranch.IsUnknown() {
		return
	}

	customBranchIsSet := !customBranch.IsNull() && customBranch.ValueString() != ""

	// If use_custom_branch is true but custom_branch is not set
	if !customBranchIsSet {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing custom_branch",
			"When use_custom_branch is set to true, custom_branch must be specified.",
		)
	}
}
