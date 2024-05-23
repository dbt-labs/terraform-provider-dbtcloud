package notification

import (
	"context"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &notificationResource{}
	_ resource.ResourceWithConfigure   = &notificationResource{}
	_ resource.ResourceWithImportState = &notificationResource{}
)

func NotificationResource() resource.Resource {
	return &notificationResource{}
}

type notificationResource struct {
	client *dbt_cloud.Client
}

func (r *notificationResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification"
}

func (r *notificationResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var data NotificationResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.NotificationType == types.Int64Value(1) &&
		!(data.ExternalEmail.IsNull() && data.SlackChannelID.IsNull() && data.SlackChannelName.IsNull()) {
		resp.Diagnostics.AddAttributeError(
			path.Root("notification_type"),
			"Notification type is not compatible with the other attributes",
			"Notification type 1 is for internal notifications only. Please remove the external email, slack channel ID, and slack channel name attributes.",
		)
	}

	if data.NotificationType == types.Int64Value(2) &&
		data.SlackChannelID.IsNull() &&
		data.SlackChannelName.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("notification_type"),
			"Notification type is not compatible with the other attributes",
			"Notification type 2 requires a Slack channel ID and Slack channel name.",
		)
	}

	if data.NotificationType == types.Int64Value(4) && data.ExternalEmail.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("notification_type"),
			"Notification type is not compatible with the other attributes",
			"Notification type 4 requires an external email.",
		)
	}
}

func (r *notificationResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	notificationID := data.ID.ValueString()
	notification, err := r.client.GetNotification(notificationID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The notification resource was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error getting the notification", err.Error())
		return
	}

	data.ID = types.StringValue(notificationID)
	data.UserID = types.Int64Value(int64(notification.UserId))
	data.OnCancel, _ = types.SetValueFrom(
		context.Background(),
		types.Int64Type,
		notification.OnCancel,
	)
	data.OnFailure, _ = types.SetValueFrom(
		context.Background(),
		types.Int64Type,
		notification.OnFailure,
	)
	data.OnSuccess, _ = types.SetValueFrom(
		context.Background(),
		types.Int64Type,
		notification.OnSuccess,
	)
	data.State = types.Int64Value(int64(notification.State))
	data.NotificationType = types.Int64Value(int64(notification.NotificationType))
	data.ExternalEmail = types.StringPointerValue(notification.ExternalEmail)
	data.SlackChannelID = types.StringPointerValue(notification.SlackChannelID)
	data.SlackChannelName = types.StringPointerValue(notification.SlackChannelName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r *notificationResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var intOnCancel, intOnFailure, intOnSuccess []int

	diags := data.OnCancel.ElementsAs(context.Background(), &intOnCancel, false)
	if diags.HasError() {
		return
	}
	diags = data.OnFailure.ElementsAs(context.Background(), &intOnFailure, false)
	if diags.HasError() {
		return
	}

	diags = data.OnSuccess.ElementsAs(context.Background(), &intOnSuccess, false)
	if diags.HasError() {
		return
	}

	notif, err := r.client.CreateNotification(
		int(data.UserID.ValueInt64()),
		intOnCancel,
		intOnFailure,
		intOnSuccess,
		int(data.State.ValueInt64()),
		int(data.NotificationType.ValueInt64()),
		data.ExternalEmail.ValueStringPointer(),
		data.SlackChannelID.ValueStringPointer(),
		data.SlackChannelName.ValueStringPointer(),
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create notification",
			"Error: "+err.Error(),
		)
		return
	}

	data.ID = types.StringValue(strconv.Itoa(*notif.Id))

	// TODO; Revise later maybe. We are saving the config instead of the value we get back from the notification call
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *notificationResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	notificationID := data.ID.ValueString()
	notification, err := r.client.GetNotification(notificationID)
	if err != nil {
		resp.Diagnostics.AddError("Error getting the notification", err.Error())
		return
	}

	notification.State = dbt_cloud.STATE_DELETED
	_, err = r.client.UpdateNotification(notificationID, *notification)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting the notification", err.Error())
		return
	}
}

func (r *notificationResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state NotificationResourceModel

	// Read plan and state values into the models
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if plan.UserID != state.UserID {
		state.UserID = plan.UserID
	}

	if !plan.OnCancel.Equal(state.OnCancel) {
		state.OnCancel = plan.OnCancel
	}

	if !plan.OnFailure.Equal(state.OnFailure) {
		state.OnFailure = plan.OnFailure
	}

	if !plan.OnSuccess.Equal(state.OnSuccess) {
		state.OnSuccess = plan.OnSuccess
	}

	if plan.State != state.State {
		state.State = plan.State
	}

	if plan.NotificationType != state.NotificationType {
		state.NotificationType = plan.NotificationType
	}

	if plan.ExternalEmail != state.ExternalEmail {
		state.ExternalEmail = plan.ExternalEmail
	}

	if plan.SlackChannelID != state.SlackChannelID {
		state.SlackChannelID = plan.SlackChannelID
	}

	if plan.SlackChannelName != state.SlackChannelName {
		state.SlackChannelName = plan.SlackChannelName
	}

	notification := ConvertStateToNotification(state)
	notification.AccountId = r.client.AccountID

	// Update the notification
	_, err := r.client.UpdateNotification(state.ID.ValueString(), notification)
	if err != nil {
		resp.Diagnostics.AddError("Error updating notification", err.Error())
		return
	}

	// Set the updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *notificationResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// // other alternative
	// resp.Diagnostics.Append(resp.State.SetAttribute(
	// 	ctx, path.Root("id"), req.ID,
	// )...)
}

func (r *notificationResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}

func ConvertStateToNotification(model NotificationResourceModel) dbt_cloud.Notification {
	notification := dbt_cloud.Notification{
		UserId:           int(model.UserID.ValueInt64()),
		OnCancel:         helper.Int64SetToIntSlice(model.OnCancel),
		OnFailure:        helper.Int64SetToIntSlice(model.OnFailure),
		OnSuccess:        helper.Int64SetToIntSlice(model.OnSuccess),
		State:            int(model.State.ValueInt64()),
		NotificationType: int(model.NotificationType.ValueInt64()),
	}

	if !model.ID.IsNull() {
		idStr := model.ID.ValueString()
		id, err := strconv.Atoi(idStr)
		if err == nil {
			notification.Id = &id
		}
	}

	if !model.ExternalEmail.IsNull() {
		externalEmail := model.ExternalEmail.ValueString()
		notification.ExternalEmail = &externalEmail
	}

	if !model.SlackChannelID.IsNull() {
		slackChannelID := model.SlackChannelID.ValueString()
		notification.SlackChannelID = &slackChannelID
	}

	if !model.SlackChannelName.IsNull() {
		slackChannelName := model.SlackChannelName.ValueString()
		notification.SlackChannelName = &slackChannelName
	}

	return notification
}
