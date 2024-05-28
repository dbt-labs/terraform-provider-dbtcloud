package group_partial_permissions

import (
	"context"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

var (
	_ resource.Resource              = &groupPartialPermissionsResource{}
	_ resource.ResourceWithConfigure = &groupPartialPermissionsResource{}
)

func GroupPartialPermissionsResource() resource.Resource {
	return &groupPartialPermissionsResource{}
}

type groupPartialPermissionsResource struct {
	client *dbt_cloud.Client
}

func (r *groupPartialPermissionsResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_group_partial_permissions"
}

func (r *groupPartialPermissionsResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state GroupPartialPermissionsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	// check if the ID exists
	groupIDFromState := state.ID.ValueInt64()
	group, err := r.client.GetGroup(int(groupIDFromState))
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The notification resource was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Issue getting Group",
			"Error: "+err.Error(),
		)
		return
	}

	// if the ID exists, make sure that it is the one we are looking for
	if group.Name != state.Name.ValueString() {
		// it doesn't match, we need to find the correct one
		groupIDs := r.client.GetAllGroupIDsByName(state.Name.ValueString())
		if len(groupIDs) > 1 {
			resp.Diagnostics.AddError(
				"More than one group with the same name",
				"Error: With the `group_partial_permissions` resource, the group name needs to be unique in dbt Cloud",
			)
			return
		}
		if len(groupIDs) == 0 {
			resp.State.RemoveResource(ctx)
			resp.Diagnostics.AddWarning(
				"Group not found",
				"Error: No group was found with the name mentioned",
			)
			return
		}

		groupID := groupIDs[0]
		group, err = r.client.GetGroup(groupID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Issue getting Group",
				"Error: "+err.Error(),
			)
			return
		}
	}

	// we set the "global" values
	state.ID = types.Int64Value(int64(*group.ID))
	state.Name = types.StringValue(group.Name)
	state.AssignByDefault = types.BoolValue(group.AssignByDefault)
	state.SSOMappingGroups, _ = types.SetValueFrom(
		context.Background(),
		types.StringType,
		group.SSOMappingGroups,
	)

	// we set the "partial" values
	var remotePermissions []GroupPermission
	for _, permission := range group.Permissions {
		perm := GroupPermission{
			PermissionSet: types.StringValue(permission.Set),
			ProjectID:     helper.SetIntToInt64OrNull(permission.ProjectID),
			AllProjects:   types.BoolValue(permission.AllProjects),
		}
		remotePermissions = append(remotePermissions, perm)
	}

	relevantPermissions := lo.Intersect(state.GroupPermissions, remotePermissions)
	state.GroupPermissions = relevantPermissions

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (r *groupPartialPermissionsResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan GroupPartialPermissionsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	assignByDefault := plan.AssignByDefault.ValueBool()
	var ssoMappingGroups []string
	diags := plan.SSOMappingGroups.ElementsAs(context.Background(), &ssoMappingGroups, false)
	if diags.HasError() {
		return
	}

	// check if it exists and if there is only one with the given name
	groupIDs := r.client.GetAllGroupIDsByName(name)
	if len(groupIDs) > 1 {
		resp.Diagnostics.AddError(
			"More than one group with the same name",
			"Error: With the `group_partial_permissions` resource, the group name needs to be unique in dbt Cloud",
		)
		return
	}

	if len(groupIDs) == 1 {
		// if it exists get the ID and:
		//   A. update the fields that are not partial, e.g. assignByDefault, ssoMappingGroups
		//   B. add the permission needed for the partial field
		groupID := groupIDs[0]

		group, err := r.client.GetGroup(groupID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Issue getting Group",
				"Error: "+err.Error(),
			)
			return
		}

		// A. update the "global" fields if required
		sameAssignByDefault := group.AssignByDefault == assignByDefault
		sameSSOGroups := lo.Every(group.SSOMappingGroups, ssoMappingGroups) &&
			lo.Every(ssoMappingGroups, group.SSOMappingGroups)

		if !sameAssignByDefault || !sameSSOGroups {
			group.AssignByDefault = assignByDefault
			group.SSOMappingGroups = ssoMappingGroups

			r.client.UpdateGroup(groupID, *group)
		}

		// B. add the permissions that are missing
		configPermissions := plan.GroupPermissions
		remotePermissions := convertGroupPermissionDataToModel(group.Permissions)

		missingPermissions := lo.Without(configPermissions, remotePermissions...)

		if len(missingPermissions) == 0 {
			plan.ID = types.Int64Value(int64(groupID))
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		}

		allPermissions := append(remotePermissions, missingPermissions...)
		allPermissionsRequest := convertGroupPermissionModelToData(
			allPermissions,
			groupID,
			group.AccountID,
		)

		_, err = r.client.UpdateGroupPermissions(*group.ID, allPermissionsRequest)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to assign permissions to the group",
				"Error: "+err.Error(),
			)
			return
		}
		plan.ID = types.Int64Value(int64(groupID))
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	} else {
		// if the group with the name given doesn't exist , create it
		// TODO: Move this to the group resources once the resource is move to the Framework

		group, err := r.client.CreateGroup(name, assignByDefault, ssoMappingGroups)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to create group",
				"Error: "+err.Error(),
			)
			return
		}

		groupPermissions := convertGroupPermissionModelToData(plan.GroupPermissions, *group.ID, group.AccountID)

		_, err = r.client.UpdateGroupPermissions(*group.ID, groupPermissions)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to assign permissions to the group",
				"Error: "+err.Error(),
			)
			return
		}
		plan.ID = types.Int64Value(int64(*group.ID))
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}

}

