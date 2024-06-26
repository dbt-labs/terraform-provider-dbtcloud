package partial_notification

import (
	"context"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/notification"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

var (
	_ resource.Resource              = &partialNotificationResource{}
	_ resource.ResourceWithConfigure = &partialNotificationResource{}
)

func PartialNotificationResource() resource.Resource {
	return &partialNotificationResource{}
}

type partialNotificationResource struct {
	client *dbt_cloud.Client
}

func (r *partialNotificationResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_partial_notification"
}

func (r *partialNotificationResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var data notification.NotificationResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.NotificationType == types.Int64Value(1) &&
		!(data.ExternalEmail.IsNull() && data.SlackChannelID.IsNull() && data.SlackChannelName.IsNull()) {
		resp.Diagnostics.AddAttributeError(
			path.Root("notification_type"),
			"Notification type is not compatible with the other attributes",
			"Notification type 1 is for internal notifications only. Please remove the external email, slack channel ID, and slack channel name attributes.",
		)
	}

	if data.NotificationType == types.Int64Value(2) &&
		data.SlackChannelID.IsNull() &&
		data.SlackChannelName.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("notification_type"),
			"Notification type is not compatible with the other attributes",
			"Notification type 2 requires a Slack channel ID and Slack channel name.",
		)
	}

	if data.NotificationType == types.Int64Value(4) && data.ExternalEmail.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("notification_type"),
			"Notification type is not compatible with the other attributes",
			"Notification type 4 requires an external email.",
		)
	}
}

