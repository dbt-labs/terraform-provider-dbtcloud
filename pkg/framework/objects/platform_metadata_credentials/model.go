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
	AuthType                      types.String `tfsdk:"auth_type"`
	User                          types.String `tfsdk:"user"`
	Password                      types.String `tfsdk:"password"`
	PasswordWo                    types.String `tfsdk:"password_wo"`
	PasswordWoVersion             types.Int64  `tfsdk:"password_wo_version"`
	PrivateKey                    types.String `tfsdk:"private_key"`
	PrivateKeyWo                  types.String `tfsdk:"private_key_wo"`
	PrivateKeyWoVersion           types.Int64  `tfsdk:"private_key_wo_version"`
	PrivateKeyPassphrase          types.String `tfsdk:"private_key_passphrase"`
	PrivateKeyPassphraseWo        types.String `tfsdk:"private_key_passphrase_wo"`
	PrivateKeyPassphraseWoVersion types.Int64  `tfsdk:"private_key_passphrase_wo_version"`
	Role                          types.String `tfsdk:"role"`
	Warehouse                     types.String `tfsdk:"warehouse"`

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
	Token          types.String `tfsdk:"token"`
	TokenWo        types.String `tfsdk:"token_wo"`
	TokenWoVersion types.Int64  `tfsdk:"token_wo_version"`
	Catalog        types.String `tfsdk:"catalog"`

	// Read-only fields
	AdapterVersion types.String `tfsdk:"adapter_version"`
}
