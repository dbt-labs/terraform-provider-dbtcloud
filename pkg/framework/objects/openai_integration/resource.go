package openai_integration

import (
	"context"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = &openAIIntegrationResource{}
	_ resource.ResourceWithConfigure      = &openAIIntegrationResource{}
	_ resource.ResourceWithImportState    = &openAIIntegrationResource{}
	_ resource.ResourceWithValidateConfig = &openAIIntegrationResource{}
)

func OpenAIIntegrationResource() resource.Resource {
	return &openAIIntegrationResource{}
}

type openAIIntegrationResource struct {
	client *dbt_cloud.Client
}

func (r *openAIIntegrationResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_openai_integration"
}

func (r *openAIIntegrationResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*dbt_cloud.Client)
}

// ── ValidateConfig ────────────────────────────────────────────────────────────

func (r *openAIIntegrationResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var config OpenAIIntegrationResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyType := config.KeyType.ValueString()

	azureFields := map[string]types.String{
		"azure_endpoint":        config.AzureEndpoint,
		"azure_deployment_name": config.AzureDeploymentName,
		"azure_api_version":     config.AzureAPIVersion,
	}

	hasKey := !config.KeyValue.IsNull() || !config.KeyValueWO.IsNull()

	switch keyType {
	case "azure_openai":
		for field, val := range azureFields {
			if val.IsNull() || val.IsUnknown() || val.ValueString() == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root(field),
					"Missing required field",
					field+" is required when key_type is azure_openai.",
				)
			}
		}
		if !hasKey {
			resp.Diagnostics.AddError(
				"Missing API key",
				"Either key_value or key_value_wo must be set when key_type is azure_openai.",
			)
		}

	case "openai":
		for field, val := range azureFields {
			if !val.IsNull() && !val.IsUnknown() {
				resp.Diagnostics.AddAttributeError(
					path.Root(field),
					"Invalid field",
					field+" must not be set when key_type is openai.",
				)
			}
		}
		if !hasKey {
			resp.Diagnostics.AddError(
				"Missing API key",
				"Either key_value or key_value_wo must be set when key_type is openai.",
			)
		}
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// applyAPIResponse maps the API response back onto the model.
// key_value, key_value_wo, and key_value_wo_version are never returned by the
// API and must not be overwritten.
func applyAPIResponse(api *dbt_cloud.OpenAIIntegration, m *OpenAIIntegrationResourceModel) {
	m.ID = types.Int64Value(*api.ID)
	m.AccountID = types.Int64Value(api.AccountID)
	m.KeyType = types.StringValue(api.KeyType)

	if api.AzureEndpoint != nil {
		m.AzureEndpoint = types.StringValue(*api.AzureEndpoint)
	} else {
		m.AzureEndpoint = types.StringNull()
	}
	if api.AzureDeploymentName != nil {
		m.AzureDeploymentName = types.StringValue(*api.AzureDeploymentName)
	} else {
		m.AzureDeploymentName = types.StringNull()
	}
	if api.AzureAPIVersion != nil {
		m.AzureAPIVersion = types.StringValue(*api.AzureAPIVersion)
	} else {
		m.AzureAPIVersion = types.StringNull()
	}
	if api.CreatedAt != nil {
		m.CreatedAt = types.StringValue(*api.CreatedAt)
	}
	if api.UpdatedAt != nil {
		m.UpdatedAt = types.StringValue(*api.UpdatedAt)
	}
}

// buildIntegration constructs the API payload. key_value_wo comes from config
// (write-only fields are stripped from plan); key_value comes from plan.
func buildIntegration(plan, config OpenAIIntegrationResourceModel) dbt_cloud.OpenAIIntegration {
	integration := dbt_cloud.OpenAIIntegration{KeyType: plan.KeyType.ValueString()}

	// Prefer write-only; fall back to regular sensitive attribute.
	keyValue := helper.ResolveWriteOnlyString(config.KeyValueWO, plan.KeyValue)
	if keyValue != "" {
		integration.KeyValue = &keyValue
	}
	if !plan.AzureEndpoint.IsNull() {
		v := plan.AzureEndpoint.ValueString()
		integration.AzureEndpoint = &v
	}
	if !plan.AzureDeploymentName.IsNull() {
		v := plan.AzureDeploymentName.ValueString()
		integration.AzureDeploymentName = &v
	}
	if !plan.AzureAPIVersion.IsNull() {
		v := plan.AzureAPIVersion.ValueString()
		integration.AzureAPIVersion = &v
	}
	return integration
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *openAIIntegrationResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan OpenAIIntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	var config OpenAIIntegrationResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateOpenAIIntegration(buildIntegration(plan, config))
	if err != nil {
		resp.Diagnostics.AddError("Unable to create OpenAI integration", err.Error())
		return
	}

	applyAPIResponse(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *openAIIntegrationResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state OpenAIIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	integration, err := r.client.GetOpenAIIntegration(state.ID.ValueInt64())
	if err != nil {
		if helper.HandleResourceNotFound(ctx, err, &resp.Diagnostics, &resp.State, "OpenAI integration") {
			return
		}
		resp.Diagnostics.AddError("Error reading OpenAI integration", err.Error())
		return
	}

	applyAPIResponse(integration, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *openAIIntegrationResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan OpenAIIntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	var config OpenAIIntegrationResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	var state OpenAIIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateOpenAIIntegration(state.ID.ValueInt64(), buildIntegration(plan, config))
	if err != nil {
		resp.Diagnostics.AddError("Unable to update OpenAI integration", err.Error())
		return
	}

	applyAPIResponse(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *openAIIntegrationResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state OpenAIIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteOpenAIIntegration(state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Unable to delete OpenAI integration", err.Error())
	}
}

// ── Import ────────────────────────────────────────────────────────────────────

func (r *openAIIntegrationResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing OpenAI integration ID for import", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