func (r *partialNotificationResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state notification.NotificationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	// check if the ID exists
	notificationID := state.ID.ValueString()
	notification, err := r.client.GetNotification(notificationID)
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

	// if the ID exists, make sure that it is the one we are looking for
	if !matchPartial(state, *notification) {
		// read all the notifications and check if one exists
		allNotifications, err := r.client.GetAllNotifications()
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to get all notifications",
				"Error: "+err.Error(),
			)
			return
		}

		var fullNotification *dbt_cloud.Notification
		for _, notification := range allNotifications {
			if matchPartial(state, notification) {
				// it exists, we stop here
				fullNotification = &notification
				break
			}
		}

		// if it was not found, it means that the notification was deleted
		if fullNotification == nil {
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The notification resource was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}

		// if it is found, we set it correctly
		notificationID = strconv.Itoa(*fullNotification.Id)
		notification = fullNotification
	}

	// we set the "global" values
	state.ID = types.StringValue(notificationID)
	state.UserID = types.Int64Value(int64(notification.UserId))
	state.State = types.Int64Value(int64(notification.State))
	state.NotificationType = types.Int64Value(int64(notification.NotificationType))
	state.ExternalEmail = types.StringPointerValue(notification.ExternalEmail)
	state.SlackChannelID = types.StringPointerValue(notification.SlackChannelID)
	state.SlackChannelName = types.StringPointerValue(notification.SlackChannelName)

	// we set the "partial" values by intersecting the config with the remote
	intOnCancel, intOnFailure, intOnWarning, intOnSuccess, ok := extractModelJobLists(state)
	if !ok {
		resp.Diagnostics.AddError("Error extracting the model job lists", "")
		return
	}

	state.OnCancel, _ = types.SetValueFrom(
		context.Background(),
		types.Int64Type,
		lo.Intersect(intOnCancel, notification.OnCancel),
	)

	state.OnFailure, _ = types.SetValueFrom(
		context.Background(),
		types.Int64Type,
		lo.Intersect(intOnFailure, notification.OnFailure),
	)

	state.OnWarning, _ = types.SetValueFrom(
		context.Background(),
		types.Int64Type,
		lo.Intersect(intOnWarning, notification.OnWarning),
	)

	state.OnSuccess, _ = types.SetValueFrom(
		context.Background(),
		types.Int64Type,
		lo.Intersect(intOnSuccess, notification.OnSuccess),
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (r *partialNotificationResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan notification.NotificationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// we read the values from the config
	intOnCancel, intOnFailure, intOnWarning, intOnSuccess, ok := extractModelJobLists(plan)
	if !ok {
		resp.Diagnostics.AddError("Error extracting the model job lists", "")
		return
	}

	// check if it exists
	// we don't need to check uniqueness and can just return the first as the API only allows one notification per user
	allNotifications, err := r.client.GetAllNotifications()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get all notifications",
			"Error: "+err.Error(),
		)
		return
	}

	var fullNotification *dbt_cloud.Notification
	for _, notification := range allNotifications {
		if matchPartial(plan, notification) {
			// it exists, we stop here
			fullNotification = &notification
			break
		}
	}

	if fullNotification != nil {
		// if it exists, we get the ID
		notificationID := strconv.Itoa(*fullNotification.Id)
		plan.ID = types.StringValue(notificationID)

		// and we calculate all the partial fields
		// the global ones are already set in the plan
		configOnCancel := intOnCancel
		remoteOnCancel := fullNotification.OnCancel
		missingOnCancel, _ := lo.Difference(configOnCancel, remoteOnCancel)

		configOnFailure := intOnFailure
		remoteOnFailure := fullNotification.OnFailure
		missingOnFailure, _ := lo.Difference(configOnFailure, remoteOnFailure)

		configOnWarning := intOnWarning
		remoteOnWarning := fullNotification.OnWarning
		missingOnWarning := lo.Without(configOnWarning, remoteOnWarning...)

		configOnSuccess := intOnSuccess
		remoteOnSuccess := fullNotification.OnSuccess
		missingOnSuccess, _ := lo.Difference(configOnSuccess, remoteOnSuccess)

		// we only update if something global, but not part of the ID is different or if something partial needs to be added
		if plan.State == types.Int64Value(int64(fullNotification.State)) &&
			plan.UserID == types.Int64Value(int64(fullNotification.UserId)) &&
			len(missingOnCancel) == 0 &&
			len(missingOnFailure) == 0 &&
			len(missingOnWarning) == 0 &&
			len(missingOnSuccess) == 0 {
			// nothing to do if they are all the same
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		} else {
			// if one of them is different, we get the new values for all
			// and we update the notification
			allOnCancel := append(remoteOnCancel, missingOnCancel...)
			allOnFailure := append(remoteOnFailure, missingOnFailure...)
			allOnWarning := append(remoteOnWarning, missingOnWarning...)
			allOnSuccess := append(remoteOnSuccess, missingOnSuccess...)

			_, err := r.client.UpdateNotification(
				notificationID,
				dbt_cloud.Notification{
					AccountId:        r.client.AccountID,
					Id:               fullNotification.Id,
					UserId:           int(plan.UserID.ValueInt64()),
					OnCancel:         allOnCancel,
					OnFailure:        allOnFailure,
					OnWarning:        allOnWarning,
					OnSuccess:        allOnSuccess,
					State:            int(plan.State.ValueInt64()),
					NotificationType: int(plan.NotificationType.ValueInt64()),
					ExternalEmail:    plan.ExternalEmail.ValueStringPointer(),
					SlackChannelID:   plan.SlackChannelID.ValueStringPointer(),
					SlackChannelName: plan.SlackChannelName.ValueStringPointer(),
				},
			)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to update the existing notification",
					"Error: "+err.Error(),
				)
				return
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		}

	} else {
		// it doesn't exist so we create it
		notif, err := r.client.CreateNotification(
			int(plan.UserID.ValueInt64()),
			intOnCancel,
			intOnFailure,
			intOnWarning,
			intOnSuccess,
			int(plan.State.ValueInt64()),
			int(plan.NotificationType.ValueInt64()),
			plan.ExternalEmail.ValueStringPointer(),
			plan.SlackChannelID.ValueStringPointer(),
			plan.SlackChannelName.ValueStringPointer(),
		)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to create notification",
				"Error: "+err.Error(),
			)
			return
		}

		plan.ID = types.StringValue(strconv.Itoa(*notif.Id))
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *partialNotificationResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state notification.NotificationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	notificationID := state.ID.ValueString()
	notification, err := r.client.GetNotification(notificationID)
	if err != nil {
		resp.Diagnostics.AddError("Error getting the notification", err.Error())
		return
	}

	// we read the values from the config
	intOnCancel, intOnFailure, intOnWarning, intOnSuccess, ok := extractModelJobLists(state)
	if !ok {
		resp.Diagnostics.AddError("Error extracting the model job lists", "")
		return
	}

	configOnCancel := intOnCancel
	remoteOnCancel := notification.OnCancel
	requiredOnCancel, _ := lo.Difference(remoteOnCancel, configOnCancel)

	configOnFailure := intOnFailure
	remoteOnFailure := notification.OnFailure
	requiredOnFailure, _ := lo.Difference(remoteOnFailure, configOnFailure)

	configOnWarning := intOnWarning
	remoteOnWarning := notification.OnWarning
	requiredOnWarning := lo.Without(remoteOnWarning, configOnWarning...)

	configOnSuccess := intOnSuccess
	remoteOnSuccess := notification.OnSuccess
	requiredOnSuccess, _ := lo.Difference(remoteOnSuccess, configOnSuccess)

	if len(requiredOnCancel) > 0 || len(requiredOnFailure) > 0 || len(requiredOnWarning) > 0 || len(requiredOnSuccess) > 0 {
		// we update the notification if there are some jobs left
		// but we leave the notification existing, without deleting it entirely
		_, err = r.client.UpdateNotification(
			notificationID,
			dbt_cloud.Notification{
				AccountId:        r.client.AccountID,
				Id:               notification.Id,
				UserId:           int(state.UserID.ValueInt64()),
				OnCancel:         requiredOnCancel,
				OnFailure:        requiredOnFailure,
				OnWarning:        requiredOnWarning,
				OnSuccess:        requiredOnSuccess,
				State:            int(state.State.ValueInt64()),
				NotificationType: int(state.NotificationType.ValueInt64()),
				ExternalEmail:    state.ExternalEmail.ValueStringPointer(),
				SlackChannelID:   state.SlackChannelID.ValueStringPointer(),
				SlackChannelName: state.SlackChannelName.ValueStringPointer(),
			},
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update the existing notification",
				"Error: "+err.Error(),
			)
			return
		}
	} else {
		// we delete the notification if there are no jobs left
		notification.State = dbt_cloud.STATE_DELETED
		_, err = r.client.UpdateNotification(notificationID, *notification)
		if err != nil {
			resp.Diagnostics.AddError("Error deleting the notification", err.Error())
			return
		}
	}
}

