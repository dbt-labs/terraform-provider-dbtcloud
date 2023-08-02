package data_sources

import (
	"context"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var notificationSchema = map[string]*schema.Schema{
	"notification_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "ID of the notification",
	},
	"user_id": &schema.Schema{
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Internal dbt Cloud User ID. Must be the user_id for an existing user even if the notification is an external one",
	},
	"on_cancel": &schema.Schema{
		Type:     schema.TypeSet,
		Computed: true,
		Elem: &schema.Schema{
			Type: schema.TypeInt,
		},
		Description: "List of job IDs to trigger the webhook on cancel",
	},
	"on_failure": &schema.Schema{
		Type:     schema.TypeSet,
		Computed: true,
		Elem: &schema.Schema{
			Type: schema.TypeInt,
		},
		Description: "List of job IDs to trigger the webhook on failure",
	},
	"on_success": &schema.Schema{
		Type:     schema.TypeSet,
		Computed: true,
		Elem: &schema.Schema{
			Type: schema.TypeInt,
		},
		Description: "List of job IDs to trigger the webhook on success",
	},
	"notification_type": &schema.Schema{
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Type of notification (1 = dbt Cloud user email (default): does not require an external_email ; 4 = external email: requires setting an external_email)",
	},
	"external_email": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The external email to receive the notification",
	},
}

func DatasourceNotification() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceNotificationRead,
		Schema:      notificationSchema,
	}
}

func datasourceNotificationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	notificationId := strconv.Itoa(d.Get("notification_id").(int))

	notification, err := c.GetNotification(notificationId)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("user_id", notification.UserId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("on_cancel", notification.OnCancel); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("on_failure", notification.OnFailure); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("on_success", notification.OnSuccess); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("notification_type", notification.NotificationType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("external_email", notification.ExternalEmail); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(notificationId)

	return diags
}
