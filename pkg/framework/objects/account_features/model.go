package account_features

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AccountFeaturesResourceModel struct {
	ID             types.String `tfsdk:"id"`
	AdvancedCI     types.Bool   `tfsdk:"advanced_ci"`
	PartialParsing types.Bool   `tfsdk:"partial_parsing"`
	RepoCaching    types.Bool   `tfsdk:"repo_caching"`
}
