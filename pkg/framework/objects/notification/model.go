package notification

import (
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NotificationResourceModel struct {
	ID               types.String `tfsdk:"id"`
	UserID           types.Int64  `tfsdk:"user_id"`
	OnCancel         types.Set    `tfsdk:"on_cancel"`
	OnFailure        types.Set    `tfsdk:"on_failure"`
	OnSuccess        types.Set    `tfsdk:"on_success"`
	State            types.Int64  `tfsdk:"state"`
	NotificationType types.Int64  `tfsdk:"notification_type"`
	ExternalEmail    types.String `tfsdk:"external_email"`
	SlackChannelID   types.String `tfsdk:"slack_channel_id"`
	SlackChannelName types.String `tfsdk:"slack_channel_name"`
}

type NotificationDataSourceModel struct {
	NotificationID   types.Int64  `tfsdk:"notification_id"`
	UserID           types.Int64  `tfsdk:"user_id"`
	OnCancel         types.Set    `tfsdk:"on_cancel"`
	OnFailure        types.Set    `tfsdk:"on_failure"`
	OnSuccess        types.Set    `tfsdk:"on_success"`
	State            types.Int64  `tfsdk:"state"`
	NotificationType types.Int64  `tfsdk:"notification_type"`
	ExternalEmail    types.String `tfsdk:"external_email"`
	SlackChannelID   types.String `tfsdk:"slack_channel_id"`
	SlackChannelName types.String `tfsdk:"slack_channel_name"`
}

func ConvertNotificationModelToData(model NotificationResourceModel) dbt_cloud.Notification {
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
