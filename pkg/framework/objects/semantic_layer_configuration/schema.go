package semantic_layer_configuration

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func (r *semanticLayerConfigurationResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resource_schema.Schema{
		Description: helper.DocString(
			`The resource allows basic configuration of the Semantic Layer for a specific project. For the feature to be completely functional, a Semantic Layer Credential is also required.
			See the documentationh ttps://docs.getdbt.com/docs/use-dbt-semantic-layer/dbt-sl for more information on the Semantic Layer.`,
		),
		Attributes: map[string]resource_schema.Attribute{
			"id": resource_schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the configuration",
			},
			"project_id": resource_schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the project",
			},
			"environment_id": resource_schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the environment",
			},
		},
	}
}
