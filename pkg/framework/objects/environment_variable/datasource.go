package environment_variable

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &environmentVariableDataSource{}
	_ datasource.DataSourceWithConfigure = &environmentVariableDataSource{}
)

// EnvironmentVariableDataSource is a helper function to simplify the provider implementation.
func EnvironmentVariableDataSource() datasource.DataSource {
	return &environmentVariableDataSource{}
}

// environmentVariableDataSource is the data source implementation.
type environmentVariableDataSource struct {
	client *dbt_cloud.Client
}

// Configure adds the provider configured client to the data source.
func (d *environmentVariableDataSource) Configure(
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
func (d *environmentVariableDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_environment_variable"
}

// Schema defines the schema for the data source.
func (d *environmentVariableDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = datasourceSchema
}

// Read refreshes the Terraform state with the latest data.
func (d *environmentVariableDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state EnvironmentVariableDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	name := state.Name.ValueString()

	envVar, err := d.client.GetEnvironmentVariable(projectID, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading environment variable",
			"Could not read environment variable ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response body to model
	state.ID = types.StringValue(fmt.Sprintf("%d:%s", envVar.ProjectID, envVar.Name))
	state.Name = types.StringValue(envVar.Name)

	envVarElements := make(map[string]attr.Value)
	for key, value := range envVar.EnvironmentNameValues {
		envVarElements[key] = types.StringValue(value.Value)
	}

	envVarMap, diag := types.MapValue(types.StringType, envVarElements)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.EnvironmentValues = envVarMap

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
