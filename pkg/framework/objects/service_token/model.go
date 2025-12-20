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

// ConvertServiceTokenPermissionModelToGrant converts Terraform model permissions to API permission grants
// Used for creating service tokens via permission_grants field
func ConvertServiceTokenPermissionModelToGrant(
	ctx context.Context,
	requiredAllPermissions []ServiceTokenPermission,
) ([]dbt_cloud.ServiceTokenPermissionGrant, diag.Diagnostics) {
	allGrants := make([]dbt_cloud.ServiceTokenPermissionGrant, len(requiredAllPermissions))
	allDiags := diag.Diagnostics{}

	for i, permission := range requiredAllPermissions {
		grant := &allGrants[i]

		grant.PermissionSet = permission.PermissionSet.ValueString()

		// Only set project_id if not all_projects (if all_projects, omit project_id)
		if !permission.AllProjects.ValueBool() {
			projectID := int(permission.ProjectID.ValueInt64())
			grant.ProjectID = &projectID
		}
		// If all_projects, ProjectID remains nil (omitempty)

		// Handle writable_environment_categories
		if !permission.WritableEnvironmentCategories.IsUnknown() {
			writableEnvs := make([]dbt_cloud.EnvironmentCategory, 0, len(permission.WritableEnvironmentCategories.Elements()))
			allDiags.Append(permission.WritableEnvironmentCategories.ElementsAs(ctx, &writableEnvs, false)...)

			// small hack to avoid sending all environments if all is present
			if lo.Contains(writableEnvs, dbt_cloud.EnvironmentCategory_All) {
				// Don't set WritableEnvironmentCategories if "all" is present (API default behavior)
				grant.WritableEnvironmentCategories = nil
			} else {
				// Explicitly set empty array if empty, or the actual values
				// Use pointer to distinguish between nil (not set) and empty slice (explicitly empty)
				grant.WritableEnvironmentCategories = &writableEnvs
			}
		} else {
			// If unknown, don't set (let API use default)
			grant.WritableEnvironmentCategories = nil
		}
	}
	return allGrants, allDiags
}

// ConvertServiceTokenPermissionModelToDataForCreation converts permissions for token creation
// ServiceTokenID is not set as it will be assigned by the API during creation
// DEPRECATED: Use ConvertServiceTokenPermissionModelToGrant instead for creation
func ConvertServiceTokenPermissionModelToDataForCreation(
	ctx context.Context,
	requiredAllPermissions []ServiceTokenPermission,
	accountID int,
) ([]dbt_cloud.ServiceTokenPermission, diag.Diagnostics) {
	allPermissionsRequest := make([]dbt_cloud.ServiceTokenPermission, len(requiredAllPermissions))
	allDiags := diag.Diagnostics{}

	for i, permission := range requiredAllPermissions {
		permissionRequest := &allPermissionsRequest[i]

		// ServiceTokenID is not set during creation - API will assign it
		permissionRequest.AccountID = accountID
		permissionRequest.Set = permission.PermissionSet.ValueString()
		permissionRequest.AllProjects = permission.AllProjects.ValueBool()

		if !permissionRequest.AllProjects {
			permissionRequest.ProjectID = int(permission.ProjectID.ValueInt64())
		}

		// Always set WritableEnvs explicitly, even if empty, to ensure consistent API behavior
		if !permission.WritableEnvironmentCategories.IsUnknown() {
			writableEnvs := make([]dbt_cloud.EnvironmentCategory, 0, len(permission.WritableEnvironmentCategories.Elements()))
			allDiags.Append(permission.WritableEnvironmentCategories.ElementsAs(ctx, &writableEnvs, false)...)

			// small hack to avoid sending all environments if all is present
			if lo.Contains(writableEnvs, dbt_cloud.EnvironmentCategory_All) {
				// Don't set WritableEnvs if "all" is present (API default behavior)
				permissionRequest.WritableEnvs = nil
			} else {
				// Explicitly set empty array if empty, or the actual values
				permissionRequest.WritableEnvs = writableEnvs
			}
		} else {
			// If unknown, don't set (let API use default)
			permissionRequest.WritableEnvs = nil
		}

	}
	return allPermissionsRequest, allDiags
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
