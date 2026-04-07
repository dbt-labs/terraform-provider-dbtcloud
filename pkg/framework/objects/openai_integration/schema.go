package openai_integration

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var ValidKeyTypes = []string{"openai", "azure_openai"}

func (r *openAIIntegrationResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resource_schema.Schema{
		Description: helper.DocString(
			`Manages a bring-your-own-key OpenAI integration for a dbt Cloud account, enabling AI-powered features such as dbt Copilot.

			Two key types are supported:
			- ~~~openai~~~ — your own OpenAI API key
			- ~~~azure_openai~~~ — your own Azure OpenAI deployment

			**Lifecycle note:** dbt Cloud defaults to a dbt Labs-managed OpenAI key when no integration record exists. Creating this resource switches the account to a customer-managed key. Destroying it (or removing it from the Terraform config) deletes the record and automatically reverts the account to the dbt Labs-managed key — no additional steps are required.

			**Secret handling:** the API key is write-only and never returned after creation. Use ~~~key_value_wo~~~ with ~~~key_value_wo_version~~~ (Terraform 1.11+) to keep the secret out of state entirely. Use ~~~key_value~~~ for older Terraform versions — it is stored as a sensitive value in state.`,
		),
		Attributes: map[string]resource_schema.Attribute{
			"id": resource_schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the OpenAI integration.",
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
			"key_type": resource_schema.StringAttribute{
				Required:    true,
				Description: "The type of OpenAI key. One of: ~~~openai~~~, ~~~azure_openai~~~. To revert to the dbt Labs-managed key, destroy this resource.",
				Validators: []validator.String{
					stringvalidator.OneOf(ValidKeyTypes...),
				},
			},
			"key_value": resource_schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The OpenAI or Azure OpenAI API key. Stored as a sensitive value in Terraform state. Conflicts with ~~~key_value_wo~~~. For Terraform 1.11+, prefer ~~~key_value_wo~~~ to avoid storing secrets in state.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("key_value_wo")),
				},
			},
			"key_value_wo": resource_schema.StringAttribute{
				Optional:    true,
				WriteOnly:   true,
				Description: "Write-only variant of the API key (Terraform 1.11+). Never stored in state. Increment ~~~key_value_wo_version~~~ to rotate the key. Conflicts with ~~~key_value~~~.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("key_value")),
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("key_value")),
				},
			},
			"key_value_wo_version": resource_schema.Int64Attribute{
				Optional:    true,
				Description: "Increment this value to rotate the key when using ~~~key_value_wo~~~.",
			},
			"azure_endpoint": resource_schema.StringAttribute{
				Optional:    true,
				Description: "The Azure OpenAI endpoint URL. Required when ~~~key_type~~~ is ~~~azure_openai~~~.",
			},
			"azure_deployment_name": resource_schema.StringAttribute{
				Optional:    true,
				Description: "The Azure OpenAI deployment name. Required when ~~~key_type~~~ is ~~~azure_openai~~~.",
			},
			"azure_api_version": resource_schema.StringAttribute{
				Optional:    true,
				Description: "The Azure OpenAI API version (e.g. ~~~2024-02-01~~~). Required when ~~~key_type~~~ is ~~~azure_openai~~~.",
			},
			"created_at": resource_schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the integration was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": resource_schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the integration was last updated.",
			},
		},
	}
}
