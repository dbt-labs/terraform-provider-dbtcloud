package clickhouse_credential

import (
	"context"
	"fmt"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &clickhouseCredentialDataSource{}
	_ datasource.DataSourceWithConfigure = &clickhouseCredentialDataSource{}
)

func ClickhouseCredentialDataSource() datasource.DataSource {
	return &clickhouseCredentialDataSource{}
}

type clickhouseCredentialDataSource struct {
	client *dbt_cloud.Client
}

func (d *clickhouseCredentialDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_clickhouse_credential"
}

func (d *clickhouseCredentialDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = dataSourceSchema
}

func (d *clickhouseCredentialDataSource) Configure(
	_ context.Context,
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

func (d *clickhouseCredentialDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var config ClickhouseCredentialDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(config.ProjectID.ValueInt64())
	credentialID := int(config.CredentialID.ValueInt64())

	credential, err := d.client.GetClickhouseCredential(projectID, credentialID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddError(
				"ClickHouse credential not found",
				fmt.Sprintf("Could not find ClickHouse credential with ID %d in project %d", credentialID, projectID),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading ClickHouse credential",
			fmt.Sprintf("Could not read ClickHouse credential ID %d in project %d: %s", credentialID, projectID, err.Error()),
		)
		return
	}

	config.ID = types.StringValue(fmt.Sprintf("%d:%d", projectID, *credential.ID))
	config.User = types.StringValue(credential.UnencryptedCredentialDetails.User)
	config.Schema = types.StringValue(credential.UnencryptedCredentialDetails.Schema)
	config.TargetName = types.StringValue(credential.UnencryptedCredentialDetails.TargetName)
	config.Threads = threadsInt64Value(credential.UnencryptedCredentialDetails.Threads)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
