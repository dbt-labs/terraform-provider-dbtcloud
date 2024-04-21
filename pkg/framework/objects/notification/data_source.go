package notification

import (
	"context"
	"fmt"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &notificationDataSource{}
	_ datasource.DataSourceWithConfigure = &notificationDataSource{}
)

func NotificationDataSource() datasource.DataSource {
	return &notificationDataSource{}
}

type notificationDataSource struct {
	client *dbt_cloud.Client
}

func (d *notificationDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification"
}

func (d *notificationDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieve notification details",
		Attributes: map[string]schema.Attribute{
			"notification_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the notification",
			},
			"user_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Internal dbt Cloud User ID. Must be the user_id for an existing user even if the notification is an external one",
			},
			"on_cancel": schema.SetAttribute{
				ElementType: types.Int64Type,
				Computed:    true,
				Description: "List of job IDs to trigger the webhook on cancel",
			},
			"on_failure": schema.SetAttribute{
				ElementType: types.Int64Type,
				Computed:    true,
				Description: "List of job IDs to trigger the webhook on failure",
			},
			"on_success": schema.SetAttribute{
				ElementType: types.Int64Type,
				Computed:    true,
				Description: "List of job IDs to trigger the webhook on success",
			},
			"state": schema.Int64Attribute{
				Computed:    true,
				Description: "State of the notification (1 = active (default), 2 = inactive)",
			},
			"notification_type": schema.Int64Attribute{
				Computed:    true,
				Description: "Type of notification (1 = dbt Cloud user email (default): does not require an external_email ; 2 = Slack channel: requires `slack_channel_id` and `slack_channel_name` ; 4 = external email: requires setting an `external_email`)",
			},
			"external_email": schema.StringAttribute{
				Computed:    true,
				Description: "The external email to receive the notification",
			},
			"slack_channel_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Slack channel to receive the notification. It can be found at the bottom of the Slack channel settings",
			},
			"slack_channel_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the slack channel",
			},
		},
	}
}

func (d *notificationDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {

	var data NotificationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	notificationID := data.NotificationID.ValueInt64()
	notification, err := d.client.GetNotification(fmt.Sprintf("%d", notificationID))
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

	data.NotificationID = types.Int64Value(int64(notificationID))
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
	data.NotificationType = types.Int64Value(int64(notification.NotificationType))
	data.ExternalEmail = types.StringPointerValue(notification.ExternalEmail)
	data.SlackChannelID = types.StringPointerValue(notification.SlackChannelID)
	data.SlackChannelName = types.StringPointerValue(notification.SlackChannelName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *notificationDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*dbt_cloud.Client)
}
