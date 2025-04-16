package project

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &projectsDataSource{}
	_ datasource.DataSourceWithConfigure = &projectsDataSource{}
)

func ProjectsDataSource() datasource.DataSource {
	return &projectsDataSource{}
}

type projectsDataSource struct {
	client *dbt_cloud.Client
}

func (d *projectsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

func (d *projectsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var config ProjectsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	projectNameContains := config.NameContains.ValueString()

	apiProjects, err := d.client.GetAllProjects(projectNameContains)

	if err != nil {
		resp.Diagnostics.AddError(
			"Issue when retrieving projects",
			err.Error(),
		)
		return
	}

	state := config

	allProjects := []ProjectConnectionRepository{}
	for _, project := range apiProjects {

		currentProject := ProjectConnectionRepository{}
		currentProject.ID = types.Int64Value(project.ID)
		currentProject.Name = types.StringValue(project.Name)
		currentProject.Description = types.StringValue(project.Description)
		currentProject.SemanticLayerConfigID = types.Int64PointerValue(
			project.SemanticLayerConfigID,
		)
		currentProject.DbtProjectSubdirectory = types.StringValue(
			project.DbtProjectSubdirectory,
		)
		currentProject.CreatedAt = types.StringValue(project.CreatedAt)
		currentProject.UpdatedAt = types.StringValue(project.UpdatedAt)

		if project.Connection != nil {
			currentProject.Connection = &ProjectConnection{
				ID:             types.Int64PointerValue(project.Connection.ID),
				Name:           types.StringPointerValue(project.Connection.Name),
				AdapterVersion: types.StringPointerValue(project.Connection.AdapterVersion),
			}
		}

		if project.Repository != nil {
			currentProject.Repository = &ProjectRepository{
				ID: types.Int64PointerValue(
					helper.IntPointerToInt64Pointer(project.Repository.ID),
				),
				RemoteUrl: types.StringValue(project.Repository.RemoteUrl),
				PullRequestURLTemplate: types.StringValue(
					project.Repository.PullRequestURLTemplate,
				),
			}
		}

		allProjects = append(allProjects, currentProject)
	}
	state.Projects = allProjects

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *projectsDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*dbt_cloud.Client)
}

func (d *projectsDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = datasourceSchema
}
