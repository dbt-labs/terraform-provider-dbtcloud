package webhook

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource = &webhookDataSource{}
)

func WebhookDataSource() datasource.DataSource {
	return &webhookDataSource{}
}

type webhookDataSource struct {
	client *dbt_cloud.Client
}

func (d *webhookDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (d *webhookDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state WebhookDataSourceModel
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	webhook, err := d.client.GetWebhook(state.WebhookID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Did not find webhook with id: %s", state.WebhookID.ValueString()),
			err.Error(),
		)
		return
	}

	state.WebhookID = types.StringValue(webhook.WebhookId)
	state.Name = types.StringValue(webhook.Name)
	state.Description = types.StringValue(webhook.Description)
	state.ClientURL = types.StringValue(webhook.ClientUrl)

	state.EventTypes, diags = helper.SliceStringToTypesListStringValue(webhook.EventTypes)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	state.JobIDs, diags = helper.SliceStringToTypesListInt64Value(webhook.JobIds)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	state.Active = types.BoolValue(webhook.Active)
	state.HTTPStatusCode = helper.ConvertStringPointer(webhook.HttpStatusCode)
	state.AccountIdentifier = helper.ConvertStringPointer(webhook.AccountIdentifier)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *webhookDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*dbt_cloud.Client)
}

func (d *webhookDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = dataSourceSchema
}
