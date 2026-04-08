package openai_integration

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type OpenAIIntegrationResourceModel struct {
	ID                  types.Int64  `tfsdk:"id"`
	AccountID           types.Int64  `tfsdk:"account_id"`
	KeyType             types.String `tfsdk:"key_type"`
	KeyValue            types.String `tfsdk:"key_value"`
	KeyValueWO          types.String `tfsdk:"key_value_wo"`
	KeyValueWOVersion   types.Int64  `tfsdk:"key_value_wo_version"`
	AzureEndpoint       types.String `tfsdk:"azure_endpoint"`
	AzureDeploymentName types.String `tfsdk:"azure_deployment_name"`
	AzureAPIVersion     types.String `tfsdk:"azure_api_version"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}
