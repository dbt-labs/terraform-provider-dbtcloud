package platform_metadata_credentials

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = &bigQueryPlatformMetadataCredentialResource{}
	_ resource.ResourceWithConfigure      = &bigQueryPlatformMetadataCredentialResource{}
	_ resource.ResourceWithImportState    = &bigQueryPlatformMetadataCredentialResource{}
	_ resource.ResourceWithValidateConfig = &bigQueryPlatformMetadataCredentialResource{}
)

// BigQueryPlatformMetadataCredentialResource returns a new resource instance
func BigQueryPlatformMetadataCredentialResource() resource.Resource {
	return &bigQueryPlatformMetadataCredentialResource{}
}

type bigQueryPlatformMetadataCredentialResource struct {
	client *dbt_cloud.Client
}

func (r *bigQueryPlatformMetadataCredentialResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_bigquery_platform_metadata_credential"
}

func (r *bigQueryPlatformMetadataCredentialResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = BigQueryPlatformMetadataCredentialSchema
}

func (r *bigQueryPlatformMetadataCredentialResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*dbt_cloud.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *dbt_cloud.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}

	r.client = client
}

// ValidateConfig checks the fields required by each auth_type. A null auth_type means the
// default, service-account-json.
func (r *bigQueryPlatformMetadataCredentialResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var config BigQueryPlatformMetadataCredentialResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.AuthType.IsUnknown() {
		return
	}

	authType := config.AuthType.ValueString()
	authTypeNote := ""
	if config.AuthType.IsNull() {
		authType = bigQueryAuthTypeServiceAccountJSON
		authTypeNote = " (auth_type defaults to " + bigQueryAuthTypeServiceAccountJSON + ")"
	}

	serviceAccountFields := map[string]types.String{
		"private_key_id":              config.PrivateKeyID,
		"client_email":                config.ClientEmail,
		"client_id":                   config.ClientID,
		"auth_uri":                    config.AuthURI,
		"token_uri":                   config.TokenURI,
		"auth_provider_x509_cert_url": config.AuthProviderX509CertURL,
		"client_x509_cert_url":        config.ClientX509CertURL,
	}
	wifFields := map[string]types.String{
		"workload_pool_provider_path":       config.WorkloadPoolProviderPath,
		"service_account_impersonation_url": config.ServiceAccountImpersonationURL,
	}

	switch authType {
	case bigQueryAuthTypeServiceAccountJSON:
		for name, value := range serviceAccountFields {
			if value.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root(name),
					"Missing required attribute",
					fmt.Sprintf("%s is required when auth_type is %s%s.", name, authType, authTypeNote),
				)
			}
		}
		if config.PrivateKey.IsNull() && config.PrivateKeyWo.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("private_key"),
				"Missing required attribute",
				fmt.Sprintf(
					"One of private_key or private_key_wo is required when auth_type is %s%s.",
					authType, authTypeNote,
				),
			)
		}
		for name, value := range wifFields {
			if isSet(value) {
				resp.Diagnostics.AddAttributeError(
					path.Root(name),
					"Invalid attribute for auth_type",
					fmt.Sprintf("%s can only be set when auth_type is %s%s.", name, bigQueryAuthTypeWIF, authTypeNote),
				)
			}
		}

	case bigQueryAuthTypeWIF:
		if config.WorkloadPoolProviderPath.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("workload_pool_provider_path"),
				"Missing required attribute",
				fmt.Sprintf("workload_pool_provider_path is required when auth_type is %s.", authType),
			)
		}
		serviceAccountFields["private_key"] = config.PrivateKey
		serviceAccountFields["private_key_wo"] = config.PrivateKeyWo
		for name, value := range serviceAccountFields {
			if isSet(value) {
				resp.Diagnostics.AddAttributeError(
					path.Root(name),
					"Invalid attribute for auth_type",
					fmt.Sprintf("%s can only be set when auth_type is %s.", name, bigQueryAuthTypeServiceAccountJSON),
				)
			}
		}
		if !config.PrivateKeyWoVersion.IsNull() && !config.PrivateKeyWoVersion.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("private_key_wo_version"),
				"Invalid attribute for auth_type",
				fmt.Sprintf("private_key_wo_version can only be set when auth_type is %s.", bigQueryAuthTypeServiceAccountJSON),
			)
		}
	}
}

