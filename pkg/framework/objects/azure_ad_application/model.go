package azure_ad_application

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AzureADApplicationResourceModel struct {
	ID                               types.Int64  `tfsdk:"id"`
	AccountID                        types.Int64  `tfsdk:"account_id"`
	OrganizationName                 types.String `tfsdk:"organization_name"`
	ClientID                         types.String `tfsdk:"client_id"`
	ClientSecret                     types.String `tfsdk:"client_secret"`
	TenantID                         types.String `tfsdk:"tenant_id"`
	AzureServiceAuthenticationMethod types.String `tfsdk:"azure_service_authentication_method"`
	OAuthRedirectURIDomain           types.String `tfsdk:"oauth_redirect_uri_domain"`
	CreatedAt                        types.String `tfsdk:"created_at"`
	UpdatedAt                        types.String `tfsdk:"updated_at"`
}
