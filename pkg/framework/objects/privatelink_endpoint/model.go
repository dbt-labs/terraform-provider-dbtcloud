package privatelink_endpoint

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PrivatelinkEndpointDataSourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	PrivatelinkEndpointType types.String `tfsdk:"type"`
	PrivatelinkEndpointURL  types.String `tfsdk:"private_link_endpoint_url"`
	CIDRRange               types.String `tfsdk:"cidr_range"`
}
