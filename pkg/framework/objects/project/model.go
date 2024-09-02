package project

import "github.com/hashicorp/terraform-plugin-framework/types"

type ProjectsDataSourceModel struct {
	NameContains types.String                  `tfsdk:"name_contains"`
	Projects     []ProjectConnectionRepository `tfsdk:"projects"`
}

type ProjectConnectionRepository struct {
	ID                     types.Int64        `tfsdk:"id"`
	Name                   types.String       `tfsdk:"name"`
	Description            types.String       `tfsdk:"description"`
	SemanticLayerConfigID  types.Int64        `tfsdk:"semantic_layer_config_id"`
	DbtProjectSubdirectory types.String       `tfsdk:"dbt_project_subdirectory"`
	CreatedAt              types.String       `tfsdk:"created_at"`
	UpdatedAt              types.String       `tfsdk:"updated_at"`
	Connection             *ProjectConnection `tfsdk:"connection"`
	Repository             *ProjectRepository `tfsdk:"repository"`
}

type ProjectRepository struct {
	ID                     types.Int64  `tfsdk:"id"`
	RemoteUrl              types.String `tfsdk:"remote_url"`
	PullRequestURLTemplate types.String `tfsdk:"pull_request_url_template"`
}

type ProjectConnection struct {
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	AdapterVersion types.String `tfsdk:"adapter_version"`
}