// buildConfig returns the config fields for the planned auth_type. The fields of the other auth
// type are left out.
func (r *bigQueryPlatformMetadataCredentialResource) buildConfig(
	plan BigQueryPlatformMetadataCredentialResourceModel,
	privateKey string,
) dbt_cloud.PlatformMetadataCredentialConfig {
	config := dbt_cloud.PlatformMetadataCredentialConfig{
		AuthType: plan.AuthType.ValueString(),
	}

	if plan.AuthType.ValueString() == bigQueryAuthTypeWIF {
		config.WorkloadPoolProviderPath = plan.WorkloadPoolProviderPath.ValueString()
		// ValueString is empty for null and unknown, so omitempty leaves it out and dbt Cloud fills it
		config.ServiceAccountImpersonationURL = plan.ServiceAccountImpersonationURL.ValueString()
		return config
	}

	config.PrivateKey = privateKey
	config.PrivateKeyID = plan.PrivateKeyID.ValueString()
	config.ClientEmail = plan.ClientEmail.ValueString()
	config.ClientID = plan.ClientID.ValueString()
	config.AuthURI = plan.AuthURI.ValueString()
	config.TokenURI = plan.TokenURI.ValueString()
	config.AuthProviderX509CertURL = plan.AuthProviderX509CertURL.ValueString()
	config.ClientX509CertURL = plan.ClientX509CertURL.ValueString()
	return config
}

// setComputedFromResponse fills the values the plan left unknown. Everything else stays as planned.
func setComputedFromResponse(
	plan *BigQueryPlatformMetadataCredentialResourceModel,
	credential *dbt_cloud.PlatformMetadataCredential,
) {
	plan.AdapterVersion = types.StringValue(credential.AdapterVersion)
	if plan.ServiceAccountImpersonationURL.IsUnknown() {
		plan.ServiceAccountImpersonationURL = stringOrNull(credential.Config.ServiceAccountImpersonationURL)
	}
	// write-only values are never stored in state
	plan.PrivateKeyWo = types.StringNull()
}

// addAPIError reports a create or update error. A WIF credential on a bigquery_v0 connection is
// validated as a service account key, so dbt Cloud answers with a 400 listing the missing service
// account fields. The client turns a 400 into "resource-not-found: <body>".
func addAPIError(diags *diag.Diagnostics, action string, authType string, err error) {
	message := err.Error()
	detail := fmt.Sprintf("Could not %s BigQuery platform metadata credential: %s", action, message)
	if authType == bigQueryAuthTypeWIF &&
		strings.HasPrefix(message, "resource-not-found: ") &&
		(strings.Contains(message, "private_key_id") || strings.Contains(message, "client_email")) {
		detail += "\n\nexternal-oauth-wif requires a bigquery_v1 connection " +
			"(use_latest_adapter = true on dbtcloud_global_connection)."
	}
	diags.AddError(fmt.Sprintf("Error %s BigQuery platform metadata credential", gerund(action)), detail)
}

func gerund(action string) string {
	return strings.TrimSuffix(action, "e") + "ing"
}

// isSet reports a known, non-null value. Unknown values are checked again at apply.
func isSet(value types.String) bool {
	return !value.IsNull() && !value.IsUnknown()
}

