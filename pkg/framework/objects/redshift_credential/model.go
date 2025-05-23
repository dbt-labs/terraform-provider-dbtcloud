package redshift_credential

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RedshiftCredentialResourceModel struct {
	ID            types.String `tfsdk:"id"`
	CredentialID  types.Int64  `tfsdk:"credential_id"`
	ProjectID     types.Int64  `tfsdk:"project_id"`
	IsActive      types.Bool   `tfsdk:"is_active"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	DefaultSchema types.String `tfsdk:"default_schema"`
	NumThreads    types.Int64  `tfsdk:"num_threads"`
}

type RedshiftCredentialDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	CredentialID  types.Int64  `tfsdk:"credential_id"`
	ProjectID     types.Int64  `tfsdk:"project_id"`
	IsActive      types.Bool   `tfsdk:"is_active"`
	NumThreads    types.Int64  `tfsdk:"num_threads"`
	DefaultSchema types.String `tfsdk:"default_schema"`
}
