package group

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

type GroupResourceModel struct {
	ID               types.Int64       `tfsdk:"id"`
	Name             types.String      `tfsdk:"name"`
	AssignByDefault  types.Bool        `tfsdk:"assign_by_default"`
	SSOMappingGroups types.Set         `tfsdk:"sso_mapping_groups"`
	GroupPermissions []GroupPermission `tfsdk:"group_permissions"`
}

// we need a different one just because historically the data source uses `group_id` instead of `id`
type GroupDataSourceModel struct {
	ID               types.Int64       `tfsdk:"id"`
	GroupID          types.Int64       `tfsdk:"group_id"`
	Name             types.String      `tfsdk:"name"`
	AssignByDefault  types.Bool        `tfsdk:"assign_by_default"`
	SSOMappingGroups types.Set         `tfsdk:"sso_mapping_groups"`
	GroupPermissions []GroupPermission `tfsdk:"group_permissions"`
}

type GroupsDataSourceModel struct {
	Name         types.String `tfsdk:"name"`
	NameContains types.String `tfsdk:"name_contains"`
	State        types.String `tfsdk:"state"`
	Groups       []GroupInfo  `tfsdk:"groups"`
}

type GroupInfo struct {
	ID               types.Int64  `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	State            types.Int64  `tfsdk:"state"`
	AssignByDefault  types.Bool   `tfsdk:"assign_by_default"`
	SSOMappingGroups types.Set    `tfsdk:"sso_mapping_groups"`
	ScimManaged      types.Bool   `tfsdk:"scim_managed"`
}

type GroupPermission struct {
	PermissionSet                 types.String `tfsdk:"permission_set"`
	ProjectID                     types.Int64  `tfsdk:"project_id"`
	AllProjects                   types.Bool   `tfsdk:"all_projects"`
	WritableEnvironmentCategories types.Set    `tfsdk:"writable_environment_categories"`
}

func ConvertGroupPermissionModelToData(
	requiredAllPermissions []GroupPermission,
	groupID int,
	accountID int64,
) []dbt_cloud.GroupPermission {
	allPermissionsRequest := make([]dbt_cloud.GroupPermission, len(requiredAllPermissions))
	for i, permission := range requiredAllPermissions {
		allPermissionsRequest[i] = dbt_cloud.GroupPermission{
			GroupID:     groupID,
			AccountID:   accountID,
			Set:         permission.PermissionSet.ValueString(),
			ProjectID:   int(permission.ProjectID.ValueInt64()),
			AllProjects: permission.AllProjects.ValueBool(),
			WritableEnvironmentCategories: helper.StringSetToStringSlice(
				permission.WritableEnvironmentCategories,
			),
		}
	}
	return allPermissionsRequest
}

func ConvertGroupPermissionDataToModel(
	allPermissions []dbt_cloud.GroupPermission,
) []GroupPermission {
	allPermissionsModel := make([]GroupPermission, len(allPermissions))
	for i, permission := range allPermissions {

		writeableEnvs, _ := types.SetValueFrom(
			context.Background(),
			types.StringType,
			permission.WritableEnvironmentCategories,
		)

		allPermissionsModel[i] = GroupPermission{
			PermissionSet:                 types.StringValue(permission.Set),
			ProjectID:                     helper.SetIntToInt64OrNull(permission.ProjectID),
			AllProjects:                   types.BoolValue(permission.AllProjects),
			WritableEnvironmentCategories: writeableEnvs,
		}
	}
	return allPermissionsModel
}

func CompareGroupPermissions(
	group1, group2 GroupPermission,
) bool {
	listGroup1Envs := helper.StringSetToStringSlice(group1.WritableEnvironmentCategories)
	listGroup2Envs := helper.StringSetToStringSlice(group2.WritableEnvironmentCategories)

	diffEnv1, diffEnv2 := lo.Difference(
		listGroup1Envs,
		listGroup2Envs,
	)
	return group1.PermissionSet == group2.PermissionSet &&
		group1.ProjectID == group2.ProjectID &&
		group1.AllProjects == group2.AllProjects &&
		len(diffEnv1) == 0 &&
		len(diffEnv2) == 0
}
