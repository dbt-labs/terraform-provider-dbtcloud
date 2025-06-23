package project

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &projectDataSource{}
	_ datasource.DataSourceWithConfigure = &projectDataSource{}
)

func ProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

type projectDataSource struct {
	client *dbt_cloud.Client
}

func (d *projectDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *projectDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = singleDatasourceSchema
}

func (d *projectDataSource) Configure(
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

func (d *projectDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state ProjectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var project *dbt_cloud.Project

	// Determine if we're looking up by project_id or name
	if !state.ID.IsNull() {
		if !state.Name.IsNull() {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"Both project_id and name were provided, only one is allowed",
			)
			return
		}

		var err error
		project, err = d.client.GetProject(state.ID.String())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read DBT Cloud Project",
				"An error occurred when reading the project by ID: "+err.Error(),
			)
			return
		}
	} else if !state.Name.IsNull() {
		var err error
		project, err = d.client.GetProjectByName(state.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read DBT Cloud Project",
				"An error occurred when reading the project by name: "+err.Error(),
			)
			return
		}
	} else {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either project_id or name must be provided",
		)
		return
	}

	// Map the project data to the state model
	state.ID = types.Int64Value(int64(*project.ID))
	state.Name = types.StringValue(project.Name)
	state.Description = types.StringValue(project.Description)
	state.DbtProjectSubdirectory = helper.ConvertStringPointer(project.DbtProjectSubdirectory)
	if project.ConnectionID != nil {
		state.ProjectConnection = &ProjectConnection{}
		state.ProjectConnection.ID = types.Int64Value(int64(*project.ConnectionID))
	} else {
		state.ProjectConnection = nil
	}

	if project.RepositoryID != nil {
		state.Repository = &ProjectRepository{}
		state.Repository.ID = types.Int64Value(int64(*project.RepositoryID))
	} else {
		state.Repository = nil
	}

	if project.FreshnessJobId != nil {
		state.FreshnessJobID = types.Int64Value(int64(*project.FreshnessJobId))
	} else {
		state.FreshnessJobID = types.Int64Null()
	}

	if project.DocsJobId != nil {
		state.DocsJobID = types.Int64Value(int64(*project.DocsJobId))
	} else {
		state.DocsJobID = types.Int64Null()
	}

	state.State = types.Int64Value(int64(project.State))
	state.DbtProjectType = types.Int64Value(int64(project.DbtProjectType))
	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
