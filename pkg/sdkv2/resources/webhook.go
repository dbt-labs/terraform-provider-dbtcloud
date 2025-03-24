package resources

import (
	"context"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	eventTypes = []string{
		"job.run.completed",
		"job.run.started",
		"job.run.errored",
	}
)

var webhookSchema = map[string]*schema.Schema{
	"webhook_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Webhooks ID",
	},
	"name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Webhooks Name",
	},
	"description": {
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "",
		Description: "Webhooks Description",
	},
	"client_url": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Webhooks Client URL",
	},
	"event_types": {
		Type:        schema.TypeList,
		Required:    true,
		Description: "Webhooks Event Types",
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice(eventTypes, false),
		},
	},
	"job_ids": {
		Type:        schema.TypeList,
		Optional:    true,
		Description: "List of job IDs to trigger the webhook, An empty list will trigger on all jobs",
		Elem: &schema.Schema{
			Type: schema.TypeInt,
		},
	},
	"active": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "Webhooks active flag",
	},
	"hmac_secret": {
		Type:        schema.TypeString,
		Computed:    true,
		Sensitive:   true,
		Description: "Secret key for the webhook. Can be used to validate the authenticity of the webhook.",
	},
	"http_status_code": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Latest HTTP status of the webhook",
	},
	"account_identifier": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Webhooks Account Identifier",
	},
}

func ResourceWebhook() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWebhookCreate,
		ReadContext:   resourceWebhookRead,
		UpdateContext: resourceWebhookUpdate,
		DeleteContext: resourceWebhookDelete,

		Schema: webhookSchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceWebhookRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	webhookId := d.Id()

	webhook, err := c.GetWebhook(webhookId)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
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

	// we need to convert from string to int as we get string from the API but store as int
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

	return diags
}

func resourceWebhookCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	webhookId := ""
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	clientUrl := d.Get("client_url").(string)
	eventTypes := d.Get("event_types").([]interface{})
	jobIds := d.Get("job_ids").([]interface{})
	active := d.Get("active").(bool)

	typedEventTypes := []string{}
	for _, eventType := range eventTypes {
		typedEventTypes = append(typedEventTypes, eventType.(string))

	}
	typedJobIds := []int{}
	for _, jobId := range jobIds {
		typedJobIds = append(typedJobIds, jobId.(int))
	}

	w, err := c.CreateWebhook(
		webhookId,
		name,
		description,
		clientUrl,
		typedEventTypes,
		typedJobIds,
		active,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(w.WebhookId)

	// we have to set it at the create level
	if err := d.Set("hmac_secret", w.HmacSecret); err != nil {
		return diag.FromErr(err)
	}

	resourceWebhookRead(ctx, d, m)

	return diags
}

func resourceWebhookUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)
	webhookId := d.Id()

	if d.HasChange("name") ||
		d.HasChange("description") ||
		d.HasChange("client_url") ||
		d.HasChange("event_types") ||
		d.HasChange("job_ids") {
		webhookRead, err := c.GetWebhook(webhookId)

		jobIdsWrite := make([]int, len(webhookRead.JobIds))
		for i, step := range webhookRead.JobIds {
			jobIdsWrite[i], _ = strconv.Atoi(step)
		}

		webhook := dbt_cloud.WebhookWrite{
			WebhookId:   webhookId,
			Name:        webhookRead.Name,
			Description: webhookRead.Description,
			ClientUrl:   webhookRead.ClientUrl,
			EventTypes:  webhookRead.EventTypes,
			JobIds:      jobIdsWrite,
			Active:      webhookRead.Active,
		}

		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("name") {
			name := d.Get("name").(string)
			webhook.Name = name
		}
		if d.HasChange("description") {
			description := d.Get("description").(string)
			webhook.Description = description
		}
		if d.HasChange("client_url") {
			clientUrl := d.Get("client_url").(string)
			webhook.ClientUrl = clientUrl
		}
		if d.HasChange("event_types") {
			eventTypes := make([]string, len(d.Get("event_types").([]interface{})))
			for i, step := range d.Get("event_types").([]interface{}) {
				eventTypes[i] = step.(string)
			}
			webhook.EventTypes = eventTypes
		}
		if d.HasChange("job_ids") {
			jobIDs := make([]int, len(d.Get("job_ids").([]interface{})))
			for i, step := range d.Get("job_ids").([]interface{}) {
				jobIDs[i] = step.(int)
			}
			webhook.JobIds = jobIDs
		}

		_, err = c.UpdateWebhook(webhookId, webhook)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceWebhookRead(ctx, d, m)
}

func resourceWebhookDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)
	webhookId := d.Id()

	var diags diag.Diagnostics

	_, err := c.DeleteWebhook(webhookId)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
