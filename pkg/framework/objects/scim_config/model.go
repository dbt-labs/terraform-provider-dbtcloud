package scim_config

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SCIMConfigResourceModel struct {
	ID                        types.String `tfsdk:"id"`
	Enabled                   types.Bool   `tfsdk:"enabled"`
	ManualUpdatesAllowed      types.Bool   `tfsdk:"manual_updates_allowed"`
	SCIMControlledLicenseType types.Bool   `tfsdk:"scim_controlled_license_type"`
}
