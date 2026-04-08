package auth_provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AuthProviderResourceModel struct {
	ID    types.Int64  `tfsdk:"id"`
	State types.Int64  `tfsdk:"state"`

	// Common
	Type                  types.String `tfsdk:"type"`
	Slug                  types.String `tfsdk:"slug"`
	AllowPasswordBackdoor types.Bool   `tfsdk:"allow_password_backdoor"`
	LoginURL              types.String `tfsdk:"login_url"`
	CreatedAt             types.String `tfsdk:"created_at"`
	UpdatedAt             types.String `tfsdk:"updated_at"`

	// SAML / Okta
	EntityID       types.String `tfsdk:"entity_id"`
	SsoURL         types.String `tfsdk:"sso_url"`
	Cert           types.String `tfsdk:"cert"`
	CertWo         types.String `tfsdk:"cert_wo"`
	CertWoVersion  types.Int64  `tfsdk:"cert_wo_version"`
	CertExpiryDate types.String `tfsdk:"cert_expiry_date"`
	SignRequest    types.Bool   `tfsdk:"sign_request"`
	AttributeMap   types.String `tfsdk:"attribute_map"`

	// Azure AD / Google Workspace
	ClientID              types.String `tfsdk:"client_id"`
	ClientSecret          types.String `tfsdk:"client_secret"`
	ClientSecretWo        types.String `tfsdk:"client_secret_wo"`
	ClientSecretWoVersion types.Int64  `tfsdk:"client_secret_wo_version"`

	// Azure AD only
	TenantID              types.String `tfsdk:"tenant_id"`
	Domain                types.String `tfsdk:"domain"`
	IncludeIndirectGroups types.Bool   `tfsdk:"include_indirect_groups"`
	MaxGroupsToRetrieve   types.Int64  `tfsdk:"max_groups_to_retrieve"`

	// Google Workspace only
	GsuiteAdminID     types.String `tfsdk:"gsuite_admin_id"`
	AuthorizationURL  types.String `tfsdk:"authorization_url"`
	AdminRefreshToken types.String `tfsdk:"admin_refresh_token"`
}
