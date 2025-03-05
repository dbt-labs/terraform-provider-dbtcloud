package model_notifications

import (
	"context"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &modelNotificationsResource{}
	_ resource.ResourceWithConfigure   = &modelNotificationsResource{}
	_ resource.ResourceWithImportState = &modelNotificationsResource{}
)

func ModelNotificationsResource() resource.Resource {
	return &modelNotificationsResource{}
}

type modelNotificationsResource struct {
	client *dbt_cloud.Client
}

func (r *modelNotificationsResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_model_notifications"
}

func (r *modelNotificationsResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data ModelNotificationsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	environmentID := data.EnvironmentID.ValueString()
	modelNotifications, err := r.client.GetModelNotifications(environmentID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The model notifications resource was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error getting the model notifications", err.Error())
		return
	}

	data.ID = types.StringValue(strconv.Itoa(*modelNotifications.ID))
	data.EnvironmentID = types.StringValue(strconv.Itoa(modelNotifications.EnvironmentID))
	data.Enabled = types.BoolValue(modelNotifications.Enabled)
	data.OnSuccess = types.BoolValue(modelNotifications.OnSuccess)
	data.OnFailure = types.BoolValue(modelNotifications.OnFailure)
	data.OnWarning = types.BoolValue(modelNotifications.OnWarning)
	data.OnSkipped = types.BoolValue(modelNotifications.OnSkipped)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *modelNotificationsResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data ModelNotificationsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	environmentID := data.EnvironmentID.ValueString()
	modelNotifications, err := r.client.CreateModelNotifications(
		environmentID,
		data.Enabled.ValueBool(),
		data.OnSuccess.ValueBool(),
		data.OnFailure.ValueBool(),
		data.OnWarning.ValueBool(),
		data.OnSkipped.ValueBool(),
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create model notifications",
			"Error: "+err.Error(),
		)
		return
	}

	data.ID = types.StringValue(strconv.Itoa(*modelNotifications.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *modelNotificationsResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state ModelNotificationsResourceModel

	// Read plan and state values into the models
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update the model with the plan values
	modelNotifications := ConvertModelNotificationsModelToData(plan)

	environmentID := plan.EnvironmentID.ValueString()
	updatedModelNotifications, err := r.client.UpdateModelNotifications(environmentID, modelNotifications)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update model notifications",
			"Error: "+err.Error(),
		)
		return
	}

	// Update the state with the response values
	state.Enabled = types.BoolValue(updatedModelNotifications.Enabled)
	state.OnSuccess = types.BoolValue(updatedModelNotifications.OnSuccess)
	state.OnFailure = types.BoolValue(updatedModelNotifications.OnFailure)
	state.OnWarning = types.BoolValue(updatedModelNotifications.OnWarning)
	state.OnSkipped = types.BoolValue(updatedModelNotifications.OnSkipped)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *modelNotificationsResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// Model notifications cannot be deleted, they can only be disabled
	var data ModelNotificationsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Disable the model notifications
	environmentID := data.EnvironmentID.ValueString()
	modelNotifications := ConvertModelNotificationsModelToData(data)
	modelNotifications.Enabled = false

	_, err := r.client.UpdateModelNotifications(environmentID, modelNotifications)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to disable model notifications",
			"Error: "+err.Error(),
		)
		return
	}
}

func (r *modelNotificationsResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// The import ID is the environment ID
	environmentID := req.ID

	// Set the environment_id attribute
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), environmentID)...)

	// Read the resource to populate the rest of the state
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *modelNotificationsResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}
