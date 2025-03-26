package databricks_credentials

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &databricksCredentialDataSource{}
	_ datasource.DataSourceWithConfigure = &databricksCredentialDataSource{}
)

func DatabricksCredentialDataSource() datasource.DataSource {
	return &databricksCredentialDataSource{}
}

type databricksCredentialDataSource struct {
	client *dbt_cloud.Client
}

func (d *databricksCredentialDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = req.ProviderData.(*dbt_cloud.Client)
}

func (d *databricksCredentialDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databricks_credential"
}

func (d *databricksCredentialDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DatabricksCredentialDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	credential, err := d.client.GetDatabricksCredential(projectID, credentialID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Databricks credential", "Could not read Databricks credential ID "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%d%s%d", credential.Project_Id, dbt_cloud.ID_DELIMITER, *credential.ID))
	state.NumThreads = types.Int64Value(int64(credential.Threads))
	state.ProjectID = types.Int64Value(int64(credential.Project_Id))
	state.AdapterID = types.Int64Value(int64(credential.Adapter_Id))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *databricksCredentialDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataSourceSchema
}
