package service_token

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

type ServiceTokenResourceModel struct {
	ID          types.String `tfsdk:"id"`
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
	accountID int64,
) ([]dbt_cloud.ServiceTokenPermission, diag.Diagnostics) {
	allPermissionsRequest := make([]dbt_cloud.ServiceTokenPermission, len(requiredAllPermissions))
	allDiags := diag.Diagnostics{}

	for i, permission := range requiredAllPermissions {
		permissionRequest := &allPermissionsRequest[i]

		permissionRequest.ServiceTokenID = serviceTokenID
		permissionRequest.AccountID = accountID
		permissionRequest.Set = permission.PermissionSet.ValueString()
		permissionRequest.AllProjects = permission.AllProjects.ValueBool()

		if !permissionRequest.AllProjects {
			permissionRequest.ProjectID = int(permission.ProjectID.ValueInt64())
		}

		if !permission.WritableEnvironmentCategories.IsUnknown() {
			writableEnvs := make([]dbt_cloud.EnvironmentCategory, 0, len(permission.WritableEnvironmentCategories.Elements()))
			allDiags.Append(permission.WritableEnvironmentCategories.ElementsAs(ctx, &writableEnvs, false)...)

			// small hack to avoid sending all environments if all is present
			if !lo.Contains(writableEnvs, dbt_cloud.EnvironmentCategory_All) {
				permissionRequest.WritableEnvs = writableEnvs
			}
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

		permissionsModel := &allPermissionsModel[i]

		permissionsModel.PermissionSet = types.StringValue(permission.Set)

		permissionsModel.AllProjects = types.BoolValue(permission.AllProjects)

		if permission.AllProjects {
			permissionsModel.ProjectID = types.Int64Null()
		} else {
			permissionsModel.ProjectID = types.Int64Value(int64(permission.ProjectID))
		}

		if len(permission.WritableEnvs) == 0 {
			permission.WritableEnvs = []dbt_cloud.EnvironmentCategory{dbt_cloud.EnvironmentCategory_All}
		}

		writableEnvs, diags := types.SetValueFrom(
			ctx,
			types.StringType,
			permission.WritableEnvs,
		)
		permissionsModel.WritableEnvironmentCategories = writableEnvs
		allDiags.Append(diags...)
	}
	return allPermissionsModel, allDiags
}
