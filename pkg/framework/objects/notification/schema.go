package notification

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *notificationResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Setup notifications on jobs success/failure to internal users, external email addresses or Slack channels",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the notification",
				// this is used so that we don't show that ID is going to change
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.Int64Attribute{
				Required:    true,
				Description: "Internal dbt Cloud User ID. Must be the user_id for an existing user even if the notification is an external one",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"on_cancel": schema.SetAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
				Computed:    true,
				Default:     helper.EmptySetDefault(types.Int64Type),
				Description: "List of job IDs to trigger the webhook on cancel",
			},
			"on_failure": schema.SetAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
				Computed:    true,
				Default:     helper.EmptySetDefault(types.Int64Type),
				Description: "List of job IDs to trigger the webhook on failure",
			},
			"on_warning": schema.SetAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
				Computed:    true,
				Default:     helper.EmptySetDefault(types.Int64Type),
				Description: "List of job IDs to trigger the webhook on warning",
			},
			"on_success": schema.SetAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
				Computed:    true,
				Default:     helper.EmptySetDefault(types.Int64Type),
				Description: "List of job IDs to trigger the webhook on success",
			},
			"state": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Description: "State of the notification (1 = active (default), 2 = inactive)",
			},
			"notification_type": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Description: "Type of notification (1 = dbt Cloud user email (default): does not require an external_email ; 2 = Slack channel: requires `slack_channel_id` and `slack_channel_name` ; 4 = external email: requires setting an `external_email`)",
			},
			"external_email": schema.StringAttribute{
				Optional:    true,
				Description: "The external email to receive the notification",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("slack_channel_id"),
						path.MatchRoot("slack_channel_name"),
					),
				},
			},
			"slack_channel_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the Slack channel to receive the notification. It can be found at the bottom of the Slack channel settings",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("external_email")),
				},
			},
			"slack_channel_name": schema.StringAttribute{
				Optional:    true,
				Description: "The name of the slack channel",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("external_email")),
				},
			},
		},
	}
}
