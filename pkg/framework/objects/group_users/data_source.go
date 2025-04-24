package group_users

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &groupUsersDataSource{}
	_ datasource.DataSourceWithConfigure = &groupUsersDataSource{}
)

// GroupUsersDataSource is a helper function to simplify the provider implementation.
func GroupUsersDataSource() datasource.DataSource {
	return &groupUsersDataSource{}
}

// groupUsersDataSource is the data source implementation.
type groupUsersDataSource struct {
	client *dbt_cloud.Client
}

// Configure adds the provider configured client to the data source.
func (d *groupUsersDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*dbt_cloud.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf(
				"Expected *dbt_cloud.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}

	d.client = client
}

// Metadata returns the data source type name.
func (d *groupUsersDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_group_users"
}

// Schema defines the schema for the data source.
func (d *groupUsersDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = dataSourceSchema
}

// Read refreshes the Terraform state with the latest data.
func (d *groupUsersDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state GroupUsersDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := int(state.GroupID.ValueInt64())

	users, err := d.client.GetUsers()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading users",
			"Could not read users "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	usersModels := []userDataSourceModel{}
	for _, user := range users {
		userGroups := user.Permissions[0].Groups

		userInGroup := false
		for _, userGroup := range userGroups {
			if userGroup.ID == groupID {
				userInGroup = true
				// we can stop looping
				break
			}
		}

		if userInGroup {
			userModel := userDataSourceModel{
				ID:    types.Int64Value(int64(user.ID)),
				Email: types.StringValue(user.Email),
			}
			usersModels = append(usersModels, userModel)
		}
	}

	state.Users = usersModels

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
