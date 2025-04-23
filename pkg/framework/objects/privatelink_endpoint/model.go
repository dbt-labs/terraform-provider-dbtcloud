package privatelink_endpoint

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PrivatelinkEndpointDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.Int64  `tfsdk:"project_id"`
	CredentialID  types.Int64  `tfsdk:"credential_id"`
	IsActive      types.Bool   `tfsdk:"is_active"`
	DefaultSchema types.String `tfsdk:"default_schema"`
	Username      types.String `tfsdk:"username"`
	NumThreads    types.Int64  `tfsdk:"num_threads"`
}
