package partial_license_map

import (
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TODO: Move the model to the non partial when moved to the Framework
type LicenseMapResourceModel struct {
	ID                      types.Int64  `tfsdk:"id"`
	LicenseType             types.String `tfsdk:"license_type"`
	SSOLicenseMappingGroups types.Set    `tfsdk:"sso_license_mapping_groups"`
}

func matchPartial(
	licenseMapModel LicenseMapResourceModel,
	licenseTypeResponse dbt_cloud.LicenseMap,
) bool {
	return licenseMapModel.LicenseType == types.StringValue(licenseTypeResponse.LicenseType)
}
