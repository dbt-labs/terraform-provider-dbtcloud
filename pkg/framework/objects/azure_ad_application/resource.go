package azure_ad_application

import (
	"context"
	"strings"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &azureADApplicationResource{}
	_ resource.ResourceWithConfigure   = &azureADApplicationResource{}
	_ resource.ResourceWithImportState = &azureADApplicationResource{}
)

func AzureADApplicationResource() resource.Resource {
	return &azureADApplicationResource{}
}

type azureADApplicationResource struct {
	client *dbt_cloud.Client
}

func (r *azureADApplicationResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_azure_ad_application"
}

func (r *azureADApplicationResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*dbt_cloud.Client)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// modelToApp converts the Terraform model to the API struct.
// client_id, client_secret, tenant_id are always included because the API
// requires all four fields on both create and update.
func modelToApp(m AzureADApplicationResourceModel) dbt_cloud.AzureADApplication {
	clientID := m.ClientID.ValueString()
	clientSecret := m.ClientSecret.ValueString()
	tenantID := m.TenantID.ValueString()

	return dbt_cloud.AzureADApplication{
		OrganizationName:                 m.OrganizationName.ValueString(),
		ClientID:                         &clientID,
		ClientSecret:                     &clientSecret,
		TenantID:                         &tenantID,
		AzureServiceAuthenticationMethod: m.AzureServiceAuthenticationMethod.ValueString(),
	}
}

// applyAPIResponse maps the API response back onto the model.
// client_id, client_secret, tenant_id are never returned by the API — the
// values already in state are preserved.
func applyAPIResponse(api *dbt_cloud.AzureADApplication, m *AzureADApplicationResourceModel) {
	m.ID = types.Int64Value(*api.ID)
	m.AccountID = types.Int64Value(api.AccountID)
	m.OrganizationName = types.StringValue(api.OrganizationName)
	m.AzureServiceAuthenticationMethod = types.StringValue(api.AzureServiceAuthenticationMethod)

	if api.OAuthRedirectURIDomain != nil {
		m.OAuthRedirectURIDomain = types.StringValue(*api.OAuthRedirectURIDomain)
	} else {
		m.OAuthRedirectURIDomain = types.StringNull()
	}
	if api.CreatedAt != nil {
		m.CreatedAt = types.StringValue(*api.CreatedAt)
	}
	if api.UpdatedAt != nil {
		m.UpdatedAt = types.StringValue(*api.UpdatedAt)
	}
	// client_id, client_secret, tenant_id: leave untouched — API never returns them.
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *azureADApplicationResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan AzureADApplicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// One Azure AD application is allowed per account (unique constraint).
	// If one already exists, adopt it and update rather than failing.
	existing, err := r.client.GetAzureADApplicationForAccount()
	if err != nil {
		resp.Diagnostics.AddError("Unable to check existing Azure AD application", err.Error())
		return
	}

	var result *dbt_cloud.AzureADApplication
	if existing != nil {
		// Adopt the existing record and update it with the desired config.
		result, err = r.client.UpdateAzureADApplication(*existing.ID, modelToApp(plan))
		if err != nil {
			resp.Diagnostics.AddError("Unable to update existing Azure AD application", err.Error())
			return
		}
	} else {
		result, err = r.client.CreateAzureADApplication(modelToApp(plan))
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value") {
				resp.Diagnostics.AddError(
					"Unable to create Azure AD application",
					"An Azure AD application already exists for this dbt Cloud account but is "+
						"no longer visible via the API (it was soft-deleted by a previous destroy). "+
						"To recover:\n"+
						"  1. Contact dbt Cloud support to remove the orphaned record for "+
						"account ID "+strconv.FormatInt(r.client.AccountID, 10)+".\n"+
						"  2. Alternatively, if you know the record ID, re-adopt it with:\n"+
						"     terraform import dbtcloud_azure_ad_application.<name> <id>",
				)
			} else {
				resp.Diagnostics.AddError("Unable to create Azure AD application", err.Error())
			}
			return
		}
	}

	applyAPIResponse(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *azureADApplicationResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state AzureADApplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.GetAzureADApplication(state.ID.ValueInt64())
	if err != nil {
		if helper.HandleResourceNotFound(ctx, err, &resp.Diagnostics, &resp.State, "Azure AD application") {
			return
		}
		resp.Diagnostics.AddError("Error reading Azure AD application", err.Error())
		return
	}

	applyAPIResponse(app, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *azureADApplicationResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan AzureADApplicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	var state AzureADApplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateAzureADApplication(state.ID.ValueInt64(), modelToApp(plan))
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Azure AD application", err.Error())
		return
	}

	applyAPIResponse(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *azureADApplicationResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state AzureADApplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAzureADApplication(state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Unable to delete Azure AD application", err.Error())
	}
}

// ── Import ────────────────────────────────────────────────────────────────────
//
// Import by numeric application ID. Note: client_id, client_secret, and
// tenant_id will be empty after import — the API never returns them. You must
// set them in config to avoid drift on the next apply.

func (r *azureADApplicationResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing Azure AD application ID for import", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
