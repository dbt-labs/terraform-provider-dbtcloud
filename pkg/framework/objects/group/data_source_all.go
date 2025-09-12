package group

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &groupsDataSource{}
	_ datasource.DataSourceWithConfigure = &groupsDataSource{}
)

func GroupsDataSource() datasource.DataSource {
	return &groupsDataSource{}
}

type groupsDataSource struct {
	client *dbt_cloud.Client
}

func (d *groupsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (d *groupsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var config GroupsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	name := config.Name.ValueString()
	nameContains := config.NameContains.ValueString()
	stateFilter := config.State.ValueString()

	apiGroups, err := d.client.GetAllGroups(name, nameContains, stateFilter)

	if err != nil {
		resp.Diagnostics.AddError(
			"Issue when retrieving groups",
			err.Error(),
		)
		return
	}

	state := config

	allGroups := []GroupInfo{}
	for _, group := range apiGroups {
		currentGroup := GroupInfo{}
		currentGroup.ID = types.Int64Value(int64(*group.ID))
		currentGroup.Name = types.StringValue(group.Name)
		currentGroup.State = types.Int64Value(int64(group.State))
		currentGroup.AssignByDefault = types.BoolValue(group.AssignByDefault)
		currentGroup.ScimManaged = types.BoolValue(group.ScimManaged)
		currentGroup.SSOMappingGroups, _ = types.SetValueFrom(
			context.Background(),
			types.StringType,
			group.SSOMappingGroups,
		)

		allGroups = append(allGroups, currentGroup)
	}
	state.Groups = allGroups

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *groupsDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*dbt_cloud.Client)
}

func (d *groupsDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = groupsDataSourceSchema
}
