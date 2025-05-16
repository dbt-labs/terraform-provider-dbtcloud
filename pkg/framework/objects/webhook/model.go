package webhook

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type webhookDataSourceModel struct {
	WebhookID         types.String   `tfsdk:"webhook_id"`
	Name              types.String   `tfsdk:"name"`
	Description       types.String   `tfsdk:"description"`
	ClientURL         types.String   `tfsdk:"client_url"`
	EventTypes        []types.String `tfsdk:"event_types"`
	JobIDs            []types.Int64  `tfsdk:"job_ids"`
	Active            types.Bool     `tfsdk:"active"`
	HTTPStatusCode    types.String   `tfsdk:"http_status_code"`
	AccountIdentifier types.String   `tfsdk:"account_identifier"`
}
