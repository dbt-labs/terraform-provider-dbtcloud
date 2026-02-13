package profile

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &profileDataSource{}
	_ datasource.DataSourceWithConfigure = &profileDataSource{}
)

func ProfileDataSource() datasource.DataSource {
	return &profileDataSource{}
}

type profileDataSource struct {
	client *dbt_cloud.Client
}

func (d *profileDataSource) Configure(
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

func (d *profileDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_profile"
}

func (d *profileDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = dataSourceSchema
}

func (d *profileDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state ProfileDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	profileID := int(state.ProfileID.ValueInt64())

	profile, err := d.client.GetProfile(projectID, profileID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading profile",
			"Could not read profile: "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(fmt.Sprintf(
		"%d%s%d",
		profile.ProjectID,
		dbt_cloud.ID_DELIMITER,
		*profile.ID,
	))
	state.ProfileID = types.Int64Value(int64(*profile.ID))
	state.ProjectID = types.Int64Value(int64(profile.ProjectID))
	state.Key = types.StringValue(profile.Key)
	state.ConnectionID = types.Int64Value(int64(profile.ConnectionID))
	state.CredentialsID = types.Int64Value(int64(profile.CredentialsID))

	if profile.ExtendedAttributesID != nil {
		state.ExtendedAttributesID = types.Int64Value(int64(*profile.ExtendedAttributesID))
	} else {
		state.ExtendedAttributesID = types.Int64Null()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
