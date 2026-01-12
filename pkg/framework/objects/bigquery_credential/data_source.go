package bigquery_credential

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &bigqueryCredentialDataSource{}
	_ datasource.DataSourceWithConfigure = &bigqueryCredentialDataSource{}
)

// BigqueryCredentialDataSource is a helper function to simplify the provider implementation.
func BigqueryCredentialDataSource() datasource.DataSource {
	return &bigqueryCredentialDataSource{}
}

// bigqueryCredentialDataSource is the data source implementation.
type bigqueryCredentialDataSource struct {
	client *dbt_cloud.Client
}

// Configure adds the provider configured client to the data source.
func (d *bigqueryCredentialDataSource) Configure(
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
func (d *bigqueryCredentialDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_bigquery_credential"
}

// Schema defines the schema for the data source.
func (d *bigqueryCredentialDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = datasourceSchema
}

// Read refreshes the Terraform state with the latest data.
func (d *bigqueryCredentialDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state BigqueryCredentialDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	credential, err := d.client.GetBigQueryCredential(projectID, credentialID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Bigquery credential",
			"Could not read Bigquery credential ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response body to model
	state.ID = types.StringValue(fmt.Sprintf("%d%s%d", credential.Project_Id, dbt_cloud.ID_DELIMITER, *credential.ID))
	// Use helper methods to get dataset and threads from the correct location (v0 vs v1)
	state.Dataset = types.StringValue(credential.GetDataset())
	state.NumThreads = types.Int64Value(int64(credential.GetThreads()))
	state.IsActive = types.BoolValue(credential.State == dbt_cloud.STATE_ACTIVE)
	state.ProjectID = types.Int64Value(int64(credential.Project_Id))

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
