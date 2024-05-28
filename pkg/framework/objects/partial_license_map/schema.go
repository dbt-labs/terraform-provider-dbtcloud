package partial_license_map

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

func (r *partialLicenseMapResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: helper.DocString(
			`Set up partial license maps with only a subset of SSO groups for a given license type.

			This resource is different from ~~~dbtcloud_license_map~~~ as it allows having different resources setting up different groups for the same license type.

			If a company uses only one Terraform project/workspace to manage all their dbt Cloud Account config, it is recommended to use ~~~dbt_cloud_license_map~~~ instead of ~~~dbt_cloud_group_partial_license_map~~~.

			~> This is a new resource like other "partial" ones and any feedback is welcome in the GitHub repository.
			`,
		),
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the notification",
				// this is used so that we don't show that ID is going to change
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"license_type": schema.StringAttribute{
				Required:    true,
				Description: "The license type to update",
				// we need to replace the resource when we change the license type as this is the identifier
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sso_license_mapping_groups": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "List of SSO groups to map to the license type.",
			},
		},
	}
}
