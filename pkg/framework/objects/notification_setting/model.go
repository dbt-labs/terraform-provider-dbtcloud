package notification_setting

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NotificationSettingResourceModel struct {
	ID          types.Int64                          `tfsdk:"id"`
	Name        types.String                         `tfsdk:"name"`
	Description types.String                         `tfsdk:"description"`
	IsActive    types.Bool                           `tfsdk:"is_active"`
	Channels    []NotificationSettingChannelModel    `tfsdk:"channels"`
	Rules       []NotificationSettingRuleModel       `tfsdk:"rules"`
}

type NotificationSettingChannelModel struct {
	ID                    types.Int64  `tfsdk:"id"`
	ChannelType           types.String `tfsdk:"channel_type"`
	WebhookSubscriptionID types.String `tfsdk:"webhook_subscription_id"`
	WebhookClientURL      types.String `tfsdk:"webhook_client_url"`
	WebhookHmacSecret     types.String `tfsdk:"webhook_hmac_secret"`
	TeamsTeamID           types.String `tfsdk:"teams_team_id"`
	TeamsChannelID        types.String `tfsdk:"teams_channel_id"`
}

type NotificationSettingRuleModel struct {
	ID        types.Int64  `tfsdk:"id"`
	TriggerOn types.String `tfsdk:"trigger_on"`
	JobID     types.Int64  `tfsdk:"job_id"`
	JobName   types.String `tfsdk:"job_name"`
}
