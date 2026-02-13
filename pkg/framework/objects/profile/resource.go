package profile

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &profileResource{}
	_ resource.ResourceWithConfigure   = &profileResource{}
	_ resource.ResourceWithImportState = &profileResource{}
)

func ProfileResource() resource.Resource {
	return &profileResource{}
}

type profileResource struct {
	client *dbt_cloud.Client
}

func (r *profileResource) Configure(
	ctx context.Context,
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

func (r *profileResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_profile"
}

func (r *profileResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resourceSchema
}

func (r *profileResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be in the format 'project_id:profile_id'",
		)
		return
	}

	projectID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid project ID",
			"Project ID must be a valid integer",
		)
		return
	}

	profileID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid profile ID",
			"Profile ID must be a valid integer",
		)
		return
	}

	resp.State.SetAttribute(
		ctx,
		path.Root("id"),
		fmt.Sprintf("%d%s%d", projectID, dbt_cloud.ID_DELIMITER, profileID),
	)
	resp.State.SetAttribute(ctx, path.Root("project_id"), projectID)
	resp.State.SetAttribute(ctx, path.Root("profile_id"), profileID)
}

func (r *profileResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan ProfileResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(plan.ProjectID.ValueInt64())
	key := plan.Key.ValueString()
	connectionID := int(plan.ConnectionID.ValueInt64())
	credentialsID := int(plan.CredentialsID.ValueInt64())

	var extendedAttributesID *int
	if !plan.ExtendedAttributesID.IsNull() && !plan.ExtendedAttributesID.IsUnknown() {
		v := int(plan.ExtendedAttributesID.ValueInt64())
		extendedAttributesID = &v
	}

	profile, err := r.client.CreateProfile(
		projectID,
		key,
		connectionID,
		credentialsID,
		extendedAttributesID,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating profile",
			"Could not create profile, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf(
		"%d%s%d",
		profile.ProjectID,
		dbt_cloud.ID_DELIMITER,
		*profile.ID,
	))
	plan.ProfileID = types.Int64Value(int64(*profile.ID))

	if profile.ExtendedAttributesID != nil {
		plan.ExtendedAttributesID = types.Int64Value(int64(*profile.ExtendedAttributesID))
	} else {
		plan.ExtendedAttributesID = types.Int64Null()
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *profileResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state ProfileResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID, profileID, err := helper.SplitIDToInts(
		state.ID.ValueString(),
		"dbtcloud_profile",
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing profile ID",
			"Could not parse profile ID: "+err.Error(),
		)
		return
	}

	profile, err := r.client.GetProfile(projectID, profileID)
	if err != nil {
		if strings.Contains(err.Error(), "resource-not-found") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading profile",
			"Could not read profile ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ProfileID = types.Int64Value(int64(*profile.ID))
	state.ProjectID = types.Int64Value(int64(profile.ProjectID))
	state.Key = types.StringValue(profile.Key)
	state.ConnectionID = types.Int64Value(int64(profile.ConnectionID))
	state.CredentialsID = types.Int64Value(int64(profile.CredentialsID))

	if profile.ExtendedAttributesID != nil {
		state.ExtendedAttributesID = types.Int64Value(int64(*profile.ExtendedAttributesID))
	} else {
		state.ExtendedAttributesID = types.Int64Null()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *profileResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan ProfileResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ProfileResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID, profileID, err := helper.SplitIDToInts(
		state.ID.ValueString(),
		"dbtcloud_profile",
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing profile ID",
			"Could not parse profile ID: "+err.Error(),
		)
		return
	}

	var extendedAttributesID *int
	if !plan.ExtendedAttributesID.IsNull() && !plan.ExtendedAttributesID.IsUnknown() {
		v := int(plan.ExtendedAttributesID.ValueInt64())
		extendedAttributesID = &v
	}

	updateProfile := dbt_cloud.Profile{
		AccountID:            r.client.AccountID,
		ProjectID:            int(plan.ProjectID.ValueInt64()),
		Key:                  plan.Key.ValueString(),
		ConnectionID:         int(plan.ConnectionID.ValueInt64()),
		CredentialsID:        int(plan.CredentialsID.ValueInt64()),
		ExtendedAttributesID: extendedAttributesID,
	}

	_, err = r.client.UpdateProfile(projectID, profileID, updateProfile)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating profile",
			"Could not update profile ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *profileResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state ProfileResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID, profileID, err := helper.SplitIDToInts(
		state.ID.ValueString(),
		"dbtcloud_profile",
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing profile ID",
			"Could not parse profile ID: "+err.Error(),
		)
		return
	}

	_, err = r.client.DeleteProfile(projectID, profileID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting profile",
			"Could not delete profile, unexpected error: "+err.Error(),
		)
		return
	}
}
