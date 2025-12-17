package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type BigQueryAuthValidator struct{}

func (v BigQueryAuthValidator) Description(ctx context.Context) string {
	return "Validates BigQuery authentication configuration based on deployment_env_auth_type"
}

func (v BigQueryAuthValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates BigQuery authentication configuration based on `deployment_env_auth_type`"
}

func (v BigQueryAuthValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	// Check if bigquery block exists
	var bigqueryObj types.Object
	diags := req.Config.GetAttribute(ctx, path.Root("bigquery"), &bigqueryObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If bigquery is null or unknown, skip validation
	if bigqueryObj.IsNull() || bigqueryObj.IsUnknown() {
		return
	}

	// Get deployment_env_auth_type value
	var authType types.String
	diags = req.Config.GetAttribute(ctx, path.Root("bigquery").AtName("deployment_env_auth_type"), &authType)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If deployment_env_auth_type is null or unknown, skip conditional validation
	if authType.IsNull() || authType.IsUnknown() {
		return
	}

	authTypeValue := authType.ValueString()

	// TODO: Update this validation once the API supports external-oauth-wif without requiring
	// service account fields. Currently, the API requires service account fields regardless
	// of auth type, so we only add the extra requirement for OAuth fields when using WIF.

	if authTypeValue == "external-oauth-wif" {
		// Validate that application_id and application_secret are set
		oauthFields := []string{"application_id", "application_secret"}
		for _, field := range oauthFields {
			var fieldValue types.String
			diags = req.Config.GetAttribute(ctx, path.Root("bigquery").AtName(field), &fieldValue)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			if fieldValue.IsNull() || fieldValue.IsUnknown() {
				resp.Diagnostics.AddAttributeError(
					path.Root("bigquery").AtName(field),
					"Missing Required Field for External OAuth WIF",
					"When deployment_env_auth_type is 'external-oauth-wif', the field '"+field+"' must be specified.",
				)
			}
		}
	}

	// Service account fields are currently required by the API for all auth types
	serviceAccountFields := []string{
		"private_key_id",
		"private_key",
		"client_email",
		"client_id",
		"auth_uri",
		"token_uri",
		"auth_provider_x509_cert_url",
		"client_x509_cert_url",
	}

	for _, field := range serviceAccountFields {
		var fieldValue types.String
		diags = req.Config.GetAttribute(ctx, path.Root("bigquery").AtName(field), &fieldValue)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if fieldValue.IsNull() || fieldValue.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("bigquery").AtName(field),
				"Missing Required Field for BigQuery",
				"The field '"+field+"' must be specified for BigQuery connections.",
			)
		}
	}
}
