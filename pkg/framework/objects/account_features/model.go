package account_features

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AccountFeaturesResourceModel struct {
	ID                         types.String `tfsdk:"id"`
	AdvancedCI                 types.Bool   `tfsdk:"advanced_ci"`
	PartialParsing             types.Bool   `tfsdk:"partial_parsing"`
	RepoCaching                types.Bool   `tfsdk:"repo_caching"`
	AIFeatures                 types.Bool   `tfsdk:"ai_features"`
	CatalogIngestion           types.Bool   `tfsdk:"catalog_ingestion"`
	ExplorerAccountUI          types.Bool   `tfsdk:"explorer_account_ui"`
	FusionMigrationPermissions types.Bool   `tfsdk:"fusion_migration_permissions"`
}
