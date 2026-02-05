package scim_group_permissions

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ScimGroupPermissionsResourceModel struct {
	ID               types.Int64                `tfsdk:"id"`
	GroupID          types.Int64                `tfsdk:"group_id"`
	GroupPermissions []ScimGroupPermissionModel `tfsdk:"permissions"`
}

type ScimGroupPermissionModel struct {
	PermissionSet                 types.String `tfsdk:"permission_set"`
	ProjectID                     types.Int64  `tfsdk:"project_id"`
	AllProjects                   types.Bool   `tfsdk:"all_projects"`
	WritableEnvironmentCategories types.Set    `tfsdk:"writable_environment_categories"`
}

func ConvertScimGroupPermissionModelToData(
	requiredAllPermissions []ScimGroupPermissionModel,
	groupID int,
	accountID int64,
) []dbt_cloud.GroupPermission {
	allPermissions := []dbt_cloud.GroupPermission{}

	for _, permission := range requiredAllPermissions {
		writableEnvironmentCategoriesSet, _ := permission.WritableEnvironmentCategories.ToSetValue(context.Background())
		writableEnvironmentCategories := []string{}
		if !writableEnvironmentCategoriesSet.IsNull() {
			for _, category := range writableEnvironmentCategoriesSet.Elements() {
				categoryString, _ := category.ToTerraformValue(context.Background())
				var categoryVal string
				_ = categoryString.As(&categoryVal)
				writableEnvironmentCategories = append(writableEnvironmentCategories, categoryVal)
			}
		}

		groupPermission := dbt_cloud.GroupPermission{
			AccountID:                     accountID,
			GroupID:                       groupID,
			Set:                           permission.PermissionSet.ValueString(),
			AllProjects:                   permission.AllProjects.ValueBool(),
			WritableEnvironmentCategories: writableEnvironmentCategories,
		}

		if !permission.ProjectID.IsNull() {
			projectID := int(permission.ProjectID.ValueInt64())
			groupPermission.ProjectID = projectID
		}

		allPermissions = append(allPermissions, groupPermission)
	}

	return allPermissions
}

func ConvertScimGroupPermissionDataToModel(
	groupPermissions []dbt_cloud.GroupPermission,
) []ScimGroupPermissionModel {
	permissionModels := []ScimGroupPermissionModel{}

	for _, permission := range groupPermissions {
		permissionModel := ScimGroupPermissionModel{
			PermissionSet: types.StringValue(permission.Set),
			AllProjects:   types.BoolValue(permission.AllProjects),
		}

		if permission.ProjectID != 0 {
			permissionModel.ProjectID = types.Int64Value(int64(permission.ProjectID))
		} else {
			permissionModel.ProjectID = types.Int64Null()
		}

		writableEnvironmentCategories, _ := types.SetValueFrom(
			context.Background(),
			types.StringType,
			permission.WritableEnvironmentCategories,
		)
		permissionModel.WritableEnvironmentCategories = writableEnvironmentCategories

		permissionModels = append(permissionModels, permissionModel)
	}

	// Sort for consistent ordering - use a different approach since ScimGroupPermissionModel is not comparable
	return permissionModels
}
