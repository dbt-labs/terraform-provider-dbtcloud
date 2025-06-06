package synapse_credential

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SynapseCredentialResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	CredentialID        types.Int64  `tfsdk:"credential_id"`
	ProjectID           types.Int64  `tfsdk:"project_id"`
	Authentication      types.String `tfsdk:"authentication"`
	User                types.String `tfsdk:"user"`
	Password            types.String `tfsdk:"password"`
	TenantId            types.String `tfsdk:"tenant_id"`
	ClientId            types.String `tfsdk:"client_id"`
	ClientSecret        types.String `tfsdk:"client_secret"`
	Schema              types.String `tfsdk:"schema"`
	SchemaAuthorization types.String `tfsdk:"schema_authorization"`
	AdapterType         types.String `tfsdk:"adapter_type"`
	AdapterVersion      types.String `tfsdk:"adapter_version"`
}

type SynapseCredentialDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	CredentialID        types.Int64  `tfsdk:"credential_id"`
	ProjectID           types.Int64  `tfsdk:"project_id"`
	Authentication      types.String `tfsdk:"authentication"`
	User                types.String `tfsdk:"user"`
	Schema              types.String `tfsdk:"schema"`
	TenantId            types.String `tfsdk:"tenant_id"`
	ClientId            types.String `tfsdk:"client_id"`
	SchemaAuthorization types.String `tfsdk:"schema_authorization"`
	AdapterType         types.String `tfsdk:"adapter_type"`
}
