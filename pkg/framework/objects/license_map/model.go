package license_map

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type LicenseMapResourceModel struct {
	ID                      types.Int64  `tfsdk:"id"`
	LicenseType             types.String `tfsdk:"license_type"`
	SSOLicenseMappingGroups types.Set    `tfsdk:"sso_license_mapping_groups"`
}
