package service_token

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	accountID int,
) ([]dbt_cloud.ServiceTokenPermission, diag.Diagnostics) {
	allPermissionsRequest := make([]dbt_cloud.ServiceTokenPermission, len(requiredAllPermissions))
	allDiags := diag.Diagnostics{}

	for i, permission := range requiredAllPermissions {
		permissionRequest := &allPermissionsRequest[i]

		permissionRequest.ServiceTokenID = serviceTokenID
		permissionRequest.AccountID = accountID
		permissionRequest.Set = permission.PermissionSet.ValueString()
		permissionRequest.ProjectID = int(permission.ProjectID.ValueInt64())
		permissionRequest.AllProjects = permission.AllProjects.ValueBool()

		if !permission.WritableEnvironmentCategories.IsUnknown() {
			writableEnvs := make([]dbt_cloud.EnvironmentCategory, 0, len(permission.WritableEnvironmentCategories.Elements()))
			allDiags.Append(permission.WritableEnvironmentCategories.ElementsAs(ctx, &writableEnvs, false)...)
			permissionRequest.WritableEnvs = writableEnvs
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
		permissionsModel.ProjectID = helper.SetIntToInt64OrNull(permission.ProjectID)
		permissionsModel.AllProjects = types.BoolValue(permission.AllProjects)

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
