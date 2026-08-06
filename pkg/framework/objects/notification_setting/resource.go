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
	_ resource.Resource              = &notificationSettingResource{}
	_ resource.ResourceWithConfigure = &notificationSettingResource{}
	_ resource.ResourceWithImportState = &notificationSettingResource{}
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

	apiToModel(setting, &state)

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
		// If the resource is already deleted (404), treat as success
		if helper.HandleResourceNotFound(ctx, err, &resp.Diagnostics, &resp.State, "notification setting") {
			return
		}
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
			ChannelType:    ch.ChannelType.ValueString(),
			TeamsTeamID:    ch.TeamsTeamID.ValueStringPointer(),
			TeamsChannelID: ch.TeamsChannelID.ValueStringPointer(),
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
		IsActive:    true,
		Channels:    channels,
		Rules:       rules,
	}
}

// apiToModel populates the model from the API response.
func apiToModel(
	s *dbt_cloud.NotificationSetting,
	m *NotificationSettingResourceModel,
) {
	if s.ID != nil {
		m.ID = types.Int64Value(*s.ID)
	}
	m.Name = types.StringValue(s.Name)
	m.Description = normalizeOptionalString(s.Description)

	m.Channels = make([]NotificationSettingChannelModel, len(s.Channels))
	for i, ch := range s.Channels {
		channel := NotificationSettingChannelModel{
			ChannelType:    types.StringValue(ch.ChannelType),
			TeamsTeamID:    types.StringPointerValue(ch.TeamsTeamID),
			TeamsChannelID: types.StringPointerValue(ch.TeamsChannelID),
		}
		if ch.ID != nil {
			channel.ID = types.Int64Value(*ch.ID)
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
// Channels and rules are matched by position.
func apiToModelByIndex(s *dbt_cloud.NotificationSetting, m *NotificationSettingResourceModel) {
	if s.ID != nil {
		m.ID = types.Int64Value(*s.ID)
	}
	m.Name = types.StringValue(s.Name)
	m.Description = normalizeOptionalString(s.Description)

	for i := range s.Channels {
		ch := s.Channels[i]
		if i >= len(m.Channels) {
			break
		}
		if ch.ID != nil {
			m.Channels[i].ID = types.Int64Value(*ch.ID)
		}
		m.Channels[i].ChannelType = types.StringValue(ch.ChannelType)
		m.Channels[i].TeamsTeamID = types.StringPointerValue(ch.TeamsTeamID)
		m.Channels[i].TeamsChannelID = types.StringPointerValue(ch.TeamsChannelID)
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
