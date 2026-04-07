package scim_config_token

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SCIMConfigTokenResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	TokenString types.String `tfsdk:"token_string"`
	CreatedAt   types.String `tfsdk:"created_at"`
	LastUsed    types.String `tfsdk:"last_used"`
}
