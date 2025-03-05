package model_notifications

import (
	"context"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &modelNotificationsDataSource{}
	_ datasource.DataSourceWithConfigure = &modelNotificationsDataSource{}
)

func ModelNotificationsDataSource() datasource.DataSource {
	return &modelNotificationsDataSource{}
}

type modelNotificationsDataSource struct {
	client *dbt_cloud.Client
}

func (d *modelNotificationsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_model_notifications"
}

func (d *modelNotificationsDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Get model notifications configuration for a dbt Cloud environment",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The internal ID of the model notifications configuration",
			},
			"environment_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the dbt Cloud environment",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether model notifications are enabled for this environment",
			},
			"on_success": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether to send notifications for successful model runs",
			},
			"on_failure": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether to send notifications for failed model runs",
			},
			"on_warning": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether to send notifications for model runs with warnings",
			},
			"on_skipped": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether to send notifications for skipped model runs",
			},
		},
	}
}

func (d *modelNotificationsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data ModelNotificationsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	environmentID := data.EnvironmentID.ValueString()
	modelNotifications, err := d.client.GetModelNotifications(environmentID)
	if err != nil {
		resp.Diagnostics.AddError("Error getting model notifications", err.Error())
		return
	}

	data.ID = types.Int64Value(int64(*modelNotifications.ID))
	data.EnvironmentID = types.StringValue(strconv.Itoa(modelNotifications.EnvironmentID))
	data.Enabled = types.BoolValue(modelNotifications.Enabled)
	data.OnSuccess = types.BoolValue(modelNotifications.OnSuccess)
	data.OnFailure = types.BoolValue(modelNotifications.OnFailure)
	data.OnWarning = types.BoolValue(modelNotifications.OnWarning)
	data.OnSkipped = types.BoolValue(modelNotifications.OnSkipped)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *modelNotificationsDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*dbt_cloud.Client)
}
