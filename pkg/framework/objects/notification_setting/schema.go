package notification_setting

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *notificationSettingResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: helper.DocString(
			`Configures Microsoft Teams and webhook notifications using dbt Cloud's notifications system.

			This is a separate resource from ~~~dbtcloud_notification~~~ because Microsoft Teams notifications are delivered through a redesigned notifications system that models a single setting as a collection of channels (where to send) and rules (when to send). Email and Slack notifications still go through the legacy ~~~dbtcloud_notification~~~ resource.`,
		),
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "ID of the notification setting",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable name for this notification setting",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(1024),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Optional description of what this notification setting does",
			},
			"is_active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether this notification setting is active. Defaults to true.",
			},
			"channels": schema.ListNestedAttribute{
				Required:    true,
				Description: "Delivery channels for this setting. At least one channel is required.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:    true,
							Description: "Channel ID",
						},
						"channel_type": schema.StringAttribute{
							Required:    true,
							Description: "Channel type. One of `teams` or `webhook`.",
							Validators: []validator.String{
								stringvalidator.OneOf("teams", "webhook"),
							},
						},
						"teams_team_id": schema.StringAttribute{
							Optional:    true,
							Description: "Microsoft Teams team ID. Required when `channel_type` is `teams`.",
						},
						"teams_channel_id": schema.StringAttribute{
							Optional:    true,
							Description: "Microsoft Teams channel ID. Required when `channel_type` is `teams`.",
						},
						"webhook_client_url": schema.StringAttribute{
							Optional:    true,
							Description: "Webhook URL to POST notifications to. Required when `channel_type` is `webhook`.",
						},
						"webhook_hmac_secret": schema.StringAttribute{
							Optional:    true,
							Sensitive:   true,
							Description: "HMAC secret used to sign webhook payloads. Only used when `channel_type` is `webhook`. Write-only: the API does not return this value, so changes outside Terraform cannot be detected.",
						},
						"webhook_subscription_id": schema.StringAttribute{
							Optional:    true,
							Description: "Subscription ID for webhook channels. Optional; the API generates one when omitted. Write-only: not returned by the API.",
						},
					},
				},
			},
			"rules": schema.ListNestedAttribute{
				Required:    true,
				Description: "Trigger rules. At least one rule is required.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:    true,
							Description: "Rule ID",
						},
						"trigger_on": schema.StringAttribute{
							Required:    true,
							Description: "Event that fires the notification. One of `run_started`, `run_successful`, `run_warning`, `run_cancelled`, `run_errored`, `metadata_ingested`.",
							Validators: []validator.String{
								stringvalidator.OneOf(
									"run_started",
									"run_successful",
									"run_warning",
									"run_cancelled",
									"run_errored",
									"metadata_ingested",
								),
							},
						},
						"job_id": schema.Int64Attribute{
							Optional:    true,
							Description: "Job ID this rule applies to. Omit to fire for all jobs in the account.",
						},
						"job_name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the job referenced by `job_id` (read-only).",
						},
					},
				},
			},
		},
	}
}
