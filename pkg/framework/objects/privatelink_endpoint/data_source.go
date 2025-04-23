package privatelink_endpoint

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &privatelinkEndpointDataSource{}
	_ datasource.DataSourceWithConfigure = &privatelinkEndpointDataSource{}
)

func PrivatelinkEndpointDataSource() datasource.DataSource {
	return &privatelinkEndpointDataSource{}
}

type privatelinkEndpointDataSource struct {
	client *dbt_cloud.Client
}

func (p *privatelinkEndpointDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*dbt_cloud.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *dbt_cloud.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	p.client = client
}

func (p *privatelinkEndpointDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_privatelink_endpoint"
}

func (p *privatelinkEndpointDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state PrivatelinkEndpointDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointName := state.Name.ValueString()
	endpointURL := state.PrivatelinkEndpointURL.ValueString()

	privatelinkEndpoint, err := p.client.GetPrivatelinkEndpoint(endpointName, endpointURL)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Privatelink Endpoint",
			"Could not read Privatelink Endpoint "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(privatelinkEndpoint.ID)
	state.Name = types.StringValue(privatelinkEndpoint.Name)
	state.PrivatelinkEndpointType = types.StringValue(privatelinkEndpoint.Type)
	state.PrivatelinkEndpointURL = types.StringValue(privatelinkEndpoint.PrivatelinkEndpointURL)
	state.CIDRRange = types.StringValue(privatelinkEndpoint.CIDRRange)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (p *privatelinkEndpointDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceSchema
}
