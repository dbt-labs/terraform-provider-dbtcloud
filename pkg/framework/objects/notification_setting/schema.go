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
			`Configures Microsoft Teams notifications using dbt Cloud's notifications system.

			This resource manages Teams notifications through a redesigned notifications system that models a single setting as a collection of channels (where to send) and rules (when to send).`,
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
							Description: "Channel type. Currently only `teams` is supported.",
							Validators: []validator.String{
								stringvalidator.OneOf("teams"),
							},
						},
						"teams_team_id": schema.StringAttribute{
							Required:    true,
							Description: "Microsoft Teams team ID.",
						},
						"teams_channel_id": schema.StringAttribute{
							Required:    true,
							Description: "Microsoft Teams channel ID.",
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
							Description: "Event that fires the notification. Valid values: `run_warning`, `run_successful`, `run_errored`, `run_cancelled`.",
							Validators: []validator.String{
								stringvalidator.OneOf(
									"run_warning",
									"run_successful",
									"run_errored",
									"run_cancelled",
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
