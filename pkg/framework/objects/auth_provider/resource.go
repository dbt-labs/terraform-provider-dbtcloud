package auth_provider

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = &authProviderResource{}
	_ resource.ResourceWithConfigure      = &authProviderResource{}
	_ resource.ResourceWithImportState    = &authProviderResource{}
	_ resource.ResourceWithValidateConfig = &authProviderResource{}
)

func AuthProviderResource() resource.Resource {
	return &authProviderResource{}
}

type authProviderResource struct {
	client *dbt_cloud.Client
}

func (r *authProviderResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_auth_provider"
}

func (r *authProviderResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*dbt_cloud.Client)
}

// ── ValidateConfig ───────────────────────────────────────────────────────────

func (r *authProviderResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var data AuthProviderResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	providerType := data.Type.ValueString()

	isSAML := providerType == "saml" || providerType == "okta"
	isAzure := providerType == "azure_single_tenant" ||
		providerType == "azure_multi_tenant" ||
		providerType == "azure_active_directory"
	isGSuite := providerType == "gsuite"

	switch {
	case isSAML:
		if data.EntityID.IsNull() || data.EntityID.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("entity_id"),
				"Missing required attribute",
				"`entity_id` is required for provider type `"+providerType+"`.",
			)
		}
		if data.SsoURL.IsNull() || data.SsoURL.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("sso_url"),
				"Missing required attribute",
				"`sso_url` is required for provider type `"+providerType+"`.",
			)
		}
		if data.Cert.IsNull() && data.CertWo.IsNull() {
			resp.Diagnostics.AddError(
				"Missing required attribute",
				"One of `cert` or `cert_wo` is required for provider type `"+providerType+"`.",
			)
		}

	case isAzure:
		if data.ClientID.IsNull() || data.ClientID.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("client_id"),
				"Missing required attribute",
				"`client_id` is required for provider type `"+providerType+"`.",
			)
		}
		if data.ClientSecret.IsNull() && data.ClientSecretWo.IsNull() {
			resp.Diagnostics.AddError(
				"Missing required attribute",
				"One of `client_secret` or `client_secret_wo` is required for provider type `"+providerType+"`.",
			)
		}
		if providerType == "azure_single_tenant" && (data.TenantID.IsNull() || data.TenantID.ValueString() == "") {
			resp.Diagnostics.AddAttributeError(
				path.Root("tenant_id"),
				"Missing required attribute",
				"`tenant_id` is required for provider type `azure_single_tenant`.",
			)
		}

	case isGSuite:
		if data.ClientID.IsNull() || data.ClientID.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("client_id"),
				"Missing required attribute",
				"`client_id` is required for provider type `gsuite`.",
			)
		}
		if data.ClientSecret.IsNull() && data.ClientSecretWo.IsNull() {
			resp.Diagnostics.AddError(
				"Missing required attribute",
				"One of `client_secret` or `client_secret_wo` is required for provider type `gsuite`.",
			)
		}
		if data.AdminRefreshToken.IsNull() || data.AdminRefreshToken.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("admin_refresh_token"),
				"Missing required attribute",
				"`admin_refresh_token` is required for provider type `gsuite`.",
			)
		}
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *authProviderResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan AuthProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Config is needed to access write-only attributes (cert_wo, client_secret_wo).
	var config AuthProviderResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ap := planToAuthProvider(plan, config)

	created, err := r.client.CreateAuthProvider(ap)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create auth provider", "Error: "+err.Error())
		return
	}

	applyAPIResponse(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *authProviderResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state AuthProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ap, err := r.client.GetAuthProvider(state.ID.ValueInt64())
	if err != nil {
		if helper.HandleResourceNotFound(ctx, err, &resp.Diagnostics, &resp.State, "auth provider") {
			return
		}
		resp.Diagnostics.AddError("Error getting auth provider", err.Error())
		return
	}

	applyAPIResponse(ap, &state)
	// Secrets (cert, client_secret, tenant_id, admin_refresh_token) are never
	// returned by the API in plaintext, so we leave them as-is in state.

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *authProviderResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state AuthProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Config needed for write-only attribute access.
	var config AuthProviderResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch current server state to build the update payload.
	current, err := r.client.GetAuthProvider(state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error getting auth provider for update", err.Error())
		return
	}

	// Apply plan changes onto the fetched payload so unchanged fields are preserved.
	current.AllowPasswordBackdoor = plan.AllowPasswordBackdoor.ValueBool()
	if !plan.Slug.IsNull() {
		current.Slug = plan.Slug.ValueString()
	}

	providerType := plan.Type.ValueString()
	isSAML := providerType == "saml" || providerType == "okta"
	isAzure := providerType == "azure_single_tenant" ||
		providerType == "azure_multi_tenant" ||
		providerType == "azure_active_directory"

	switch {
	case isSAML:
		current.EntityID = stringPtrOrNil(plan.EntityID)
		current.SsoURL = stringPtrOrNil(plan.SsoURL)
		if plan.Cert != state.Cert || plan.CertWoVersion != state.CertWoVersion {
			s := helper.ResolveWriteOnlyString(config.CertWo, plan.Cert)
			current.Cert = &s
		}
		signRequest := plan.SignRequest.ValueBool()
		current.SignRequest = &signRequest
		attrMap := plan.AttributeMap.ValueString()
		current.AttributeMap = &attrMap

	case isAzure:
		current.ClientID = stringPtrOrNil(plan.ClientID)
		if plan.ClientSecret != state.ClientSecret || plan.ClientSecretWoVersion != state.ClientSecretWoVersion {
			s := helper.ResolveWriteOnlyString(config.ClientSecretWo, plan.ClientSecret)
			current.ClientSecret = &s
		}
		current.TenantID = stringPtrOrNil(plan.TenantID)
		current.Domain = stringPtrOrNil(plan.Domain)
		includeIndirect := plan.IncludeIndirectGroups.ValueBool()
		current.IncludeIndirectGroups = &includeIndirect
		maxGroups := int(plan.MaxGroupsToRetrieve.ValueInt64())
		current.MaxGroupsToRetrieve = &maxGroups

	default: // gsuite
		current.ClientID = stringPtrOrNil(plan.ClientID)
		if plan.ClientSecret != state.ClientSecret || plan.ClientSecretWoVersion != state.ClientSecretWoVersion {
			s := helper.ResolveWriteOnlyString(config.ClientSecretWo, plan.ClientSecret)
			current.ClientSecret = &s
		}
		current.Domain = stringPtrOrNil(plan.Domain)
		current.GsuiteAdminID = stringPtrOrNil(plan.GsuiteAdminID)
		current.AuthorizationURL = stringPtrOrNil(plan.AuthorizationURL)
		if !plan.AdminRefreshToken.IsNull() {
			current.AdminRefreshToken = stringPtrOrNil(plan.AdminRefreshToken)
		}
	}

	updated, err := r.client.UpdateAuthProvider(state.ID.ValueInt64(), *current)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update auth provider", "Error: "+err.Error())
		return
	}

	applyAPIResponse(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *authProviderResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state AuthProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAuthProvider(state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Issue deleting auth provider", "Error: "+err.Error())
	}
}

