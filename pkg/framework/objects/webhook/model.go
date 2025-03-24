package webhook

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WebhookDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	WebhookID         types.String `tfsdk:"webhook_id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	ClientURL         types.String `tfsdk:"client_url"`
	EventTypes        types.List   `tfsdk:"event_types"`
	JobIDs            types.List   `tfsdk:"job_ids"`
	Active            types.Bool   `tfsdk:"active"`
	HTTPStatusCode    types.String `tfsdk:"http_status_code"`
	AccountIdentifier types.String `tfsdk:"account_identifier"`
}

type WebhookResourceModel struct {
	ID                types.String `tfsdk:"id"`
	WebhookID         types.String `tfsdk:"webhook_id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	ClientURL         types.String `tfsdk:"client_url"`
	EventTypes        types.List   `tfsdk:"event_types"`
	JobIDs            types.List   `tfsdk:"job_ids"`
	Active            types.Bool   `tfsdk:"active"`
	HmacSecret        types.String `tfsdk:"hmac_secret"`
	HTTPStatusCode    types.String `tfsdk:"http_status_code"`
	AccountIdentifier types.String `tfsdk:"account_identifier"`
}
