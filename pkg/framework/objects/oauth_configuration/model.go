package oauth_configuration

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type OAuthConfigurationResourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	Type             types.String `tfsdk:"type"`
	Name             types.String `tfsdk:"name"`
	ClientId         types.String `tfsdk:"client_id"`
	ClientSecret     types.String `tfsdk:"client_secret"`
	AuthorizeUrl     types.String `tfsdk:"authorize_url"`
	TokenUrl         types.String `tfsdk:"token_url"`
	RedirectUri      types.String `tfsdk:"redirect_uri"`
	ApplicationIdUri types.String `tfsdk:"application_id_uri"`
}
