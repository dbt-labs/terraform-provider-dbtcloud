package fabric_credential

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// FabricCredentialResourceModel is the model for the resource
type FabricCredentialResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	CredentialID        types.Int64  `tfsdk:"credential_id"`
	ProjectID           types.Int64  `tfsdk:"project_id"`
	User                types.String `tfsdk:"user"`
	Password            types.String `tfsdk:"password"`
	TenantId            types.String `tfsdk:"tenant_id"`
	ClientId            types.String `tfsdk:"client_id"`
	ClientSecret        types.String `tfsdk:"client_secret"`
	Schema              types.String `tfsdk:"schema"`
	SchemaAuthorization types.String `tfsdk:"schema_authorization"`
	AdapterID           types.Int64  `tfsdk:"adapter_id"`
}
