package global_connection

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &globalConnectionsDataSource{}
	_ datasource.DataSourceWithConfigure = &globalConnectionsDataSource{}
)

func GlobalConnectionsDataSource() datasource.DataSource {
	return &globalConnectionsDataSource{}
}

type globalConnectionsDataSource struct {
	client *dbt_cloud.Client
}

func (d *globalConnectionsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_global_connections"
}

func (d *globalConnectionsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state GlobalConnectionsDatasourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	apiAllConnections, err := d.client.GetAllConnections()
	if err != nil {
		resp.Diagnostics.AddError(
			"Issue when retrieving connections",
			err.Error(),
		)
		return
	}

	allConnections := []GlobalConnectionSummary{}
	for _, connection := range apiAllConnections {

		currentConnection := GlobalConnectionSummary{}
		currentConnection.ID = types.Int64Value(connection.ID)
		currentConnection.Name = types.StringValue(connection.Name)
		currentConnection.CreatedAt = types.StringValue(connection.CreatedAt)
		currentConnection.UpdatedAt = types.StringValue(connection.UpdatedAt)
		currentConnection.AdapterVersion = types.StringValue(connection.AdapterVersion)
		currentConnection.PrivateLinkEndpointID = types.StringPointerValue(
			connection.PrivateLinkEndpointID,
		)
		currentConnection.IsSSHTunnelEnabled = types.BoolValue(connection.IsSSHTunnelEnabled)
		currentConnection.OauthConfigurationID = types.Int64PointerValue(
			connection.OauthConfigurationID,
		)
		currentConnection.EnvironmentCount = types.Int64Value(connection.EnvironmentCount)

		allConnections = append(allConnections, currentConnection)
	}
	state.Connections = allConnections

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (d *globalConnectionsDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	switch c := req.ProviderData.(type) {
	case nil: // do nothing
	case *dbt_cloud.Client:
		d.client = c
	default:
		resp.Diagnostics.AddError("Missing client", "A client is required to configure the global connection resource")
	}
}
