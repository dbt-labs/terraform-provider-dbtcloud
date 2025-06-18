package postgres_credential

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PostgresCredentialDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.Int64  `tfsdk:"project_id"`
	CredentialID  types.Int64  `tfsdk:"credential_id"`
	IsActive      types.Bool   `tfsdk:"is_active"`
	DefaultSchema types.String `tfsdk:"default_schema"`
	Username      types.String `tfsdk:"username"`
	NumThreads    types.Int64  `tfsdk:"num_threads"`
}

type PostgresCredentialResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	ProjectID               types.Int64  `tfsdk:"project_id"`
	CredentialID            types.Int64  `tfsdk:"credential_id"`
	IsActive                types.Bool   `tfsdk:"is_active"`
	DefaultSchema           types.String `tfsdk:"default_schema"`
	Username                types.String `tfsdk:"username"`
	NumThreads              types.Int64  `tfsdk:"num_threads"`
	Type                    types.String `tfsdk:"type"`
	TargetName              types.String `tfsdk:"target_name"`
	Password                types.String `tfsdk:"password"`
	SemanticLayerCredential types.Bool   `tfsdk:"semantic_layer_credential"`
}