// ── Import ────────────────────────────────────────────────────────────────────

func (r *authProviderResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing auth provider ID for import", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// planToAuthProvider builds the API payload from Terraform plan + config.
// Config is required to access write-only attribute values.
func planToAuthProvider(plan, config AuthProviderResourceModel) dbt_cloud.AuthProvider {
	ap := dbt_cloud.AuthProvider{
		Type:                  plan.Type.ValueString(),
		AllowPasswordBackdoor: plan.AllowPasswordBackdoor.ValueBool(),
	}
	if !plan.Slug.IsNull() {
		ap.Slug = plan.Slug.ValueString()
	}

	providerType := plan.Type.ValueString()
	isSAML := providerType == "saml" || providerType == "okta"
	isAzure := providerType == "azure_single_tenant" ||
		providerType == "azure_multi_tenant" ||
		providerType == "azure_active_directory"

	switch {
	case isSAML:
		ap.EntityID = stringPtrOrNil(plan.EntityID)
		ap.SsoURL = stringPtrOrNil(plan.SsoURL)
		cert := helper.ResolveWriteOnlyString(config.CertWo, plan.Cert)
		ap.Cert = &cert
		signRequest := plan.SignRequest.ValueBool()
		ap.SignRequest = &signRequest
		attrMap := plan.AttributeMap.ValueString()
		ap.AttributeMap = &attrMap

	case isAzure:
		ap.ClientID = stringPtrOrNil(plan.ClientID)
		clientSecret := helper.ResolveWriteOnlyString(config.ClientSecretWo, plan.ClientSecret)
		ap.ClientSecret = &clientSecret
		ap.TenantID = stringPtrOrNil(plan.TenantID)
		ap.Domain = stringPtrOrNil(plan.Domain)
		includeIndirect := plan.IncludeIndirectGroups.ValueBool()
		ap.IncludeIndirectGroups = &includeIndirect
		maxGroups := int(plan.MaxGroupsToRetrieve.ValueInt64())
		ap.MaxGroupsToRetrieve = &maxGroups

	default: // gsuite
		ap.ClientID = stringPtrOrNil(plan.ClientID)
		clientSecret := helper.ResolveWriteOnlyString(config.ClientSecretWo, plan.ClientSecret)
		ap.ClientSecret = &clientSecret
		ap.Domain = stringPtrOrNil(plan.Domain)
		ap.GsuiteAdminID = stringPtrOrNil(plan.GsuiteAdminID)
		ap.AuthorizationURL = stringPtrOrNil(plan.AuthorizationURL)
		ap.AdminRefreshToken = stringPtrOrNil(plan.AdminRefreshToken)
	}

	return ap
}

// applyAPIResponse maps non-secret API response fields onto the Terraform model.
// Secret fields (cert, client_secret, tenant_id, admin_refresh_token) are intentionally
// omitted because the API never returns them in plaintext.
func applyAPIResponse(ap *dbt_cloud.AuthProvider, m *AuthProviderResourceModel) {
	if ap.ID != nil {
		m.ID = types.Int64Value(*ap.ID)
	}
	m.State = types.Int64Value(int64(ap.State))
	m.Type = types.StringValue(ap.Type)
	m.Slug = types.StringValue(ap.Slug)
	m.AllowPasswordBackdoor = types.BoolValue(ap.AllowPasswordBackdoor)
	m.LoginURL = types.StringValue(ap.LoginURL)
	m.CreatedAt = types.StringValue(ap.CreatedAt)
	m.UpdatedAt = types.StringValue(ap.UpdatedAt)

	if ap.CertExpiryDate != nil {
		m.CertExpiryDate = types.StringValue(*ap.CertExpiryDate)
	} else {
		m.CertExpiryDate = types.StringNull()
	}

	providerType := ap.Type
	isSAML := providerType == "saml" || providerType == "okta"
	isAzure := providerType == "azure_single_tenant" ||
		providerType == "azure_multi_tenant" ||
		providerType == "azure_active_directory"

	switch {
	case isSAML:
		if ap.EntityID != nil {
			m.EntityID = types.StringValue(*ap.EntityID)
		} else {
			m.EntityID = types.StringNull()
		}
		if ap.SsoURL != nil {
			m.SsoURL = types.StringValue(*ap.SsoURL)
		} else {
			m.SsoURL = types.StringNull()
		}
		if ap.SignRequest != nil {
			m.SignRequest = types.BoolValue(*ap.SignRequest)
		}
		if ap.AttributeMap != nil {
			m.AttributeMap = types.StringValue(normalizeJSON(*ap.AttributeMap))
		} else {
			m.AttributeMap = types.StringNull()
		}
		// Not applicable to SAML — must be set to a known value.
		m.AuthorizationURL = types.StringNull()

	case isAzure:
		// client_id, client_secret, tenant_id are EncryptedTextField — the API
		// never returns them in plaintext. Leave them untouched in the model so
		// the user-configured values are preserved in state across reads.
		if ap.Domain != nil {
			m.Domain = types.StringValue(*ap.Domain)
		} else {
			m.Domain = types.StringNull()
		}
		if ap.IncludeIndirectGroups != nil {
			m.IncludeIndirectGroups = types.BoolValue(*ap.IncludeIndirectGroups)
		}
		if ap.MaxGroupsToRetrieve != nil {
			m.MaxGroupsToRetrieve = types.Int64Value(int64(*ap.MaxGroupsToRetrieve))
		}
		// Not applicable to Azure — must be set to a known value.
		m.AuthorizationURL = types.StringNull()

	default: // gsuite
		// client_id, client_secret, admin_refresh_token are EncryptedTextField —
		// the API never returns them in plaintext. Leave them untouched in the model.
		if ap.Domain != nil {
			m.Domain = types.StringValue(*ap.Domain)
		} else {
			m.Domain = types.StringNull()
		}
		if ap.GsuiteAdminID != nil {
			m.GsuiteAdminID = types.StringValue(*ap.GsuiteAdminID)
		} else {
			m.GsuiteAdminID = types.StringNull()
		}
		// authorization_url may be auto-populated server-side (e.g. by Auth0).
		// Always sync it from the API response so state stays consistent.
		if ap.AuthorizationURL != nil {
			m.AuthorizationURL = types.StringValue(*ap.AuthorizationURL)
		} else {
			m.AuthorizationURL = types.StringNull()
		}
	}
}

// normalizeJSON round-trips a JSON string through unmarshal/marshal so that
// key order and whitespace are canonical. Go's json.Marshal sorts map keys
// alphabetically, matching what Terraform's jsonencode() produces.
// Returns the original string unchanged if it is not valid JSON.
func normalizeJSON(s string) string {
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return s
	}
	b, err := json.Marshal(v)
	if err != nil {
		return s
	}
	return string(b)
}

func stringPtrOrNil(v types.String) *string {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		return nil
	}
	s := v.ValueString()
	return &s
}
