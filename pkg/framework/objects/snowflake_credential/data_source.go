package snowflake_credential

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &snowflakeCredentialDataSource{}
	_ datasource.DataSourceWithConfigure = &snowflakeCredentialDataSource{}
)

// SnowflakeCredentialDataSource is a helper function to simplify the provider implementation.
func SnowflakeCredentialDataSource() datasource.DataSource {
	return &snowflakeCredentialDataSource{}
}

// snowflakeCredentialDataSource is the data source implementation.
type snowflakeCredentialDataSource struct {
	client *dbt_cloud.Client
}

// Configure adds the provider configured client to the data source.
func (d *snowflakeCredentialDataSource) Configure(
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
func (d *snowflakeCredentialDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_snowflake_credential"
}

// Schema defines the schema for the data source.
func (d *snowflakeCredentialDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = datasourceSchema
}

// Read refreshes the Terraform state with the latest data.
func (d *snowflakeCredentialDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state SnowflakeCredentialDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	credential, err := d.client.GetSnowflakeCredential(projectID, credentialID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Snowflake credential",
			"Could not read Snowflake credential ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response body to model
	state.ID = types.StringValue(fmt.Sprintf("%d:%d", projectID, credentialID))
	state.Schema = types.StringValue(credential.Schema)
	state.IsActive = types.BoolValue(credential.State == dbt_cloud.STATE_ACTIVE)
	state.AuthType = types.StringValue(credential.Auth_Type)
	state.ProjectID = types.Int64Value(int64(projectID))
	state.CredentialID = types.Int64Value(int64(credentialID))
	state.User = types.StringValue(credential.User)
	state.NumThreads = types.Int64Value(int64(credential.Threads))

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
