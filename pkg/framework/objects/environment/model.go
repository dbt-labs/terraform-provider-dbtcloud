package environment

import "github.com/hashicorp/terraform-plugin-framework/types"

type EnvironmentDataSourceModel struct {
	EnvironmentID           types.Int64  `tfsdk:"environment_id"`
	ProjectID               types.Int64  `tfsdk:"project_id"`
	CredentialsID           types.Int64  `tfsdk:"credentials_id"`
	Name                    types.String `tfsdk:"name"`
	DbtVersion              types.String `tfsdk:"dbt_version"`
	Type                    types.String `tfsdk:"type"`
	UseCustomBranch         types.Bool   `tfsdk:"use_custom_branch"`
	CustomBranch            types.String `tfsdk:"custom_branch"`
	DeploymentType          types.String `tfsdk:"deployment_type"`
	ExtendedAttributesID    types.Int64  `tfsdk:"extended_attributes_id"`
	ConnectionID            types.Int64  `tfsdk:"connection_id"`
	EnableModelQueryHistory types.Bool   `tfsdk:"enable_model_query_history"`
}

type EnvironmentsDataSourceModel struct {
	ProjectID    types.Int64                  `tfsdk:"project_id"`
	Environments []EnvironmentDataSourceModel `tfsdk:"environments"`
}

type EnvironmentResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	EnvironmentID           types.Int64  `tfsdk:"environment_id"`
	CredentialID            types.Int64  `tfsdk:"credential_id"`
	ProjectID               types.Int64  `tfsdk:"project_id"`
	IsActive                types.Bool   `tfsdk:"is_active"`
	Name                    types.String `tfsdk:"name"`
	DbtVersion              types.String `tfsdk:"dbt_version"`
	Type                    types.String `tfsdk:"type"`
	UseCustomBranch         types.Bool   `tfsdk:"use_custom_branch"`
	CustomBranch            types.String `tfsdk:"custom_branch"`
	DeploymentType          types.String `tfsdk:"deployment_type"`
	ExtendedAttributesID    types.Int64  `tfsdk:"extended_attributes_id"`
	ConnectionID            types.Int64  `tfsdk:"connection_id"`
	EnableModelQueryHistory types.Bool   `tfsdk:"enable_model_query_history"`
}
