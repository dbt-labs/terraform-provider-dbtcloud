package azure_dev_ops_repository

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type azureDevOpsRepositoryDataSource struct {
	client *dbt_cloud.Client
}

func AzureDevOpsRepositoryDataSource() datasource.DataSource {
	return &azureDevOpsRepositoryDataSource{}
}

func (d *azureDevOpsRepositoryDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_azure_dev_ops_repository"
}

func (d *azureDevOpsRepositoryDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state AzureDevopsRepositoryDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	repositoryName := state.Name.ValueString()
	azureDevOpsProjectID := state.AzureDevOpsProjectID.ValueString()

	azureDevOpsRepository, err := d.client.GetAzureDevOpsRepository(repositoryName, azureDevOpsProjectID)

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to get Azure Dev Ops repository %s in project %s", repositoryName, azureDevOpsProjectID),
			err.Error(),
		)
	}

	state.ID = types.StringValue(azureDevOpsRepository.ID)
	state.Name = types.StringValue(azureDevOpsRepository.Name)
	state.AzureDevOpsProjectID = types.StringValue(azureDevOpsProjectID)
	state.DetailsURL = types.StringValue(azureDevOpsRepository.DetailsURL)
	state.RemoteURL = types.StringValue(azureDevOpsRepository.RemoteURL)
	state.WebURL = types.StringValue(azureDevOpsRepository.WebURL)
	state.DefaultBranch = types.StringValue(azureDevOpsRepository.DefaultBranch)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *azureDevOpsRepositoryDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	switch c := req.ProviderData.(type) {
	case nil: // do nothing
	case *dbt_cloud.Client:
		d.client = c
	default:
		resp.Diagnostics.AddError(
			"Missing client",
			"A client is required to configure the Azure DevOps repository data source",
		)
	}
}