func stringOrNull(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func (r *bigQueryPlatformMetadataCredentialResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan BigQueryPlatformMetadataCredentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// write-only attributes are only in the config
	var config BigQueryPlatformMetadataCredentialResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	credential := dbt_cloud.PlatformMetadataCredential{
		ConnectionID:            plan.ConnectionID.ValueInt64(),
		CatalogIngestionEnabled: plan.CatalogIngestionEnabled.ValueBool(),
		CostOptimizationEnabled: plan.CostOptimizationEnabled.ValueBool(),
		CostInsightsEnabled:     plan.CostInsightsEnabled.ValueBool(),
		Config: r.buildConfig(
			plan,
			helper.ResolveWriteOnlyString(config.PrivateKeyWo, plan.PrivateKey),
		),
	}

	created, err := r.client.CreatePlatformMetadataCredential(credential)
	if err != nil {
		addAPIError(&resp.Diagnostics, "create", plan.AuthType.ValueString(), err)
		return
	}

	if created.ID == nil {
		resp.Diagnostics.AddError(
			"Error creating BigQuery platform metadata credential",
			"API returned nil ID for created credential",
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%d:%d", r.client.AccountID, *created.ID))
	plan.CredentialID = types.Int64Value(*created.ID)
	setComputedFromResponse(&plan, created)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *bigQueryPlatformMetadataCredentialResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state BigQueryPlatformMetadataCredentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	credential, err := r.client.GetPlatformMetadataCredential(state.CredentialID.ValueInt64())
	if err != nil {
		if helper.HandleResourceNotFound(ctx, err, &resp.Diagnostics, &resp.State, "BigQuery platform metadata credential") {
			return
		}
		resp.Diagnostics.AddError(
			"Error reading BigQuery platform metadata credential",
			"Could not read BigQuery platform metadata credential ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ConnectionID = types.Int64Value(credential.ConnectionID)
	state.CatalogIngestionEnabled = types.BoolValue(credential.CatalogIngestionEnabled)
	state.CostOptimizationEnabled = types.BoolValue(credential.CostOptimizationEnabled)
	state.CostInsightsEnabled = types.BoolValue(credential.CostInsightsEnabled)
	state.AdapterVersion = types.StringValue(credential.AdapterVersion)

	// bigquery_v0 has no auth_type and only supports a service account key
	state.AuthType = types.StringValue(bigQueryAuthTypeServiceAccountJSON)
	if credential.Config.AuthType != "" {
		state.AuthType = types.StringValue(credential.Config.AuthType)
	}

	// the fields of the other auth type are absent from the response, so they become null
	state.PrivateKeyID = stringOrNull(credential.Config.PrivateKeyID)
	state.ClientEmail = stringOrNull(credential.Config.ClientEmail)
	state.ClientID = stringOrNull(credential.Config.ClientID)
	state.AuthURI = stringOrNull(credential.Config.AuthURI)
	state.TokenURI = stringOrNull(credential.Config.TokenURI)
	state.AuthProviderX509CertURL = stringOrNull(credential.Config.AuthProviderX509CertURL)
	state.ClientX509CertURL = stringOrNull(credential.Config.ClientX509CertURL)
	state.WorkloadPoolProviderPath = stringOrNull(credential.Config.WorkloadPoolProviderPath)
	state.ServiceAccountImpersonationURL = stringOrNull(credential.Config.ServiceAccountImpersonationURL)

	// the private key comes back scrubbed, so private_key and private_key_wo_version keep their
	// state values
	state.PrivateKeyWo = types.StringNull()

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bigQueryPlatformMetadataCredentialResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, config, state BigQueryPlatformMetadataCredentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// dbt Cloud merges the PATCH over the stored config, so the stored key is kept when none is
	// sent. A config stored as WIF has no key to merge, so switching to a service account key has
	// to send it.
	privateKey := ""
	switchesToServiceAccount := plan.AuthType.ValueString() == bigQueryAuthTypeServiceAccountJSON &&
		!plan.AuthType.Equal(state.AuthType)
	if !plan.PrivateKey.Equal(state.PrivateKey) ||
		!plan.PrivateKeyWoVersion.Equal(state.PrivateKeyWoVersion) ||
		!plan.PrivateKeyID.Equal(state.PrivateKeyID) ||
		switchesToServiceAccount {
		privateKey = helper.ResolveWriteOnlyString(config.PrivateKeyWo, plan.PrivateKey)
	}

	// connection_id can't change, so it is not sent
	credential := dbt_cloud.PlatformMetadataCredential{
		CatalogIngestionEnabled: plan.CatalogIngestionEnabled.ValueBool(),
		CostOptimizationEnabled: plan.CostOptimizationEnabled.ValueBool(),
		CostInsightsEnabled:     plan.CostInsightsEnabled.ValueBool(),
		Config:                  r.buildConfig(plan, privateKey),
	}

	updated, err := r.client.UpdatePlatformMetadataCredential(state.CredentialID.ValueInt64(), credential)
	if err != nil {
		addAPIError(&resp.Diagnostics, "update", plan.AuthType.ValueString(), err)
		return
	}

	setComputedFromResponse(&plan, updated)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *bigQueryPlatformMetadataCredentialResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state BigQueryPlatformMetadataCredentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePlatformMetadataCredential(state.CredentialID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting BigQuery platform metadata credential",
			"Could not delete BigQuery platform metadata credential: "+err.Error(),
		)
	}
}

func (r *bigQueryPlatformMetadataCredentialResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// accepts "credential_id" and "account_id:credential_id"
	idParts := strings.Split(req.ID, ":")
	if len(idParts) > 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format: credential_id or account_id:credential_id. Got: %s", req.ID),
		)
		return
	}

	rawCredentialID := idParts[len(idParts)-1]
	credentialID, err := strconv.Atoi(rawCredentialID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Could not parse credential_id as integer: %s", rawCredentialID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("id"),
		fmt.Sprintf("%d:%d", r.client.AccountID, credentialID),
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("credential_id"),
		credentialID,
	)...)
}
