package scim_config

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &scimConfigResource{}
	_ resource.ResourceWithConfigure   = &scimConfigResource{}
	_ resource.ResourceWithImportState = &scimConfigResource{}
)

func SCIMConfigResource() resource.Resource {
	return &scimConfigResource{}
}

type scimConfigResource struct {
	client *dbt_cloud.Client
}

func (r *scimConfigResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_scim_config"
}

func (r *scimConfigResource) Configure(
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

func (r *scimConfigResource) applyAPIResponse(cfg *dbt_cloud.SCIMConfig, m *SCIMConfigResourceModel) {
	m.ID = types.StringValue(fmt.Sprintf("%d", r.client.AccountID))
	m.Enabled = types.BoolValue(cfg.Enabled)
	m.ManualUpdatesAllowed = types.BoolValue(cfg.ManualUpdatesAllowed)
	m.SCIMControlledLicenseType = types.BoolValue(cfg.SCIMControlledLicenseType)
}

func modelToSCIMConfig(m SCIMConfigResourceModel) dbt_cloud.SCIMConfig {
	return dbt_cloud.SCIMConfig{
		Enabled:                   m.Enabled.ValueBool(),
		ManualUpdatesAllowed:      m.ManualUpdatesAllowed.ValueBool(),
		SCIMControlledLicenseType: m.SCIMControlledLicenseType.ValueBool(),
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *scimConfigResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SCIMConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.UpdateSCIMConfig(modelToSCIMConfig(plan))
	if err != nil {
		resp.Diagnostics.AddError("Unable to configure SCIM", err.Error())
		return
	}

	r.applyAPIResponse(cfg, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *scimConfigResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state SCIMConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.GetSCIMConfig()
	if err != nil {
		resp.Diagnostics.AddError("Error reading SCIM config", err.Error())
		return
	}

	r.applyAPIResponse(cfg, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *scimConfigResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan SCIMConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.UpdateSCIMConfig(modelToSCIMConfig(plan))
	if err != nil {
		resp.Diagnostics.AddError("Unable to update SCIM config", err.Error())
		return
	}

	r.applyAPIResponse(cfg, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────
//
// The SCIM config endpoint has no DELETE. On destroy we reset to safe defaults:
// SCIM disabled, manual updates allowed, license type not SCIM-controlled.

func (r *scimConfigResource) Delete(
	_ context.Context,
	_ resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	_, err := r.client.UpdateSCIMConfig(dbt_cloud.SCIMConfig{
		Enabled:                   false,
		ManualUpdatesAllowed:      false,
		SCIMControlledLicenseType: false,
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to reset SCIM config on destroy", err.Error())
	}
}

// ── Import ────────────────────────────────────────────────────────────────────
//
// Singleton resource — import with any non-empty string (conventionally the
// account ID). The actual state is fetched from the API during the subsequent
// Read.

func (r *scimConfigResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...,
	)
}
