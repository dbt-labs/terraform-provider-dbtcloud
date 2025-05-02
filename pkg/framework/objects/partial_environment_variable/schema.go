package partial_environment_variable

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *partialEnvironmentVariableResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: helper.DocString(
			`Set up partial environment variables with only a subset of environment values for a given environment variable.

			This resource is different from ~~~dbtcloud_environment_variable~~~ as it allows having different resources setting up different environment values for the same environment variable.

			If a company uses only one Terraform project/workspace to manage all their dbt Cloud Account config, it is recommended to use ~~~dbt_cloud_environment_variable~~~ instead of ~~~dbt_cloud_partial_environment_variable~~~.

			~> This resource allows provider users to update specific environment values without knowing or changing values for other environments.

			**IMPORTANT** This resource can also manage other resources' fields. We strongly advise against overlapping scope (i.e. updating values managed by other resources) as this could lead to unexpected changes in the remote state.

## Example usage:
` + "```" + `terraform
# Only mentions one of the environments set up on the project, instead of all of them
resource "dbtcloud_partial_environment_variable" "test_env_var_partial" {
  project_id = dbtcloud_project.test_project.id
  name       = "DBT_TESTVAR"
  environment_values = {
    (dbtcloud_environment.test_env_prod.name) = "prodval"
  }
` + "```" + `
			`,
		),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the environment variable in the format 'project_id:name'",
				// this is used so that we don't show that ID is going to change
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.Int64Attribute{
				Required:    true,
				Description: "Project ID to create or update the environment variable in",
				// we need to replace the resource when we change the project ID as this is part of the identifier
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name for the variable, must be unique within a project, must be prefixed with 'DBT_'",
				// we need to replace the resource when we change the name as this is part of the identifier
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_values": schema.MapAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "Map from environment names to respective variable value. This field is not set as sensitive so take precautions when using secret environment variables. Only the specified environment values will be managed by this resource.",
			},
		},
	}
}
