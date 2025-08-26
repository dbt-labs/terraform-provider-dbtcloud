package privatelink_endpoint

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &privatelinkEndpointDataSourceAll{}
	_ datasource.DataSourceWithConfigure = &privatelinkEndpointDataSourceAll{}
)

func PrivatelinkEndpointDataSourceAll() datasource.DataSource {
	return &privatelinkEndpointDataSourceAll{}
}

type privatelinkEndpointDataSourceAll struct {
	client *dbt_cloud.Client
}

func (p *privatelinkEndpointDataSourceAll) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (p *privatelinkEndpointDataSourceAll) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_privatelink_endpoints"
}

func (p *privatelinkEndpointDataSourceAll) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state PrivatelinkEndpointDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	privatelinkEndpoint, err := p.client.GetAllPrivatelinkEndpoints()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Privatelink Endpoints",
			"Could not read Privatelink Endpoints: "+err.Error(),
		)
		return
	}

	for _, endpoint := range privatelinkEndpoint {
		state.ID = types.StringValue(endpoint.ID)
		state.Name = types.StringValue(endpoint.Name)
		state.PrivatelinkEndpointType = types.StringValue(endpoint.Type)
		state.PrivatelinkEndpointURL = types.StringValue(endpoint.PrivatelinkEndpointURL)
		state.CIDRRange = types.StringValue(endpoint.CIDRRange)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
