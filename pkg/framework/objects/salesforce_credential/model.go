package salesforce_credential

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SalesforceCredentialDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	CredentialID types.Int64  `tfsdk:"credential_id"`
	ProjectID    types.Int64  `tfsdk:"project_id"`
	Username     types.String `tfsdk:"username"`
	TargetName   types.String `tfsdk:"target_name"`
	NumThreads   types.Int64  `tfsdk:"num_threads"`
}

type SalesforceCredentialResourceModel struct {
	ID           types.String `tfsdk:"id"`
	CredentialID types.Int64  `tfsdk:"credential_id"`
	ProjectID    types.Int64  `tfsdk:"project_id"`
	Username     types.String `tfsdk:"username"`
	ClientID     types.String `tfsdk:"client_id"`
	PrivateKey   types.String `tfsdk:"private_key"`
	TargetName   types.String `tfsdk:"target_name"`
	NumThreads   types.Int64  `tfsdk:"num_threads"`
}
