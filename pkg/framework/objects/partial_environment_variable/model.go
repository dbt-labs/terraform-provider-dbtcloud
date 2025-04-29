package partial_environment_variable

import (
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/environment_variable"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func matchPartial(
	environmentVariableModel environment_variable.EnvironmentVariableResourceModel,
	envVarResponse dbt_cloud.AbstractedEnvironmentVariable,
) bool {
	return environmentVariableModel.Name == types.StringValue(envVarResponse.Name) &&
		environmentVariableModel.ProjectID == types.Int64Value(int64(envVarResponse.ProjectID))
}
