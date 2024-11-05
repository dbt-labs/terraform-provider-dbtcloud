package oauth_configuration

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *oAuthConfigurationResource) Schema(
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
				Description: "The ID of the OAuth configuration",
				// this is used so that we don't show that ID is going to change
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"type": resource_schema.StringAttribute{
				Required:    true,
				Description: "The type of OAuth integration (`entra` or `okta`)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("okta", "entra"),
				},
			},
			"name": resource_schema.StringAttribute{
				Required:    true,
				Description: "The name of OAuth integration",
			},
			"client_id": resource_schema.StringAttribute{
				Required:    true,
				Description: "The Client ID for the OAuth integration",
			},
			"client_secret": resource_schema.StringAttribute{
				Required:    true,
				Description: "The Client secret for the OAuth integration",
				Sensitive:   true,
			},
			"authorize_url": resource_schema.StringAttribute{
				Required:    true,
				Description: "The Authorize URL for the OAuth integration",
			},
			"token_url": resource_schema.StringAttribute{
				Required:    true,
				Description: "The Token URL for the OAuth integration",
			},
			"redirect_uri": resource_schema.StringAttribute{
				Required:    true,
				Description: "The redirect URL for the OAuth integration",
			},
			"application_id_uri": resource_schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The Application ID URI for the OAuth integration. Only for Entra",
				Default:     stringdefault.StaticString(""),
			},
		},
	}
}