func (r *groupPartialPermissionsResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state GroupPartialPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := int(state.ID.ValueInt64())
	group, err := r.client.GetGroup(groupID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Issue getting Group",
			"Error: "+err.Error(),
		)
	}

	remotePermissions := convertGroupPermissionDataToModel(group.Permissions)
	requiredAllPermissions := lo.Without(remotePermissions, state.GroupPermissions...)

	if len(requiredAllPermissions) > 0 {
		// if there are permissions left, we delete the ones from the resource
		// but we keep the remote group
		allPermissionsRequest := convertGroupPermissionModelToData(
			requiredAllPermissions,
			groupID,
			group.AccountID,
		)

		_, err = r.client.UpdateGroupPermissions(groupID, allPermissionsRequest)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to assign permissions to the group",
				"Error: "+err.Error(),
			)
			return
		}

	} else {
		// otherwise, we delete the group entirely if there is no permission
		group.State = dbt_cloud.STATE_DELETED
		_, err = r.client.UpdateGroup(groupID, *group)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to delete group",
				"Error: "+err.Error(),
			)
			return
		}
	}

}

func (r *groupPartialPermissionsResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state GroupPartialPermissionsResourceModel

	// Read plan and state values into the models
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	groupID := int(state.ID.ValueInt64())
	group, err := r.client.GetGroup(groupID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Issue getting Group",
			"Error: "+err.Error(),
		)
		return
	}

	planAssignByDefault := plan.AssignByDefault.ValueBool()
	var planSsoMappingGroups []string
	diags := plan.SSOMappingGroups.ElementsAs(context.Background(), &planSsoMappingGroups, false)
	if diags.HasError() {
		return
	}

	stateAssignByDefault := state.AssignByDefault.ValueBool()
	var stateSsoMappingGroups []string
	diags = state.SSOMappingGroups.ElementsAs(context.Background(), &stateSsoMappingGroups, false)
	if diags.HasError() {
		return
	}

	// A. we compare the global objects and update them if needed
	sameAssignByDefault := planAssignByDefault == stateAssignByDefault
	sameSSOGroups := lo.Every(planSsoMappingGroups, stateSsoMappingGroups) &&
		lo.Every(stateSsoMappingGroups, planSsoMappingGroups)

	if !sameAssignByDefault || !sameSSOGroups {

		group.AssignByDefault = planAssignByDefault
		group.SSOMappingGroups = planSsoMappingGroups

		_, err = r.client.UpdateGroup(groupID, *group)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update group",
				"Error: "+err.Error(),
			)
		}
		state.AssignByDefault = plan.AssignByDefault
		state.SSOMappingGroups = plan.SSOMappingGroups

	}

	// B. we compare the permissions and update them if needed

	statePermissions := state.GroupPermissions
	planPermissions := plan.GroupPermissions

	remotePermissions := convertGroupPermissionDataToModel(group.Permissions)

	deletedPermissions := lo.Without(statePermissions, planPermissions...)
	newPermissions := lo.Without(planPermissions, statePermissions...)

	requiredAllPermissions := lo.Without(
		lo.Union(remotePermissions, newPermissions),
		deletedPermissions...)

	if len(deletedPermissions) > 0 || len(newPermissions) > 0 {

		allPermissionsRequest := convertGroupPermissionModelToData(
			requiredAllPermissions,
			groupID,
			group.AccountID,
		)

		_, err = r.client.UpdateGroupPermissions(groupID, allPermissionsRequest)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to assign permissions to the group",
				"Error: "+err.Error(),
			)
			return

		}
		state.GroupPermissions = plan.GroupPermissions
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *groupPartialPermissionsResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}
