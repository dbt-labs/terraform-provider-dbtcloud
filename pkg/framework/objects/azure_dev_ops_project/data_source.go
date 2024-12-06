package azure_dev_ops_project

import (
	"context"
	"fmt"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &azureDevOpsProjectDataSource{}
	_ datasource.DataSourceWithConfigure = &azureDevOpsProjectDataSource{}
)

func AzureDevOpsProjectDataSource() datasource.DataSource {
	return &azureDevOpsProjectDataSource{}
}

type azureDevOpsProjectDataSource struct {
	client *dbt_cloud.Client
}

func (d *azureDevOpsProjectDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_azure_dev_ops_project"
}

func (d *azureDevOpsProjectDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state AzureDevOpsProjectDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	projectName := state.Name.ValueString()

	azureDevOpsProject, err := d.client.GetAzureDevOpsProject(projectName)

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Did not find Azure DevOps Project with name: %s", state.Name),
			err.Error(),
		)
		return
	}

	state.Name = types.StringValue(azureDevOpsProject.Name)
	state.ID = types.StringValue(azureDevOpsProject.ID)
	state.URL = types.StringValue(azureDevOpsProject.URL)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *azureDevOpsProjectDataSource) Configure(
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
			"A client is required to configure the Azure DevOps Project data source",
		)
	}
}
