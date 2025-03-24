package user_groups

import (
	"context"
	"fmt"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &userGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &userGroupDataSource{}
)

func UserGroupDataSource() datasource.DataSource {
	return &userGroupDataSource{}
}


type userGroupDataSource struct {
	client *dbt_cloud.Client
}

func (d *userGroupDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_user_groups"
}

func (d *userGroupDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*dbt_cloud.Client)
}

func (d *userGroupDataSource) Schema(
	ctx context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = datasourceSchema
}

func (d *userGroupDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data UserGroupsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := data.UserID.ValueInt64()
	retrievedUserGroups, err := d.client.GetUserGroups(int(userID))

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

	data.ID = types.StringValue(fmt.Sprintf("%d", userID))
	data.UserID = types.Int64Value(userID)

	groupIDs := []int{}
	for _, group := range retrievedUserGroups.Groups {
		groupIDs = append(groupIDs, *group.ID)
	}

	groupIDsSet, _ := types.SetValueFrom(ctx, types.Int64Type, groupIDs)
	if resp.Diagnostics.HasError() {
		return
	}

	data.GroupIDs = groupIDsSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}