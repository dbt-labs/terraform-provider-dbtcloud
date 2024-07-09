package group

import (
	"context"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &groupDataSource{}
	_ datasource.DataSourceWithConfigure = &groupDataSource{}
)

func GroupDataSource() datasource.DataSource {
	return &groupDataSource{}
}

type groupDataSource struct {
	client *dbt_cloud.Client
}

func (d *groupDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (d *groupDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {

	var data GroupDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	groupID := data.GroupID.ValueInt64()
	retrievedGroup, err := d.client.GetGroup(int(groupID))

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

	data.GroupID = types.Int64Value(int64(*retrievedGroup.ID))
	data.ID = types.Int64Value(int64(*retrievedGroup.ID))
	data.Name = types.StringValue(retrievedGroup.Name)
	data.AssignByDefault = types.BoolValue(retrievedGroup.AssignByDefault)
	data.SSOMappingGroups, _ = types.SetValueFrom(
		context.Background(),
		types.StringType,
		retrievedGroup.SSOMappingGroups,
	)

	remotePermissions := ConvertGroupPermissionDataToModel(retrievedGroup.Permissions)
	data.GroupPermissions = remotePermissions

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *groupDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*dbt_cloud.Client)
}
