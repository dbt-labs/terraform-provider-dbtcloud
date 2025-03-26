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
	EnvironmentID           types.Int64  `tfsdk:"environment_id"`
	IsActive                types.Bool   `tfsdk:"is_active"`
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

// isActive := d.Get("is_active").(bool)
// credentialId := d.Get("credential_id").(int)
// name := d.Get("name").(string)
// dbtVersion := d.Get("dbt_version").(string)
// type_ := d.Get("type").(string)
// useCustomBranch := d.Get("use_custom_branch").(bool)
// customBranch := d.Get("custom_branch").(string)
// deploymentType := d.Get("deployment_type").(string)
// extendedAttributesID := d.Get("extended_attributes_id").(int)
// connectionID := d.Get("connection_id").(int)
// enableModelQueryHistory := d.Get("enable_model_query_history").(bool)
// projectId := d.Get("project_id").(int)
