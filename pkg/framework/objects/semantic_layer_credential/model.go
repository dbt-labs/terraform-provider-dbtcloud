package semantic_layer_credential

import (
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/snowflake_credential"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SemanticLayerConfigurationModel struct {
	Name           types.String `tfsdk:"name"`
	AdapterVersion types.String `tfsdk:"adapter_version"`
	ProjectID      types.Int64  `tfsdk:"project_id"`
}

type SnowflakeSLCredentialModel struct {
	ID            types.Int64                                           `tfsdk:"id"`
	Configuration SemanticLayerConfigurationModel                       `tfsdk:"configuration"`
	Credential    snowflake_credential.SnowflakeCredentialResourceModel `tfsdk:"credential"`
}
