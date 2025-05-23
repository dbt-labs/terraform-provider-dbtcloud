package redshift_credential

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &redshiftCredentialDataSource{}
	_ datasource.DataSourceWithConfigure = &redshiftCredentialDataSource{}
)

// RedshiftCredentialDataSource is a helper function to simplify the provider implementation.
func RedshiftCredentialDataSource() datasource.DataSource {
	return &redshiftCredentialDataSource{}
}

// redshiftCredentialDataSource is the data source implementation.
type redshiftCredentialDataSource struct {
	client *dbt_cloud.Client
}

// Configure adds the provider configured client to the data source.
func (d *redshiftCredentialDataSource) Configure(
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
func (d *redshiftCredentialDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_redshift_credential"
}

// Schema defines the schema for the data source.
func (d *redshiftCredentialDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = RedshiftDatasourceSchema
}

// Read refreshes the Terraform state with the latest data.
func (d *redshiftCredentialDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state RedshiftCredentialDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	credential, err := d.client.GetRedshiftCredential(projectID, credentialID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Redshift credential",
			"Could not read Redshift credential ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response body to model
	state.ID = types.StringValue(fmt.Sprintf("%d%s%d", credential.Project_Id, dbt_cloud.ID_DELIMITER, *credential.ID))
	state.Schema = types.StringValue(credential.Schema)
	state.NumThreads = types.Int64Value(int64(credential.Threads))
	state.IsActive = types.BoolValue(credential.State == dbt_cloud.STATE_ACTIVE)
	state.ProjectID = types.Int64Value(int64(credential.Project_Id))

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
