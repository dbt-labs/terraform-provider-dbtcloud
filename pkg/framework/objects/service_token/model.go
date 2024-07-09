package service_token

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ServiceTokenResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	UID         types.String `tfsdk:"uid"`
	Name        types.String `tfsdk:"name"`
	TokenString types.String `tfsdk:"token_string"`
	State       types.Int64  `tfsdk:"state"`

	ServiceTokenPermissions []ServiceTokenPermission `tfsdk:"service_token_permissions"`
}

type ServiceTokenDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	ServiceTokenID types.Int64  `tfsdk:"service_token_id"`
	UID            types.String `tfsdk:"uid"`
	Name           types.String `tfsdk:"name"`

	ServiceTokenPermissions []ServiceTokenPermission `tfsdk:"service_token_permissions"`
}

type ServiceTokenPermission struct {
	PermissionSet                 types.String `tfsdk:"permission_set"`
	AllProjects                   types.Bool   `tfsdk:"all_projects"`
	ProjectID                     types.Int64  `tfsdk:"project_id"`
	WritableEnvironmentCategories types.Set    `tfsdk:"writable_environment_categories"`
}

func ConvertServiceTokenPermissionModelToData(
	ctx context.Context,
	requiredAllPermissions []ServiceTokenPermission,
	serviceTokenID int,
	accountID int,
) ([]dbt_cloud.ServiceTokenPermission, diag.Diagnostics) {
	allPermissionsRequest := make([]dbt_cloud.ServiceTokenPermission, len(requiredAllPermissions))
	allDiags := diag.Diagnostics{}

	for i, permission := range requiredAllPermissions {
		writableEnvs := make([]dbt_cloud.EnvironmentCategory, 0, len(permission.WritableEnvironmentCategories.Elements()))

		diags := permission.WritableEnvironmentCategories.ElementsAs(ctx, &writableEnvs, false)

		allDiags.Append(diags...)

		allPermissionsRequest[i] = dbt_cloud.ServiceTokenPermission{
			ServiceTokenID: serviceTokenID,
			AccountID:      accountID,
			Set:            permission.PermissionSet.ValueString(),
			ProjectID:      int(permission.ProjectID.ValueInt64()),
			AllProjects:    permission.AllProjects.ValueBool(),
			WritableEnvs:   writableEnvs,
		}
	}
	return allPermissionsRequest, allDiags
}

func ConvertServiceTokenPermissionDataToModel(
	ctx context.Context,
	allPermissions []dbt_cloud.ServiceTokenPermission,
) ([]ServiceTokenPermission, diag.Diagnostics) {
	allPermissionsModel := make([]ServiceTokenPermission, len(allPermissions))
	allDiags := diag.Diagnostics{}
	for i, permission := range allPermissions {

		writableEnvs, diags := types.SetValueFrom(
			ctx,
			types.StringType,
			permission.WritableEnvs,
		)

		allDiags.Append(diags...)

		allPermissionsModel[i] = ServiceTokenPermission{
			PermissionSet:                 types.StringValue(permission.Set),
			ProjectID:                     helper.SetIntToInt64OrNull(permission.ProjectID),
			AllProjects:                   types.BoolValue(permission.AllProjects),
			WritableEnvironmentCategories: writableEnvs,
		}
	}
	return allPermissionsModel, allDiags
}

// func CompareGroupPermissions(
// 	group1, group2 GroupPermission,
// ) bool {
// 	listGroup1Envs := helper.StringSetToStringSlice(group1.WritableEnvironmentCategories)
// 	listGroup2Envs := helper.StringSetToStringSlice(group2.WritableEnvironmentCategories)

// 	diffEnv1, diffEnv2 := lo.Difference(
// 		listGroup1Envs,
// 		listGroup2Envs,
// 	)
// 	return group1.PermissionSet == group2.PermissionSet &&
// 		group1.ProjectID == group2.ProjectID &&
// 		group1.AllProjects == group2.AllProjects &&
// 		len(diffEnv1) == 0 &&
// 		len(diffEnv2) == 0
// }
