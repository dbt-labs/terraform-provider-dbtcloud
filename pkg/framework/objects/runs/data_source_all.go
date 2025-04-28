package runs

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &runsDataSource{}
	_ datasource.DataSourceWithConfigure = &runsDataSource{}
)

func RunsDataSource() datasource.DataSource {
	return &runsDataSource{}
}

type runsDataSource struct {
	client *dbt_cloud.Client
}

func (d *runsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_runs"
}

func (d *runsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state RunsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	filter := dbt_cloud.RunFilter{}

	if state.Filter != (RunFilterModel{}) {
		filter.EnvironmentID = int(state.Filter.EnvironmentID.ValueInt64())
		filter.Status = int(state.Filter.Status.ValueInt64())
		filter.ProjectID = int(state.Filter.ProjectID.ValueInt64())
		filter.JobDefinitionID = int(state.Filter.JobDefinitionID.ValueInt64())
		filter.PullRequestID = int(state.Filter.PullRequestID.ValueInt64())
		filter.TriggerID = int(state.Filter.TriggerID.ValueInt64())
		filter.Limit = int(state.Filter.Limit.ValueInt64())
	}

	runs, err := d.client.GetRuns(&filter)

	if err != nil {
		resp.Diagnostics.AddError(
			"Issue when retrieving runs",
			err.Error(),
		)
		return
	}

	for _, run := range *runs {
		state.Runs = append(state.Runs, RunDataSourceModel{
			ID:                  types.Int64Value(run.ID),
			AccountID:           types.Int64Value(run.AccountID),
			GitSHA:              types.StringValue(run.GitSHA),
			GitBranch:           types.StringValue(run.GitBranch),
			GitHubPullRequestID: types.StringValue(run.GitHubPullRequestID),
			SchemaOverride:      types.StringValue(run.SchemaOverride),
			Cause:               types.StringValue(run.Cause),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *runsDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*dbt_cloud.Client)
}

func (d *runsDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = allDatasourceSchema
}
