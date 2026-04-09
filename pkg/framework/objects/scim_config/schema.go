package scim_config

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *scimConfigResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resource_schema.Schema{
		Description: helper.DocString(
			`Manages the SCIM configuration for a dbt Cloud account. SCIM (System for Cross-domain Identity Management) allows identity providers such as Okta and Azure AD to automatically provision and deprovision users and groups.

			This is a singleton resource — only one SCIM configuration exists per account. Destroying the resource resets the configuration to its defaults (SCIM disabled, manual updates allowed, license type not SCIM-controlled).

			Requires the SCIM feature to be enabled on the account (enterprise plans only).`,
		),
		Attributes: map[string]resource_schema.Attribute{
			"id": resource_schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the resource (matches the dbt Cloud account ID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": resource_schema.BoolAttribute{
				Required:    true,
				Description: "Whether SCIM provisioning is enabled for the account.",
			},
			"manual_updates_allowed": resource_schema.BoolAttribute{
				Required:    true,
				Description: "Whether administrators can manually update users and groups that are managed by SCIM. When set to ~~~false~~~, SCIM is the sole source of truth for user and group management.",
			},
			"scim_controlled_license_type": resource_schema.BoolAttribute{
				Required:    true,
				Description: "Whether the dbt Cloud license type (Developer, Read-Only, IT) is controlled by SCIM attribute mapping. When set to ~~~false~~~, license types are managed manually inside dbt Cloud.",
			},
			"resource_metadata": resource_schema.DynamicAttribute{
				Optional:    true,
				Description: "Metadata for tracking resource identity during account migrations. Stored in Terraform state only and not sent to the API.",
			},
		},
	}
}
