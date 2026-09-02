package global_connection

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &globalConnectionDataSource{}
	_ datasource.DataSourceWithConfigure = &globalConnectionDataSource{}
)

func GlobalConnectionDataSource() datasource.DataSource {
	return &globalConnectionDataSource{}
}

type globalConnectionDataSource struct {
	client *dbt_cloud.Client
}

func (d *globalConnectionDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_global_connection"
}

func (d *globalConnectionDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state GlobalConnectionResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	connectionID := state.ID.ValueInt64()

	globalConnectionResponse, err := d.client.GetGlobalConnectionAdapter(connectionID)
	if err != nil {
		resp.Diagnostics.AddError("Error getting the connection type", err.Error())
		return
	}

	newState, action, err := readGeneric(
		d.client,
		&state,
		globalConnectionResponse.Data.AdapterVersion,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error reading the connection", err.Error())
		return
	}

	if action == "removeFromState" {
		resp.Diagnostics.AddWarning(
			"Resource not found",
			"The connection resource was not found and has been removed from the state.",
		)
		resp.State.RemoveResource(ctx)
		return
	}

	// readGeneric can't infer this one: the API only reports the adapter version, so
	// the resource carries `use_latest_adapter` over from its own state instead
	if newState.BigQueryConfig != nil {
		newState.BigQueryConfig.UseLatestAdapter = types.BoolValue(
			globalConnectionResponse.Data.AdapterVersion == dbt_cloud.BigQueryConfig{}.LatestAdapterVersion(),
		)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (d *globalConnectionDataSource) Configure(
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
