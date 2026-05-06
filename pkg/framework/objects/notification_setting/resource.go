package notification_setting

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = &notificationSettingResource{}
	_ resource.ResourceWithConfigure      = &notificationSettingResource{}
	_ resource.ResourceWithImportState    = &notificationSettingResource{}
	_ resource.ResourceWithValidateConfig = &notificationSettingResource{}
)

func NotificationSettingResource() resource.Resource {
	return &notificationSettingResource{}
}

type notificationSettingResource struct {
	client *dbt_cloud.Client
}

func (r *notificationSettingResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_setting"
}

func (r *notificationSettingResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*dbt_cloud.Client)
}

func (r *notificationSettingResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var data NotificationSettingResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamsValidTriggers := map[string]bool{
		"run_warning":    true,
		"run_successful": true,
		"run_errored":    true,
		"run_cancelled":  true,
	}
	webhookValidTriggers := map[string]bool{
		"run_started":       true,
		"run_errored":       true,
		"metadata_ingested": true,
	}

	for i, rule := range data.Rules {
		if rule.TriggerOn.IsNull() || rule.TriggerOn.IsUnknown() {
			continue
		}
		trigger := rule.TriggerOn.ValueString()
		rulePath := path.Root("rules").AtListIndex(i)

		for _, ch := range data.Channels {
			if ch.ChannelType.IsNull() || ch.ChannelType.IsUnknown() {
				continue
			}
			channelType := ch.ChannelType.ValueString()

			switch channelType {
			case "teams":
				if !teamsValidTriggers[trigger] {
					resp.Diagnostics.AddAttributeError(
						rulePath,
						"Invalid trigger for Teams channel",
						fmt.Sprintf("`trigger_on = %q` is not supported for Teams channels. Teams supports: run_warning, run_successful, run_errored, run_cancelled.", trigger),
					)
				}
			case "webhook":
				if !webhookValidTriggers[trigger] {
					resp.Diagnostics.AddAttributeError(
						rulePath,
						"Invalid trigger for webhook channel",
						fmt.Sprintf("`trigger_on = %q` is not supported for webhook channels. Webhooks support: run_started, run_errored, metadata_ingested.", trigger),
					)
				}
			}
		}
	}
}

