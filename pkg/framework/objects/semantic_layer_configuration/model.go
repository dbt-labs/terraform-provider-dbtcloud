package semantic_layer_configuration

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SemanticLayerConfigurationModel struct {
	ID              types.Int64  `tfsdk:"id"`
	ProjectID       types.Int64  `tfsdk:"project_id"`
	EnvironmentID   types.Int64  `tfsdk:"environment_id"`
}
