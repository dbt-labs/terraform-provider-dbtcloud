package semantic_layer_credential_service_token_mapping

import "github.com/hashicorp/terraform-plugin-framework/types"

type SemanticLayerCredentialServiceTokenMapping struct {
	ID                         types.Int64  `tfsdk:"id"`
	ProjectID                  types.Int64  `tfsdk:"project_id"`
	SemanticLayerCredentialID  types.Int64  `tfsdk:"semantic_layer_credential_id"`
	ServiceTokenID             types.Int64  `tfsdk:"service_token_id"`
}
