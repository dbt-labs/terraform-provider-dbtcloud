package synapse_credential

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &synapseCredentialDataSource{}
	_ datasource.DataSourceWithConfigure = &synapseCredentialDataSource{}
)

// SynapseCredentialDataSource is a helper function to simplify the provider implementation.
func SynapseCredentialDataSource() datasource.DataSource {
	return &synapseCredentialDataSource{}
}

// synapseCredentialDataSource is the data source implementation.
type synapseCredentialDataSource struct {
	client *dbt_cloud.Client
}

// Configure adds the provider configured client to the data source.
func (d *synapseCredentialDataSource) Configure(
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
func (d *synapseCredentialDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_synapse_credential"
}

// Schema defines the schema for the data source.
func (d *synapseCredentialDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = datasourceSchema
}

// Read refreshes the Terraform state with the latest data.
func (d *synapseCredentialDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state SynapseCredentialDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	credential, err := d.client.GetSynapseCredential(projectID, credentialID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Synapse credential",
			"Could not read Synapse credential ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%d:%d", projectID, *credential.ID))
	state.ProjectID = types.Int64Value(int64(projectID))
	state.CredentialID = types.Int64Value(int64(*credential.ID))
	state.Authentication = types.StringValue(credential.UnencryptedCredentialDetails.Authentication)
	state.User = types.StringValue(credential.UnencryptedCredentialDetails.User)
	state.Schema = types.StringValue(credential.UnencryptedCredentialDetails.Schema)
	state.SchemaAuthorization = types.StringValue(credential.UnencryptedCredentialDetails.SchemaAuthorization)
	state.TenantId = types.StringValue(credential.UnencryptedCredentialDetails.TenantId)
	state.ClientId = types.StringValue(credential.UnencryptedCredentialDetails.ClientId)
	state.AdapterType = types.StringValue("synapse")

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
