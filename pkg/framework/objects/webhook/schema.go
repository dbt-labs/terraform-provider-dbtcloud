package webhook

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *webhookDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieve webhook details",
		Attributes: map[string]schema.Attribute{
			"webhook_id": schema.StringAttribute{
				Required:    true,
				Description: "Webhooks ID",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Webhooks Name",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Webhooks Description",
			},
			"client_url": schema.StringAttribute{
				Computed:    true,
				Description: "Webhooks Client URL",
			},
			"event_types": schema.ListAttribute{
				Computed:    true,
				Description: "Webhooks Event Types",
				ElementType: types.StringType,
			},
			"job_ids": schema.ListAttribute{
				Computed:    true,
				Description: "List of job IDs to trigger the webhook",
				ElementType: types.Int64Type,
			},
			"active": schema.BoolAttribute{
				Computed:    true,
				Description: "Webhooks active flag",
			},
			"http_status_code": schema.StringAttribute{
				Computed:    true,
				Description: "Webhooks HTTP Status Code",
			},
			"account_identifier": schema.StringAttribute{
				Computed:    true,
				Description: "Webhooks Account Identifier",
			},
		},
	}
}
