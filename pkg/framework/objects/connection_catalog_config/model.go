package connection_catalog_config

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ConnectionCatalogConfigResourceModel represents the Terraform state for this resource
type ConnectionCatalogConfigResourceModel struct {
	ID           types.String `tfsdk:"id"`
	ConnectionID types.Int64  `tfsdk:"connection_id"`

	// Filter lists
	DatabaseAllow types.List `tfsdk:"database_allow"`
	DatabaseDeny  types.List `tfsdk:"database_deny"`
	SchemaAllow   types.List `tfsdk:"schema_allow"`
	SchemaDeny    types.List `tfsdk:"schema_deny"`
	TableAllow    types.List `tfsdk:"table_allow"`
	TableDeny     types.List `tfsdk:"table_deny"`
	ViewAllow     types.List `tfsdk:"view_allow"`
	ViewDeny      types.List `tfsdk:"view_deny"`
}
