package extended_attributes

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ExtendedAttributesResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	ExtendedAttributesID types.Int64  `tfsdk:"extended_attributes_id"`
	State                types.Int64  `tfsdk:"state"`
	ProjectID            types.Int64  `tfsdk:"project_id"`
	ExtendedAttributes   types.String `tfsdk:"extended_attributes"`
}

type ExtendedAttributesDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	ExtendedAttributesID types.Int64  `tfsdk:"extended_attributes_id"`
	ProjectID            types.Int64  `tfsdk:"project_id"`
	State                types.Int64  `tfsdk:"state"`
	ExtendedAttributes   types.String `tfsdk:"extended_attributes"`
}
