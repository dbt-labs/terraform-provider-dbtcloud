package user_groups

import (
	"context"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func (u *userGroupsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	u.client = req.ProviderData.(*dbt_cloud.Client)
}

func (u *userGroupsResource) Create(_ context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	panic("unimplemented")
}

func (u *userGroupsResource) Delete(_ context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	panic("unimplemented")
}

func (u *userGroupsResource) ImportState(_ context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	panic("unimplemented")
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

	state.ID = types.Int64Value(userID)
	state.UserID = types.Int64Value(userID)

	groupIDs := []int{}
	
	for _, group := range retrievedUserGroups.Groups {
		groupIDs = append(groupIDs, *group.ID)
	}

	state.GroupIDs, _ = types.SetValueFrom(
		context.Background(),
		types.Int64Type,
		groupIDs,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (u *userGroupsResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema
}

func (u *userGroupsResource) Update(_ context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	panic("unimplemented")
}

func (u *userGroupsResource) Metadata(_ context.Context,req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_groups"
}
