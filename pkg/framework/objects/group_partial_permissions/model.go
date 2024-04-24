package group_partial_permissions

import (
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GroupPartialPermissionsResourceModel struct {
	ID               types.Int64       `tfsdk:"id"`
	Name             types.String      `tfsdk:"name"`
	AssignByDefault  types.Bool        `tfsdk:"assign_by_default"`
	SSOMappingGroups types.Set         `tfsdk:"sso_mapping_groups"`
	GroupPermissions []GroupPermission `tfsdk:"group_permissions"`
}

type GroupPermission struct {
	PermissionSet types.String `tfsdk:"permission_set"`
	ProjectID     types.Int64  `tfsdk:"project_id"`
	AllProjects   types.Bool   `tfsdk:"all_projects"`
}

func convertGroupPermissionModelToData(
	requiredAllPermissions []GroupPermission,
	groupID int,
	accountID int,
) []dbt_cloud.GroupPermission {
	allPermissionsRequest := make([]dbt_cloud.GroupPermission, len(requiredAllPermissions))
	for i, permission := range requiredAllPermissions {
		allPermissionsRequest[i] = dbt_cloud.GroupPermission{
			GroupID:     groupID,
			AccountID:   accountID,
			Set:         permission.PermissionSet.ValueString(),
			ProjectID:   int(permission.ProjectID.ValueInt64()),
			AllProjects: permission.AllProjects.ValueBool(),
		}
	}
	return allPermissionsRequest
}

func convertGroupPermissionDataToModel(
	allPermissions []dbt_cloud.GroupPermission,
) []GroupPermission {
	allPermissionsModel := make([]GroupPermission, len(allPermissions))
	for i, permission := range allPermissions {
		allPermissionsModel[i] = GroupPermission{
			PermissionSet: types.StringValue(permission.Set),
			ProjectID:     helper.SetIntToInt64OrNull(permission.ProjectID),
			AllProjects:   types.BoolValue(permission.AllProjects),
		}
	}
	return allPermissionsModel
}
