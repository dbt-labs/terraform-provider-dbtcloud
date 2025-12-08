package group

import (
	"context"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

var (
	_ resource.Resource                = &groupResource{}
	_ resource.ResourceWithConfigure   = &groupResource{}
	_ resource.ResourceWithImportState = &groupResource{}
)

func GroupResource() resource.Resource {
	return &groupResource{}
}

type groupResource struct {
	client *dbt_cloud.Client
}

func (r *groupResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *groupResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state GroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	groupID := state.ID.ValueInt64()
	retrievedGroup, err := r.client.GetGroup(int(groupID))

	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The group was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error getting the group", err.Error())
		return
	}

	state.ID = types.Int64Value(int64(*retrievedGroup.ID))
	state.Name = types.StringValue(retrievedGroup.Name)
	state.AssignByDefault = types.BoolValue(retrievedGroup.AssignByDefault)

	stateSSOMappingGroups, diags := types.SetValueFrom(
		context.Background(),
		types.StringType,
		retrievedGroup.SSOMappingGroups,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags.Errors()...)
		return
	}

	state.SSOMappingGroups = stateSSOMappingGroups

	remotePermissions := ConvertGroupPermissionDataToModel(retrievedGroup.Permissions)
	state.GroupPermissions = remotePermissions

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (r *groupResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan GroupResourceModel

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

	createdGroup, err := r.client.CreateGroup(name, assignByDefault, ssoMappingGroups)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create group",
			"Error: "+err.Error(),
		)
		return
	}

	groupPermissions := ConvertGroupPermissionModelToData(
		plan.GroupPermissions,
		*createdGroup.ID,
		createdGroup.AccountID,
	)

	_, err = r.client.UpdateGroupPermissions(*createdGroup.ID, groupPermissions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to assign permissions to the group",
			"Error: "+err.Error(),
		)

		// TODO: Delete the group if the permissions update fails
		return
	}

	plan.ID = types.Int64Value(int64(*createdGroup.ID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *groupResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state GroupResourceModel

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
	}
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

func (r *groupResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state GroupResourceModel

	// Read plan and state values into the models
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
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

	// we check the group data
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

	// we check the permission data
	statePermissions := state.GroupPermissions
	planPermissions := plan.GroupPermissions

	diff1, diff2 := helper.DifferenceBy(
		statePermissions,
		planPermissions,
		CompareGroupPermissions)
	// if there is any difference, we update the group permissions
	if len(diff1) > 0 || len(diff2) > 0 {
		groupPermissions := ConvertGroupPermissionModelToData(
			planPermissions,
			groupID,
			retrievedGroup.AccountID,
		)

		_, err = r.client.UpdateGroupPermissions(groupID, groupPermissions)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update group permissions",
				"Error: "+err.Error(),
			)
		}
		state.GroupPermissions = plan.GroupPermissions
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *groupResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {

	// I think we need this conversion because the ID is a string
	groupIDStr := req.ID
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing group ID for import", err.Error())
		return
	}

	// and for some arcane reason, we need to initiate the SSO Set to its type
	// ¯\_(ツ)_/¯
	ssoSetVal, _ := types.SetValue(types.StringType, nil)
	state := GroupResourceModel{
		ID:               types.Int64Value(int64(groupID)),
		SSOMappingGroups: ssoSetVal,
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *groupResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}
