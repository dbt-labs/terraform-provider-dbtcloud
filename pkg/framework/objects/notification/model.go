package notification

import "github.com/hashicorp/terraform-plugin-framework/types"

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
