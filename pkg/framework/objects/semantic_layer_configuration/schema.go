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
			`Configure an external OAuth integration for the data warehouse. Currently supports Okta and Entra ID (i.e. Azure AD) for Snowflake.
			
			See the [documentation](https://docs.getdbt.com/docs/cloud/manage-access/external-oauth) for more information on how to configure it.`,
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
