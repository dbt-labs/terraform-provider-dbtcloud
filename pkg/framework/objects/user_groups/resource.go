package user_groups

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &userGroupsResource{}
	_ resource.ResourceWithConfigure   = &userGroupsResource{}
	_ resource.ResourceWithImportState = &userGroupsResource{}
)

func UserGroupsResource() resource.Resource {
	return &userGroupsResource{}
}

type userGroupsResource struct {
	client *dbt_cloud.Client
}

func checkGroupsAssigned(groupIDs []int, groupsAssigned *dbt_cloud.AssignUserGroupsResponse) error {
	groupIDsAssignedMap := map[int]bool{}
	for _, group := range groupsAssigned.Data {
		groupIDsAssignedMap[*group.ID] = true
	}

	for _, groupID := range groupIDs {
		if !groupIDsAssignedMap[groupID] {
			return fmt.Errorf("the Group %d was not assigned to the user (it's possible that it doesn't exist and needs to be removed from the config)", groupID)
		}
	}

	return nil
}

func (u *userGroupsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	u.client = req.ProviderData.(*dbt_cloud.Client)
}

func (u *userGroupsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserGroupsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := plan.UserID.ValueInt64()
	var groupIDs []int

	resp.Diagnostics.Append(plan.GroupIDs.ElementsAs(ctx, &groupIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupsAssigned, err := u.client.AssignUserGroups(int(userID), groupIDs)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error assigning user groups",
			"Error: "+err.Error(),
		)
		return
	}

	if err := checkGroupsAssigned(groupIDs, groupsAssigned); err != nil {
		resp.Diagnostics.AddError(
			"Error validating group assignments",
			fmt.Sprintf("Not all groups were assigned to user %d: %s", userID, err),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%d", userID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (u *userGroupsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Warn(ctx, "[WARN] dbtcloud_user_groups does not support delete") 
}

func (u *userGroupsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	userIDStr := req.ID
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing user ID for import", err.Error())
		return
	}

	groupIDsSet, _ := types.SetValue(types.Int64Type, nil)
	state := UserGroupsResourceModel{
		ID:       types.StringValue(userIDStr),
		UserID:   types.Int64Value(int64(userID)),
		GroupIDs: groupIDsSet,
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (u *userGroupsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserGroupsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	userID := state.UserID.ValueInt64()
	retrievedUserGroups, err := u.client.GetUserGroups(int(userID))

	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The  was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error getting the user groups", err.Error())
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%d", userID))
	state.UserID = types.Int64Value(userID)

	groupIDs := []int{}

	for _, group := range retrievedUserGroups.Groups {
		groupIDs = append(groupIDs, *group.ID)
	}

	state.GroupIDs, _ = types.SetValueFrom(
		ctx,
		types.Int64Type,
		groupIDs,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (u *userGroupsResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema
}

func (u *userGroupsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan, state UserGroupsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	userIDChanged := !plan.UserID.Equal(state.UserID)
	groupIDsChanged := !plan.GroupIDs.Equal(state.GroupIDs)

	if userIDChanged || groupIDsChanged {
		userID := plan.UserID.ValueInt64()
		var groupIDs []int

		resp.Diagnostics.Append(plan.GroupIDs.ElementsAs(ctx, &groupIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		groupsAssigned, err := u.client.AssignUserGroups(int(userID), groupIDs)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error assigning user to groups",
				fmt.Sprintf("Unable to assign user %d to groups: %s", userID, err),
			)
			return
		}

		if err := checkGroupsAssigned(groupIDs, groupsAssigned); err != nil {
			resp.Diagnostics.AddError(
				"Error validating group assignments",
				fmt.Sprintf("Not all groups were assigned to user %d: %s", userID, err),
			)
			return
		}
		
		plan.ID = types.StringValue(fmt.Sprintf("%d", userID))
	}
	
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (u *userGroupsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_groups"
}
