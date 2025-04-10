package environment_variable_job_override

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EnvironmentVariableJobOverrideResourceModel is the model for the resource
type EnvironmentVariableJobOverrideResourceModel struct {
	ID                               types.String `tfsdk:"id"`
	ProjectID                        types.Int64  `tfsdk:"project_id"`
	Name                             types.String `tfsdk:"name"`
	JobDefinitionID                  types.Int64  `tfsdk:"job_definition_id"`
	RawValue                         types.String `tfsdk:"raw_value"`
	EnvironmentVariableJobOverrideID types.Int64  `tfsdk:"environment_variable_job_override_id"`
}
