package profile

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &profilesDataSource{}
	_ datasource.DataSourceWithConfigure = &profilesDataSource{}
)

func ProfilesDataSource() datasource.DataSource {
	return &profilesDataSource{}
}

type profilesDataSource struct {
	client *dbt_cloud.Client
}

func (d *profilesDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*dbt_cloud.Client)
}

func (d *profilesDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_profiles"
}

func (d *profilesDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = dataSourceAllSchema
}

func (d *profilesDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var config ProfilesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(config.ProjectID.ValueInt64())

	profiles, err := d.client.GetAllProfiles(projectID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving profiles",
			err.Error(),
		)
		return
	}

	state := config
	allProfiles := []ProfileDataSourceModel{}

	for _, p := range profiles {
		current := ProfileDataSourceModel{}
		current.ID = types.StringValue(fmt.Sprintf(
			"%d%s%d",
			p.ProjectID,
			dbt_cloud.ID_DELIMITER,
			*p.ID,
		))
		current.ProfileID = types.Int64Value(int64(*p.ID))
		current.ProjectID = types.Int64Value(int64(p.ProjectID))
		current.Key = types.StringValue(p.Key)
		current.ConnectionID = types.Int64Value(int64(p.ConnectionID))
		current.CredentialsID = types.Int64Value(int64(p.CredentialsID))

		if p.ExtendedAttributesID != nil {
			current.ExtendedAttributesID = types.Int64Value(int64(*p.ExtendedAttributesID))
		} else {
			current.ExtendedAttributesID = types.Int64Null()
		}

		allProfiles = append(allProfiles, current)
	}
	state.Profiles = allProfiles

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
