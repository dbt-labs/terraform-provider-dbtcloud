package resources

import (
	"context"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var notificationSchema = map[string]*schema.Schema{

	"user_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Internal dbt Cloud User ID. Must be the user_id for an existing user even if the notification is an external one",
		ForceNew:    true,
	},
	"on_cancel": &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeInt,
		},
		Description: "List of job IDs to trigger the webhook on cancel",
	},
	"on_failure": &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeInt,
		},
		Description: "List of job IDs to trigger the webhook on failure",
	},
	"on_success": &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeInt,
		},
		Description: "List of job IDs to trigger the webhook on success",
	},
	"state": &schema.Schema{
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     1,
		Description: "State of the notification (1 = active (default), 2 = inactive)",
	},
	"notification_type": &schema.Schema{
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     1,
		Description: "Type of notification (1 = dbt Cloud user email (default): does not require an external_email ; 4 = external email: requires setting an external_email)",
	},
	"external_email": &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The external email to receive the notification",
	},
}

func ResourceNotification() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNotificationCreate,
		ReadContext:   resourceNotificationRead,
		UpdateContext: resourceNotificationUpdate,
		DeleteContext: resourceNotificationDelete,

		Schema: notificationSchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceNotificationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	notificationId := d.Id()

	notification, err := c.GetNotification(notificationId)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
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
	if err := d.Set("state", notification.State); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("notification_type", notification.NotificationType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("external_email", notification.ExternalEmail); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceNotificationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	userId := d.Get("user_id").(int)
	OnCancelRaw := d.Get("on_cancel").(*schema.Set).List()
	onFailureRaw := d.Get("on_failure").(*schema.Set).List()
	onSuccessRaw := d.Get("on_success").(*schema.Set).List()
	state := d.Get("state").(int)
	notificationType := d.Get("notification_type").(int)
	var externalEmailVal *string
	if d.Get("external_email").(string) == "" {
		externalEmailVal = nil
	} else {
		externalEmail := d.Get("external_email").(string)
		externalEmailVal = &externalEmail
	}

	// we need to loop through the values to convert them to ints
	onCancel := make([]int, len(OnCancelRaw))
	for i, jobId := range OnCancelRaw {
		onCancel[i] = jobId.(int)
	}

	onFailure := make([]int, len(onFailureRaw))
	for i, jobId := range onFailureRaw {
		onFailure[i] = jobId.(int)
	}

	onSuccess := make([]int, len(onSuccessRaw))
	for i, jobId := range onSuccessRaw {
		onSuccess[i] = jobId.(int)
	}

	notif, err := c.CreateNotification(
		userId,
		onCancel,
		onFailure,
		onSuccess,
		state,
		notificationType,
		externalEmailVal,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(*notif.Id))

	resourceNotificationRead(ctx, d, m)

	return diags
}

func resourceNotificationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)
	notificationId := d.Id()

	if d.HasChange("user_id") ||
		d.HasChange("on_cancel") ||
		d.HasChange("on_failure") ||
		d.HasChange("on_success") ||
		d.HasChange("state") ||
		d.HasChange("notification_type") ||
		d.HasChange("external_email") {

		notification, err := c.GetNotification(notificationId)

		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("user_id") {
			userId := d.Get("user_id").(int)
			notification.UserId = userId
		}
		if d.HasChange("on_cancel") {
			onCancelRaw := d.Get("on_cancel").(*schema.Set).List()
			onCancel := make([]int, len(onCancelRaw))
			for i, jobId := range onCancelRaw {
				onCancel[i] = jobId.(int)
			}
			notification.OnCancel = onCancel
		}
		if d.HasChange("on_failure") {
			onFailureRaw := d.Get("on_failure").(*schema.Set).List()
			onFailure := make([]int, len(onFailureRaw))
			for i, jobId := range onFailureRaw {
				onFailure[i] = jobId.(int)
			}
			notification.OnFailure = onFailure
		}
		if d.HasChange("on_success") {
			onSuccessRaw := d.Get("on_success").(*schema.Set).List()
			onSuccess := make([]int, len(onSuccessRaw))
			for i, jobId := range onSuccessRaw {
				onSuccess[i] = jobId.(int)
			}
			notification.OnSuccess = onSuccess
		}
		if d.HasChange("state") {
			state := d.Get("state").(int)
			notification.State = state
		}
		if d.HasChange("notification_type") {
			notificationType := d.Get("notification_type").(int)
			notification.NotificationType = notificationType
		}
		if d.HasChange("external_email") {
			externalEmail := d.Get("external_email").(string)
			if externalEmail == "" {
				notification.ExternalEmail = nil
			} else {
				notification.ExternalEmail = &externalEmail
			}
		}

		_, err = c.UpdateNotification(notificationId, *notification)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceNotificationRead(ctx, d, m)
}

func resourceNotificationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)
	notificationId := d.Id()

	var diags diag.Diagnostics

	notification, err := c.GetNotification(notificationId)
	if err != nil {
		return diag.FromErr(err)
	}

	notification.State = dbt_cloud.STATE_DELETED
	_, err = c.UpdateNotification(notificationId, *notification)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
