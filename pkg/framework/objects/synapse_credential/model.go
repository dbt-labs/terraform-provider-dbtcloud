package synapse_credential

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SynapseCredentialResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	CredentialID           types.Int64  `tfsdk:"credential_id"`
	ProjectID              types.Int64  `tfsdk:"project_id"`
	Authentication         types.String `tfsdk:"authentication"`
	User                   types.String `tfsdk:"user"`
	Password               types.String `tfsdk:"password"`
	PasswordWo             types.String `tfsdk:"password_wo"`
	PasswordWoVersion      types.Int64  `tfsdk:"password_wo_version"`
	TenantId               types.String `tfsdk:"tenant_id"`
	ClientId               types.String `tfsdk:"client_id"`
	ClientSecret           types.String `tfsdk:"client_secret"`
	ClientSecretWo         types.String `tfsdk:"client_secret_wo"`
	ClientSecretWoVersion  types.Int64  `tfsdk:"client_secret_wo_version"`
	Schema                 types.String `tfsdk:"schema"`
	SchemaAuthorization    types.String `tfsdk:"schema_authorization"`
	AdapterType            types.String `tfsdk:"adapter_type"`
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