func (r *notificationSettingResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state NotificationSettingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	setting, err := r.client.GetNotificationSetting(state.ID.ValueInt64())
	if err != nil {
		if helper.HandleResourceNotFound(ctx, err, &resp.Diagnostics, &resp.State, "notification setting") {
			return
		}
		resp.Diagnostics.AddError("Error reading notification setting", err.Error())
		return
	}

	// webhook_hmac_secret and webhook_subscription_id are write-only:
	// the API does not echo them on read. Preserve whatever is already in state.
	priorSecrets := make(map[int64]types.String, len(state.Channels))
	priorSubscriptionIDs := make(map[int64]types.String, len(state.Channels))
	for _, ch := range state.Channels {
		if !ch.ID.IsNull() && !ch.ID.IsUnknown() {
			priorSecrets[ch.ID.ValueInt64()] = ch.WebhookHmacSecret
			priorSubscriptionIDs[ch.ID.ValueInt64()] = ch.WebhookSubscriptionID
		}
	}

	apiToModel(setting, &state, priorSecrets, priorSubscriptionIDs)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *notificationSettingResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan NotificationSettingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateNotificationSetting(modelToAPI(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error creating notification setting", err.Error())
		return
	}

	// Carry write-only secrets back into state from the plan, keyed by position.
	apiToModelByIndex(created, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *notificationSettingResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan NotificationSettingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateNotificationSetting(plan.ID.ValueInt64(), modelToAPI(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating notification setting", err.Error())
		return
	}

	apiToModelByIndex(updated, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *notificationSettingResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state NotificationSettingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteNotificationSetting(state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error deleting notification setting", err.Error())
		return
	}
}

func (r *notificationSettingResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Notification setting ID must be an integer, got %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

// normalizeOptionalString maps API responses to the model. The notifications
// API coerces a missing `description` to "" on the wire, but Terraform users
// leave the attribute null when unset — without this normalization Terraform
// reports "Provider produced inconsistent result after apply" on every create
// where the user omitted `description`.
func normalizeOptionalString(s *string) types.String {
	if s == nil || *s == "" {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

// modelToAPI converts the Terraform plan into the dbt Cloud API payload.
func modelToAPI(m NotificationSettingResourceModel) dbt_cloud.NotificationSetting {
	channels := make([]dbt_cloud.NotificationSettingChannel, len(m.Channels))
	for i, ch := range m.Channels {
		channels[i] = dbt_cloud.NotificationSettingChannel{
			ChannelType:           ch.ChannelType.ValueString(),
			WebhookSubscriptionID: ch.WebhookSubscriptionID.ValueStringPointer(),
			WebhookClientURL:      ch.WebhookClientURL.ValueStringPointer(),
			WebhookHmacSecret:     ch.WebhookHmacSecret.ValueStringPointer(),
			TeamsTeamID:           ch.TeamsTeamID.ValueStringPointer(),
			TeamsChannelID:        ch.TeamsChannelID.ValueStringPointer(),
		}
	}

	rules := make([]dbt_cloud.NotificationSettingRule, len(m.Rules))
	for i, rule := range m.Rules {
		var jobID *int64
		if !rule.JobID.IsNull() && !rule.JobID.IsUnknown() {
			v := rule.JobID.ValueInt64()
			jobID = &v
		}
		rules[i] = dbt_cloud.NotificationSettingRule{
			TriggerOn: rule.TriggerOn.ValueString(),
			JobID:     jobID,
		}
	}

	return dbt_cloud.NotificationSetting{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueStringPointer(),
		IsActive:    m.IsActive.ValueBool(),
		Channels:    channels,
		Rules:       rules,
	}
}

// apiToModel populates the model from the API response. Write-only fields keyed
// by channel ID from the prior state are preserved (the API never returns them).
func apiToModel(
	s *dbt_cloud.NotificationSetting,
	m *NotificationSettingResourceModel,
	priorSecrets map[int64]types.String,
	priorSubscriptionIDs map[int64]types.String,
) {
	if s.ID != nil {
		m.ID = types.Int64Value(*s.ID)
	}
	m.Name = types.StringValue(s.Name)
	m.Description = normalizeOptionalString(s.Description)
	m.IsActive = types.BoolValue(s.IsActive)

	m.Channels = make([]NotificationSettingChannelModel, len(s.Channels))
	for i, ch := range s.Channels {
		channel := NotificationSettingChannelModel{
			ChannelType:           types.StringValue(ch.ChannelType),
			WebhookClientURL:      types.StringPointerValue(ch.WebhookClientURL),
			TeamsTeamID:           types.StringPointerValue(ch.TeamsTeamID),
			TeamsChannelID:        types.StringPointerValue(ch.TeamsChannelID),
			WebhookHmacSecret:     types.StringNull(),
			WebhookSubscriptionID: types.StringNull(),
		}
		if ch.ID != nil {
			channel.ID = types.Int64Value(*ch.ID)
			if prior, ok := priorSecrets[*ch.ID]; ok {
				channel.WebhookHmacSecret = prior
			}
			if prior, ok := priorSubscriptionIDs[*ch.ID]; ok {
				channel.WebhookSubscriptionID = prior
			}
		}
		m.Channels[i] = channel
	}

	m.Rules = make([]NotificationSettingRuleModel, len(s.Rules))
	for i, rule := range s.Rules {
		ruleModel := NotificationSettingRuleModel{
			TriggerOn: types.StringValue(rule.TriggerOn),
			JobID:     types.Int64PointerValue(rule.JobID),
			JobName:   types.StringPointerValue(rule.JobName),
		}
		if rule.ID != nil {
			ruleModel.ID = types.Int64Value(*rule.ID)
		}
		m.Rules[i] = ruleModel
	}
}

// apiToModelByIndex updates the model from the API response after a Create/Update.
// Channels and rules are matched by position so write-only secrets in the plan
// (which the API doesn't echo back) are preserved.
func apiToModelByIndex(s *dbt_cloud.NotificationSetting, m *NotificationSettingResourceModel) {
	if s.ID != nil {
		m.ID = types.Int64Value(*s.ID)
	}
	m.Name = types.StringValue(s.Name)
	m.Description = normalizeOptionalString(s.Description)
	m.IsActive = types.BoolValue(s.IsActive)

	for i := range s.Channels {
		ch := s.Channels[i]
		if i >= len(m.Channels) {
			break
		}
		if ch.ID != nil {
			m.Channels[i].ID = types.Int64Value(*ch.ID)
		}
		m.Channels[i].ChannelType = types.StringValue(ch.ChannelType)
		// webhook_client_url, teams_team_id, teams_channel_id come back from the API.
		// Re-set them so unset Optional fields become null rather than unknown.
		m.Channels[i].WebhookClientURL = types.StringPointerValue(ch.WebhookClientURL)
		m.Channels[i].TeamsTeamID = types.StringPointerValue(ch.TeamsTeamID)
		m.Channels[i].TeamsChannelID = types.StringPointerValue(ch.TeamsChannelID)
		// webhook_hmac_secret and webhook_subscription_id stay as supplied by the plan
		// (the API doesn't echo them).
	}

	for i := range s.Rules {
		rule := s.Rules[i]
		if i >= len(m.Rules) {
			break
		}
		if rule.ID != nil {
			m.Rules[i].ID = types.Int64Value(*rule.ID)
		}
		m.Rules[i].TriggerOn = types.StringValue(rule.TriggerOn)
		m.Rules[i].JobID = types.Int64PointerValue(rule.JobID)
		m.Rules[i].JobName = types.StringPointerValue(rule.JobName)
	}
}
