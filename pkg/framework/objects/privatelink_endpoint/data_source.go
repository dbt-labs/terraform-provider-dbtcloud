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

	projectId := int(state.ProjectID.ValueInt64())
	credentialId := int(state.CredentialID.ValueInt64())

	credential, err := p.client.GetPostgresCredential(projectId, credentialId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Postgres credential",
			"Could not read Postgres credential ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%d%s%d", credential.Project_Id, dbt_cloud.ID_DELIMITER, *credential.ID))
	state.ProjectID = types.Int64Value(int64(credential.Project_Id))
	state.CredentialID = types.Int64Value(int64(*credential.ID))
	state.IsActive = types.BoolValue(credential.State == dbt_cloud.STATE_ACTIVE)
	state.DefaultSchema = types.StringValue(credential.Default_Schema)
	state.Username = types.StringValue(credential.Username)
	state.NumThreads = types.Int64Value(int64(credential.Threads))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (p *privatelinkEndpointDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceSchema
}
