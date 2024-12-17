package partial_license_map

import (
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/license_map"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func matchPartial(
	licenseMapModel license_map.LicenseMapResourceModel,
	licenseTypeResponse dbt_cloud.LicenseMap,
) bool {
	return licenseMapModel.LicenseType == types.StringValue(licenseTypeResponse.LicenseType)
}
