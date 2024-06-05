package partial_notification

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

func (r *partialNotificationResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: helper.DocString(
			`Setup partial notifications on jobs success/failure to internal users, external email addresses or Slack channels. This is different from ~~~dbt_cloud_notification~~~ as it allows to have multiple resources updating the same notification recipient (email, user or Slack channel) and is useful for companies managing a single dbt Cloud Account configuration from different Terraform projects/workspaces.

			If a company uses only one Terraform project/workspace to manage all their dbt Cloud Account config, it is recommended to use ~~~dbt_cloud_notification~~~ instead of ~~~dbt_cloud_partial_notification~~~.

			~> This is a new resource. Feedback is welcome.

			The resource currently requires a Service Token with Account Admin access.

			The current behavior of the resource is the following:

			- when using ~~~dbt_cloud_partial_notification~~~, don't use ~~~dbt_cloud_notification~~~ for the same notification recipient in any other project/workspace. Otherwise, the behavior is undefined and partial notifications might be removed.
			- when defining a new ~~~dbt_cloud_partial_notification~~~
			  - if the notification recipient doesn't exist, it will be created
			  - if a notification config exists for the current recipient, Job IDs will be added in the list of jobs to trigger the notifications
			- in a given Terraform project/workspace, avoid having different ~~~dbt_cloud_partial_notification~~~ for the same recipient to prevent sync issues. Add all the jobs in the same resource. 
			- all resources for the same notification recipient need to have the same values for ~~~state~~~ and ~~~user_id~~~. Those fields are not considered "partial".
			- when a resource is updated, the dbt Cloud notification recipient will be updated accordingly, removing and adding job ids in the list of jobs triggering notifications
			- when the resource is deleted/destroyed, if the resulting notification recipient list of jobs is empty, the notification will be deleted ; otherwise, the notification will be updated, removing the job ids from the deleted resource
			`,
		),
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
				Description: "Internal dbt Cloud User ID. Must be the user_id for an existing user even if the notification is an external one [global]",
			},
			"on_cancel": schema.SetAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
				Computed:    true,
				Default:     helper.EmptySetDefault(types.Int64Type),
				Description: "List of job IDs to trigger the webhook on cancel. Those will be added/removed when config is added/removed.",
			},
			"on_failure": schema.SetAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
				Computed:    true,
				Default:     helper.EmptySetDefault(types.Int64Type),
				Description: "List of job IDs to trigger the webhook on failure Those will be added/removed when config is added/removed.",
			},
			"on_warning": schema.SetAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
				Computed:    true,
				Default:     helper.EmptySetDefault(types.Int64Type),
				Description: "List of job IDs to trigger the webhook on warning Those will be added/removed when config is added/removed.",
			},
			"on_success": schema.SetAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
				Computed:    true,
				Default:     helper.EmptySetDefault(types.Int64Type),
				Description: "List of job IDs to trigger the webhook on success Those will be added/removed when config is added/removed.",
			},
			"state": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Description: "State of the notification (1 = active (default), 2 = inactive) [global]",
			},
			"notification_type": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Description: "Type of notification (1 = dbt Cloud user email (default): does not require an external_email ; 2 = Slack channel: requires `slack_channel_id` and `slack_channel_name` ; 4 = external email: requires setting an `external_email`) [global, used as identifier]",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"external_email": schema.StringAttribute{
				Optional:    true,
				Description: "The external email to receive the notification [global, used as identifier]",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("slack_channel_id"),
						path.MatchRoot("slack_channel_name"),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"slack_channel_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the Slack channel to receive the notification. It can be found at the bottom of the Slack channel settings [global, used as identifier]",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("external_email")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"slack_channel_name": schema.StringAttribute{
				Optional:    true,
				Description: "The name of the slack channel [global, used as identifier]",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("external_email")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}
