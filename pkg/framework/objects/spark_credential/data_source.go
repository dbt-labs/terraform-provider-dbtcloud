package spark_credential

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &sparkCredentialDataSource{}
	_ datasource.DataSourceWithConfigure = &sparkCredentialDataSource{}
)

func SparkCredentialDataSource() datasource.DataSource {
	return &sparkCredentialDataSource{}
}

type sparkCredentialDataSource struct {
	client *dbt_cloud.Client
}

func (d *sparkCredentialDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sparkCredentialDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_spark_credential"
}

func (d *sparkCredentialDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state SparkCredentialDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	credential, err := d.client.GetSparkCredential(projectID, credentialID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Apache Spark credential", "Could not read Apache Spark credential ID "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%d%s%d", credential.Project_Id, dbt_cloud.ID_DELIMITER, *credential.ID))
	state.NumThreads = types.Int64Value(int64(credential.Threads))
	state.ProjectID = types.Int64Value(int64(credential.Project_Id))
	state.TargetName = types.StringValue(credential.Target_Name)
	state.Schema = types.StringValue(credential.UnencryptedCredentialDetails.Schema)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *sparkCredentialDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataSourceSchema
}
