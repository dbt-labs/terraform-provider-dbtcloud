package scim_config_token

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *scimConfigTokenResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resource_schema.Schema{
		Description: helper.DocString(
			`Manages a SCIM API token for a dbt Cloud account. SCIM tokens are used by identity providers (e.g. Okta, Azure AD) to provision and deprovision users and groups automatically.

			The token value is only available immediately after creation and is stored in Terraform state as a sensitive value. It cannot be retrieved from the API afterwards.

			Requires the SCIM feature to be enabled on the account (enterprise plans only).`,
		),
		Attributes: map[string]resource_schema.Attribute{
			"id": resource_schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the SCIM token.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": resource_schema.StringAttribute{
				Required:    true,
				Description: "A human-readable name for the token. Changing this value forces a new token to be created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token_string": resource_schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The SCIM token value. Only available immediately after creation — not returned by subsequent API reads. Store this value securely.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": resource_schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the token was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_used": resource_schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the token was last used. Null if never used.",
			},
			"resource_metadata": resource_schema.DynamicAttribute{
				Optional:    true,
				Description: "Metadata for tracking resource identity during account migrations. Stored in Terraform state only and not sent to the API.",
			},
		},
	}
}
