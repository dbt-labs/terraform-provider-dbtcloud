package scim_group_permissions

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &scimGroupPermissionsResource{}
	_ resource.ResourceWithConfigure = &scimGroupPermissionsResource{}
)

func ScimGroupPermissionsResource() resource.Resource {
	return &scimGroupPermissionsResource{}
}

type scimGroupPermissionsResource struct {
	client *dbt_cloud.Client
}

func (r *scimGroupPermissionsResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_scim_group_permissions"
}

func (r *scimGroupPermissionsResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan ScimGroupPermissionsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := int(plan.GroupID.ValueInt64())

	// Verify the group exists and we're not trying to create it
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

	tflog.Info(ctx, fmt.Sprintf("Managing permissions for existing group: %s (ID: %d)", existingGroup.Name, *existingGroup.ID))

	// Convert plan permissions to API format
	allPermissions := ConvertScimGroupPermissionModelToData(
		plan.GroupPermissions,
		groupID,
		r.client.AccountID,
	)

	// Update the group permissions
	_, err = r.client.UpdateGroupPermissions(groupID, allPermissions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating group permissions",
			"Could not update group permissions, unexpected error: "+err.Error(),
		)
		return
	}

	// Set the state
	plan.ID = types.Int64Value(int64(groupID))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *scimGroupPermissionsResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state ScimGroupPermissionsResourceModel

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

	// Convert permissions from API to model
	remotePermissions := ConvertScimGroupPermissionDataToModel(retrievedGroup.Permissions)

	state.GroupPermissions = remotePermissions
	state.ID = types.Int64Value(int64(*retrievedGroup.ID))
	state.GroupID = types.Int64Value(int64(*retrievedGroup.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *scimGroupPermissionsResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan ScimGroupPermissionsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := int(plan.GroupID.ValueInt64())

	// Convert plan permissions to API format
	allPermissions := ConvertScimGroupPermissionModelToData(
		plan.GroupPermissions,
		groupID,
		r.client.AccountID,
	)

	// Update the group permissions
	_, err := r.client.UpdateGroupPermissions(groupID, allPermissions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating group permissions",
			"Could not update group permissions, unexpected error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *scimGroupPermissionsResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state ScimGroupPermissionsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := int(state.GroupID.ValueInt64())

	tflog.Info(ctx, fmt.Sprintf("Removing all permissions from group %d (but not deleting the group itself)", groupID))

	// Remove all permissions by setting an empty permissions array
	emptyPermissions := []dbt_cloud.GroupPermission{}
	_, err := r.client.UpdateGroupPermissions(groupID, emptyPermissions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error removing group permissions",
			"Could not remove group permissions, unexpected error: "+err.Error(),
		)
		return
	}

	// Note: We do NOT delete the group itself, only remove permissions
	tflog.Info(ctx, fmt.Sprintf("Successfully removed all permissions from group %d. Group still exists.", groupID))
}

func (r *scimGroupPermissionsResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}

func (r *scimGroupPermissionsResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	groupID, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing scim_group_permissions",
			"Could not parse group ID: "+err.Error(),
		)
		return
	}

	// Verify the group exists
	existingGroup, err := r.client.GetGroup(groupID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing scim_group_permissions",
			fmt.Sprintf("Could not retrieve group %d: %s", groupID, err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), int64(groupID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_id"), int64(groupID))...)

	tflog.Info(ctx, fmt.Sprintf("Imported scim_group_permissions for group: %s (ID: %d)", existingGroup.Name, *existingGroup.ID))
}
