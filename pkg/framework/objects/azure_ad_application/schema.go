package azure_ad_application

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var ValidAuthMethods = []string{"service_user", "service_principal"}

func (r *azureADApplicationResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resource_schema.Schema{
		Description: helper.DocString(
			`Manages an Azure Active Directory (Microsoft Entra ID) application registration for a dbt Cloud account. This enables Azure DevOps integration, allowing dbt Cloud to access Azure DevOps repositories for project setup.

			The ~~~client_id~~~, ~~~client_secret~~~ and ~~~tenant_id~~~ are encrypted at rest and never returned by the API. They are stored as sensitive values in Terraform state so they can be resent on every update — the API requires all three on both create and update.

			**Destroy behaviour:** running ~~~terraform destroy~~~ calls the dbt Cloud DELETE endpoint, which marks the record as inactive. Due to a known dbt Cloud backend limitation, the underlying database row is retained and re-creating the resource against the same account without a backend cleanup will fail with a unique-constraint error. If you need to recreate the resource after a destroy, contact dbt Cloud support to have the orphaned record removed, or use ~~~terraform import~~~ to re-adopt the existing record ID.

			Requires the Azure DevOps integration feature to be enabled on the account (enterprise plans only).`,
		),
		Attributes: map[string]resource_schema.Attribute{
			"id": resource_schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the Azure AD application.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"account_id": resource_schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the dbt Cloud account.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"organization_name": resource_schema.StringAttribute{
				Required:    true,
				Description: "The name of the Azure DevOps organization.",
			},
			"client_id": resource_schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The client ID (application ID) of the Azure AD app registration. Stored as a sensitive value — the API never returns it.",
			},
			"client_secret": resource_schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The client secret of the Azure AD app registration. Stored as a sensitive value — the API never returns it.",
			},
			"tenant_id": resource_schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The tenant ID of the Azure AD directory. Stored as a sensitive value — the API never returns it.",
			},
			"azure_service_authentication_method": resource_schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The method used for service authentication. One of: ~~~service_user~~~, ~~~service_principal~~~. Defaults to ~~~service_user~~~.",
				Validators: []validator.String{
					stringvalidator.OneOf(ValidAuthMethods...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"oauth_redirect_uri_domain": resource_schema.StringAttribute{
				Computed:    true,
				Description: "The domain used for the OAuth redirect URI. Set automatically by dbt Cloud based on the account's subdomain.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": resource_schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the application was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": resource_schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the application was last updated.",
			},
			"resource_metadata": resource_schema.DynamicAttribute{
				Optional:    true,
				Description: "Metadata for tracking resource identity during account migrations. Stored in Terraform state only and not sent to the API.",
			},
		},
	}
}
