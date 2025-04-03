package job

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &jobDataSource{}
	_ datasource.DataSourceWithConfigure = &jobDataSource{}
)

func JobDataSource() datasource.DataSource {
	return &jobDataSource{}
}

type jobDataSource struct {
	client *dbt_cloud.Client
}


func (j *jobDataSource) Configure(_ context.Context, req datasource.ConfigureRequest,resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	j.client = req.ProviderData.(*dbt_cloud.Client)
}



func (j *jobDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job"
}


func (j *jobDataSource) Read(ctx context.Context,req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// var state JobDataSourceModel

	// resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	// jobId := state.JobId.ValueInt64()

	// _, err := j.client.GetJob(strconv.FormatInt(jobId, 10))

	// if err != nil {
	// 	resp.Diagnostics.AddError("Error getting the job", err.Error())
	// 	return
	// }

}