package data_sources

import (
	"context"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var webhookSchema = map[string]*schema.Schema{
	"webhook_id": &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "Webhooks ID",
	},
	"name": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Webhooks Name",
	},
	"description": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Webhooks Description",
	},
	"client_url": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Webhooks Client URL",
	},
	"event_types": &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Webhooks Event Types",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"job_ids": &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "List of job IDs to trigger the webhook",
		Elem: &schema.Schema{
			Type: schema.TypeInt,
		},
	},
	"active": &schema.Schema{
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Webhooks active flag",
	},
	"http_status_code": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Webhooks HTTP Status Code",
	},
	"account_identifier": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Webhooks Account Identifier",
	},
}

func DatasourceWebhook() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceWebhookRead,
		Schema:      webhookSchema,
	}
}

func datasourceWebhookRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	webhookId := d.Get("webhook_id").(string)

	webhook, err := c.GetWebhook(webhookId)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("webhook_id", webhook.WebhookId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", webhook.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", webhook.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("client_url", webhook.ClientUrl); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("event_types", webhook.EventTypes); err != nil {
		return diag.FromErr(err)
	}
	jobIdsInt := make([]int, len(webhook.JobIds))
	for i, step := range webhook.JobIds {
		jobIdsInt[i], _ = strconv.Atoi(step)
	}
	if err := d.Set("job_ids", jobIdsInt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("active", webhook.Active); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("http_status_code", webhook.HttpStatusCode); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("account_identifier", webhook.AccountIdentifier); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(webhookId)

	return diags
}
