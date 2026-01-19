package platform_metadata_credentials

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SnowflakePlatformMetadataCredentialResourceModel represents the Terraform state for Snowflake
type SnowflakePlatformMetadataCredentialResourceModel struct {
	ID           types.String `tfsdk:"id"`
	CredentialID types.Int64  `tfsdk:"credential_id"`
	ConnectionID types.Int64  `tfsdk:"connection_id"`

	// Feature flags
	CatalogIngestionEnabled types.Bool `tfsdk:"catalog_ingestion_enabled"`
	CostOptimizationEnabled types.Bool `tfsdk:"cost_optimization_enabled"`
	CostInsightsEnabled     types.Bool `tfsdk:"cost_insights_enabled"`

	// Snowflake-specific fields
	AuthType             types.String `tfsdk:"auth_type"`
	User                 types.String `tfsdk:"user"`
	Password             types.String `tfsdk:"password"`
	PrivateKey           types.String `tfsdk:"private_key"`
	PrivateKeyPassphrase types.String `tfsdk:"private_key_passphrase"`
	Role                 types.String `tfsdk:"role"`
	Warehouse            types.String `tfsdk:"warehouse"`

	// Read-only fields
	AdapterVersion types.String `tfsdk:"adapter_version"`
}

// DatabricksPlatformMetadataCredentialResourceModel represents the Terraform state for Databricks
type DatabricksPlatformMetadataCredentialResourceModel struct {
	ID           types.String `tfsdk:"id"`
	CredentialID types.Int64  `tfsdk:"credential_id"`
	ConnectionID types.Int64  `tfsdk:"connection_id"`

	// Feature flags
	CatalogIngestionEnabled types.Bool `tfsdk:"catalog_ingestion_enabled"`
	CostOptimizationEnabled types.Bool `tfsdk:"cost_optimization_enabled"`
	CostInsightsEnabled     types.Bool `tfsdk:"cost_insights_enabled"`

	// Databricks-specific fields
	Token   types.String `tfsdk:"token"`
	Catalog types.String `tfsdk:"catalog"`

	// Read-only fields
	AdapterVersion types.String `tfsdk:"adapter_version"`
}
