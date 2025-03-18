package user_groups

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = &userGroupResource{}
	_ resource.ResourceWithConfigure   = &userGroupResource{}
	_ resource.ResourceWithImportState = &userGroupResource{}
)

func UserGroupResource() resource.Resource {
	return &userGroupResource{}
}

type userGroupResource struct {
	client *dbt_cloud.Client
}

func (u *userGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	u.client = req.ProviderData.(*dbt_cloud.Client)
}

func (u *userGroupResource) Create(_ context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	panic("unimplemented")
}

func (u *userGroupResource) Delete(_ context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	panic("unimplemented")
}

func (u *userGroupResource) ImportState(_ context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	panic("unimplemented")
}

func (u *userGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserGroupsResourceModel	

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	userGroupID := state.ID.ValueInt64()

	resp.Diagnostics.AddWarning(
		"user group id is",
		fmt.Sprintf("The user group id is %d", userGroupID),
	)

}

func (u *userGroupResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema
}

func (u *userGroupResource) Update(_ context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	panic("unimplemented")
}

func (u *userGroupResource) Metadata(_ context.Context,req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_groups"
}
