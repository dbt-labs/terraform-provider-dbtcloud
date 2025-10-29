package scim_group_partial_permissions

import (
	"context"
	"fmt"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &scimGroupPartialPermissionsResource{}
	_ resource.ResourceWithConfigure = &scimGroupPartialPermissionsResource{}
)

func ScimGroupPartialPermissionsResource() resource.Resource {
	return &scimGroupPartialPermissionsResource{}
}

type scimGroupPartialPermissionsResource struct {
	client *dbt_cloud.Client
}

func (r *scimGroupPartialPermissionsResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_scim_group_partial_permissions"
}

func (r *scimGroupPartialPermissionsResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan ScimGroupPartialPermissionsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := int(plan.GroupID.ValueInt64())

	// Verify the group exists
	existingGroup, err := r.client.GetGroup(groupID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddError(
				"Group Not Found",
				fmt.Sprintf("Group with ID %d does not exist. This resource only manages permissions for existing groups and does not create groups.", groupID),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error retrieving group",
			fmt.Sprintf("Could not retrieve group %d: %s", groupID, err.Error()),
		)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Managing partial permissions for existing group: %s (ID: %d)", existingGroup.Name, *existingGroup.ID))

	// Get current permissions from the group
	remotePermissions := ConvertScimGroupPartialPermissionDataToModel(existingGroup.Permissions)

	// Get the permissions from our plan
	configPermissions := plan.GroupPermissions

	// Find permissions that are missing (need to be added)
	missingPermissions, _ := helper.DifferenceBy(
		configPermissions,
		remotePermissions,
		CompareScimGroupPartialPermissions,
	)

	tflog.Info(
		ctx,
		"CREATE - Analyzing permissions",
		map[string]any{
			"Remote Permissions":  fmt.Sprintf("%+v", remotePermissions),
			"Config Permissions":  fmt.Sprintf("%+v", configPermissions),
			"Missing Permissions": fmt.Sprintf("%+v", missingPermissions),
		},
	)

	// If there are no missing permissions, we're already in sync
	if len(missingPermissions) == 0 {
		plan.ID = types.Int64Value(int64(groupID))
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	// Combine remote permissions with the missing ones
	allPermissions := append(remotePermissions, missingPermissions...)
	allPermissionsRequest := ConvertScimGroupPartialPermissionModelToData(
		allPermissions,
		groupID,
		existingGroup.AccountID,
	)

	// Update the group with the combined permissions
	_, err = r.client.UpdateGroupPermissions(groupID, allPermissionsRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating group permissions",
			"Could not add partial permissions to group, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.Int64Value(int64(groupID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *scimGroupPartialPermissionsResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state ScimGroupPartialPermissionsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := int(state.GroupID.ValueInt64())
	retrievedGroup, err := r.client.GetGroup(groupID)

	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddWarning(
				"Group not found",
				"The group was not found and has been removed from the state. This may indicate the group was deleted outside of Terraform.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error getting the group", err.Error())
		return
	}

	// Convert remote permissions to model
	remotePermissions := ConvertScimGroupPartialPermissionDataToModel(retrievedGroup.Permissions)

	// Find the intersection - only the permissions we're managing
	relevantPermissions := helper.IntersectBy(
		state.GroupPermissions,
		remotePermissions,
		CompareScimGroupPartialPermissions,
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
	state.ID = types.Int64Value(int64(*retrievedGroup.ID))
	state.GroupID = types.Int64Value(int64(*retrievedGroup.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *scimGroupPartialPermissionsResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state ScimGroupPartialPermissionsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := int(state.GroupID.ValueInt64())
	retrievedGroup, err := r.client.GetGroup(groupID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving group",
			fmt.Sprintf("Could not retrieve group %d: %s", groupID, err.Error()),
		)
		return
	}

	// Get permissions from various sources
	statePermissions := state.GroupPermissions
	planPermissions := plan.GroupPermissions
	remotePermissions := ConvertScimGroupPartialPermissionDataToModel(retrievedGroup.Permissions)

	// Calculate what permissions were deleted and what were added
	deletedPermissions, newPermissions := helper.DifferenceBy(
		statePermissions,
		planPermissions,
		CompareScimGroupPartialPermissions,
	)

	// Calculate the final set of permissions
	// Start with remote permissions, add new ones, remove deleted ones
	requiredAllPermissions, _ := helper.DifferenceBy(
		helper.UnionBy(remotePermissions, newPermissions, CompareScimGroupPartialPermissions),
		deletedPermissions,
		CompareScimGroupPartialPermissions,
	)

	tflog.Info(
		ctx,
		"UPDATE - Permission changes",
		map[string]any{
			"Deleted Permissions":           fmt.Sprintf("%+v", deletedPermissions),
			"New Permissions":               fmt.Sprintf("%+v", newPermissions),
			"Required all Permissions":      fmt.Sprintf("%+v", requiredAllPermissions),
			"Remote Permissions":            fmt.Sprintf("%+v", remotePermissions),
			"Remote + New Permissions":      fmt.Sprintf("%+v", helper.UnionBy(remotePermissions, newPermissions, CompareScimGroupPartialPermissions)),
		},
	)

	// Only update if there are actual changes
	if len(deletedPermissions) > 0 || len(newPermissions) > 0 {
		allPermissionsRequest := ConvertScimGroupPartialPermissionModelToData(
			requiredAllPermissions,
			groupID,
			retrievedGroup.AccountID,
		)

		_, err = r.client.UpdateGroupPermissions(groupID, allPermissionsRequest)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating group permissions",
				"Could not update partial permissions, unexpected error: "+err.Error(),
			)
			return
		}
	}

	plan.ID = types.Int64Value(int64(groupID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *scimGroupPartialPermissionsResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state ScimGroupPartialPermissionsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := int(state.GroupID.ValueInt64())

	tflog.Info(ctx, fmt.Sprintf("Removing partial permissions from group %d (group will remain)", groupID))

	retrievedGroup, err := r.client.GetGroup(groupID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			// Group is already gone, nothing to do
			tflog.Info(ctx, fmt.Sprintf("Group %d not found, assuming already deleted", groupID))
			return
		}
		resp.Diagnostics.AddError(
			"Error retrieving group",
			fmt.Sprintf("Could not retrieve group %d: %s", groupID, err.Error()),
		)
		return
	}

	// Get current remote permissions
	remotePermissions := ConvertScimGroupPartialPermissionDataToModel(retrievedGroup.Permissions)

	// Calculate what permissions should remain (remote minus our state permissions)
	remainingPermissions, _ := helper.DifferenceBy(
		remotePermissions,
		state.GroupPermissions,
		CompareScimGroupPartialPermissions,
	)

	tflog.Info(
		ctx,
		"DELETE - Removing partial permissions",
		map[string]any{
			"Remote Permissions":    fmt.Sprintf("%+v", remotePermissions),
			"State Permissions":     fmt.Sprintf("%+v", state.GroupPermissions),
			"Remaining Permissions": fmt.Sprintf("%+v", remainingPermissions),
		},
	)

	// Update the group with only the remaining permissions
	allPermissionsRequest := ConvertScimGroupPartialPermissionModelToData(
		remainingPermissions,
		groupID,
		retrievedGroup.AccountID,
	)

	_, err = r.client.UpdateGroupPermissions(groupID, allPermissionsRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error removing partial permissions from group",
			"Could not remove partial permissions, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Successfully removed partial permissions from group %d. Group still exists with remaining permissions.", groupID))
}

func (r *scimGroupPartialPermissionsResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}
