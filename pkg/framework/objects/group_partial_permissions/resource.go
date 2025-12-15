package group_partial_permissions

import (
	"context"
	"fmt"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/group"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	var state group.GroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	// check if the ID exists
	groupIDFromState := state.ID.ValueInt64()
	retrievedGroup, err := r.client.GetGroup(int(groupIDFromState))
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
	if retrievedGroup.Name != state.Name.ValueString() {
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
		retrievedGroup, err = r.client.GetGroup(groupID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Issue getting Group",
				"Error: "+err.Error(),
			)
			return
		}
	}

	// we set the "global" values
	state.ID = types.Int64Value(int64(*retrievedGroup.ID))
	state.Name = types.StringValue(retrievedGroup.Name)
	state.AssignByDefault = types.BoolValue(retrievedGroup.AssignByDefault)
	state.SSOMappingGroups, _ = types.SetValueFrom(
		context.Background(),
		types.StringType,
		retrievedGroup.SSOMappingGroups,
	)

	// we set the "partial" values
	remotePermissions := group.ConvertGroupPermissionDataToModel(retrievedGroup.Permissions)

	relevantPermissions := helper.IntersectBy(
		state.GroupPermissions,
		remotePermissions,
		group.CompareGroupPermissions,
	)

	tflog.Info(
		ctx,
		"READ - Intersection of local and remote",
		map[string]any{
			"Relevant intersected Permissions": fmt.Sprintf("%+v", relevantPermissions),
			"State Permissions":                fmt.Sprintf("%+v", state.GroupPermissions),
			"Remote Permissions":               fmt.Sprintf("%+v", remotePermissions),
		},
	)

	state.GroupPermissions = relevantPermissions

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (r *groupPartialPermissionsResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan group.GroupResourceModel

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

		retrievedGroup, err := r.client.GetGroup(groupID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Issue getting Group",
				"Error: "+err.Error(),
			)
			return
		}

		// A. update the "global" fields if required
		sameAssignByDefault := retrievedGroup.AssignByDefault == assignByDefault
		sameSSOGroups := lo.Every(retrievedGroup.SSOMappingGroups, ssoMappingGroups) &&
			lo.Every(ssoMappingGroups, retrievedGroup.SSOMappingGroups)

		if !sameAssignByDefault || !sameSSOGroups {
			retrievedGroup.AssignByDefault = assignByDefault
			retrievedGroup.SSOMappingGroups = ssoMappingGroups

			r.client.UpdateGroup(groupID, *retrievedGroup)
		}

		// B. add the permissions that are missing
		configPermissions := plan.GroupPermissions
		remotePermissions := group.ConvertGroupPermissionDataToModel(retrievedGroup.Permissions)

		missingPermissions, _ := helper.DifferenceBy(
			configPermissions,
			remotePermissions,
			group.CompareGroupPermissions,
		)

		if len(missingPermissions) == 0 {
			plan.ID = types.Int64Value(int64(groupID))
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		}

		allPermissions := append(remotePermissions, missingPermissions...)
		allPermissionsRequest := group.ConvertGroupPermissionModelToData(
			allPermissions,
			groupID,
			retrievedGroup.AccountID,
		)

		_, err = r.client.UpdateGroupPermissions(*retrievedGroup.ID, allPermissionsRequest)
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
		createdGroup, err := r.client.CreateGroup(name, assignByDefault, ssoMappingGroups)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to create group",
				"Error: "+err.Error(),
			)
			return
		}

		groupPermissions := group.ConvertGroupPermissionModelToData(plan.GroupPermissions, *createdGroup.ID, createdGroup.AccountID)

		_, err = r.client.UpdateGroupPermissions(*createdGroup.ID, groupPermissions)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to assign permissions to the group",
				"Error: "+err.Error(),
			)
			return
		}
		plan.ID = types.Int64Value(int64(*createdGroup.ID))
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}

}

func (r *groupPartialPermissionsResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state group.GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := int(state.ID.ValueInt64())
	retrievedGroup, err := r.client.GetGroup(groupID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Issue getting Group",
			"Error: "+err.Error(),
		)
		return
	}

	remotePermissions := group.ConvertGroupPermissionDataToModel(retrievedGroup.Permissions)
	requiredAllPermissions, _ := helper.DifferenceBy(
		remotePermissions,
		state.GroupPermissions,
		group.CompareGroupPermissions,
	)

	if len(requiredAllPermissions) > 0 {
		// if there are permissions left, we delete the ones from the resource
		// but we keep the remote group
		allPermissionsRequest := group.ConvertGroupPermissionModelToData(
			requiredAllPermissions,
			groupID,
			retrievedGroup.AccountID,
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
		retrievedGroup.State = dbt_cloud.STATE_DELETED
		_, err = r.client.UpdateGroup(groupID, *retrievedGroup)
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
	var plan, state group.GroupResourceModel

	// Read plan and state values into the models
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	groupID := int(state.ID.ValueInt64())
	retrievedGroup, err := r.client.GetGroup(groupID)
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

		retrievedGroup.AssignByDefault = planAssignByDefault
		retrievedGroup.SSOMappingGroups = planSsoMappingGroups

		_, err = r.client.UpdateGroup(groupID, *retrievedGroup)
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

	remotePermissions := group.ConvertGroupPermissionDataToModel(retrievedGroup.Permissions)

	deletedPermissions, newPermissions := helper.DifferenceBy(
		statePermissions,
		planPermissions,
		group.CompareGroupPermissions,
	)

	requiredAllPermissions, _ := helper.DifferenceBy(
		helper.UnionBy(remotePermissions, newPermissions, group.CompareGroupPermissions),
		deletedPermissions, group.CompareGroupPermissions)

	tflog.Info(
		ctx,
		"UPDATE - Intersection of local and remote",
		map[string]any{
			"Deleted Permissions":     fmt.Sprintf("%+v", deletedPermissions),
			"New Permissions":         fmt.Sprintf("%+v", newPermissions),
			"Required all Permission": fmt.Sprintf("%+v", requiredAllPermissions),
			"Remote Permissions":      fmt.Sprintf("%+v", remotePermissions),
			"Remote Permissions and New Permissions": fmt.Sprintf(
				"%+v",
				helper.UnionBy(remotePermissions, newPermissions, group.CompareGroupPermissions),
			),
		},
	)

	if len(deletedPermissions) > 0 || len(newPermissions) > 0 {

		allPermissionsRequest := group.ConvertGroupPermissionModelToData(
			requiredAllPermissions,
			groupID,
			retrievedGroup.AccountID,
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