func (r *partialNotificationResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state notification.NotificationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	notificationID := state.ID.ValueString()
	notification, err := r.client.GetNotification(notificationID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting the notification",
			"Error: "+err.Error(),
		)
		return
	}

	// we compare the partial objects and update them if needed
	intOnCancelPlan, intOnFailurePlan, intOnWarningPlan, intOnSuccessPlan, ok := extractModelJobLists(plan)
	if !ok {
		resp.Diagnostics.AddError("Error extracting the model job lists from the plan", "")
		return
	}

	intOnCancelState, intOnFailureState, intOnWarningState, intOnSuccessState, ok := extractModelJobLists(state)
	if !ok {
		resp.Diagnostics.AddError("Error extracting the model job lists from the state", "")
		return
	}

	remoteOnCancel := notification.OnCancel
	deletedOnCancel, newOnCancel := lo.Difference(intOnCancelState, intOnCancelPlan)
	requiredOnCancel, _ := lo.Difference(lo.Union(remoteOnCancel, newOnCancel), deletedOnCancel)

	remoteOnFailure := notification.OnFailure
	deletedOnFailure, newOnFailure := lo.Difference(intOnFailureState, intOnFailurePlan)
	requiredOnFailure, _ := lo.Difference(lo.Union(remoteOnFailure, newOnFailure), deletedOnFailure)

	remoteOnWarning := notification.OnWarning
	deletedOnWarning := lo.Without(intOnWarningState, intOnWarningPlan...)
	newOnWarning := lo.Without(intOnWarningPlan, intOnWarningState...)
	requiredOnWarning := lo.Without(lo.Union(remoteOnWarning, newOnWarning), deletedOnWarning...)

	remoteOnSuccess := notification.OnSuccess
	deletedOnSuccess, newOnSuccess := lo.Difference(intOnSuccessState, intOnSuccessPlan)
	requiredOnSuccess, _ := lo.Difference(lo.Union(remoteOnSuccess, newOnSuccess), deletedOnSuccess)

	// we check if there are changes to be sent, both global and local
	if plan.UserID != state.UserID ||
		plan.State != state.State ||
		len(deletedOnCancel) > 0 ||
		len(newOnCancel) > 0 ||
		len(deletedOnFailure) > 0 ||
		len(newOnFailure) > 0 ||
		len(deletedOnWarning) > 0 ||
		len(newOnWarning) > 0 ||
		len(deletedOnSuccess) > 0 ||
		len(newOnSuccess) > 0 {

		// we update the values to be the plan ones for global
		// and the calculated ones for the local ones
		_, err = r.client.UpdateNotification(
			notificationID,
			dbt_cloud.Notification{
				AccountId:        r.client.AccountID,
				Id:               notification.Id,
				UserId:           int(plan.UserID.ValueInt64()),
				OnCancel:         requiredOnCancel,
				OnWarning:        requiredOnWarning,
				OnFailure:        requiredOnFailure,
				OnSuccess:        requiredOnSuccess,
				State:            int(plan.State.ValueInt64()),
				NotificationType: int(plan.NotificationType.ValueInt64()),
				ExternalEmail:    plan.ExternalEmail.ValueStringPointer(),
				SlackChannelID:   plan.SlackChannelID.ValueStringPointer(),
				SlackChannelName: plan.SlackChannelName.ValueStringPointer(),
			},
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update the existing notification",
				"Error: "+err.Error(),
			)
			return
		}
	}

	// Set the updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *partialNotificationResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}
