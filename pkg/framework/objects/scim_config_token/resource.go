package scim_config_token

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
	_ resource.Resource                = &scimConfigTokenResource{}
	_ resource.ResourceWithConfigure   = &scimConfigTokenResource{}
	_ resource.ResourceWithImportState = &scimConfigTokenResource{}
)

func SCIMConfigTokenResource() resource.Resource {
	return &scimConfigTokenResource{}
}

type scimConfigTokenResource struct {
	client *dbt_cloud.Client
}

func (r *scimConfigTokenResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_scim_config_token"
}

func (r *scimConfigTokenResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*dbt_cloud.Client)
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *scimConfigTokenResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SCIMConfigTokenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateSCIMConfigToken(plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create SCIM config token", "Error: "+err.Error())
		return
	}

	plan.ID = types.Int64Value(*created.ID)
	plan.CreatedAt = types.StringValue(created.CreatedAt)
	plan.LastUsed = types.StringNull()

	// token_string is only returned on creation — store it now.
	if created.TokenString != nil {
		plan.TokenString = types.StringValue(*created.TokenString)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *scimConfigTokenResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state SCIMConfigTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	token, err := r.client.GetSCIMConfigToken(state.ID.ValueInt64())
	if err != nil {
		if helper.HandleResourceNotFound(ctx, err, &resp.Diagnostics, &resp.State, "SCIM config token") {
			return
		}
		resp.Diagnostics.AddError("Error getting SCIM config token", err.Error())
		return
	}

	state.Name = types.StringValue(token.Name)
	state.CreatedAt = types.StringValue(token.CreatedAt)

	if token.LastUsed != nil {
		state.LastUsed = types.StringValue(*token.LastUsed)
	} else {
		state.LastUsed = types.StringNull()
	}

	// token_string is never returned by the API after creation.
	// Leave the value already in state untouched.

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Update ────────────────────────────────────────────────────────────────────
//
// The SCIM token API has no update endpoint. The name attribute is marked
// RequiresReplace in the schema, so Terraform will destroy and recreate the
// resource if the name changes. This method should never be called.

func (r *scimConfigTokenResource) Update(
	_ context.Context,
	_ resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"SCIM config tokens cannot be updated in place. Changing the name forces a new resource.",
	)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *scimConfigTokenResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state SCIMConfigTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSCIMConfigToken(state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Issue deleting SCIM config token", "Error: "+err.Error())
	}
}

// ── Import ────────────────────────────────────────────────────────────────────
//
// Import by numeric token ID. Note: token_string will be empty after import
// since the API never returns it — this is expected and documented.

func (r *scimConfigTokenResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing SCIM config token ID for import", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
