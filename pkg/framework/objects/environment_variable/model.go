package environment_variable

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EnvironmentVariableResourceModel is the model for the resource
type EnvironmentVariableResourceModel struct {
	ID                types.String `tfsdk:"id"`
	ProjectID         types.Int64  `tfsdk:"project_id"`
	Name              types.String `tfsdk:"name"`
	EnvironmentValues types.Map    `tfsdk:"environment_values"`
}

// EnvironmentVariableDataSourceModel is the model for the data source
type EnvironmentVariableDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	ProjectID         types.Int64  `tfsdk:"project_id"`
	Name              types.String `tfsdk:"name"`
	EnvironmentValues types.Map    `tfsdk:"environment_values"`
}
