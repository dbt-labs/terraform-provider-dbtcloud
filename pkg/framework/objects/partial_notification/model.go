package partial_notification

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/notification"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func matchPartial(
	notificationModel notification.NotificationResourceModel,
	notificationResponse dbt_cloud.Notification,
) bool {
	if notificationModel.NotificationType != types.Int64Value(
		int64(notificationResponse.NotificationType),
	) {
		return false
	}
	switch notificationResponse.NotificationType {
	case 1:
		// internal notification
		if !(notificationModel.UserID == types.Int64Value(int64(notificationResponse.UserId))) {
			return false
		}
	case 2:
		// slack notification
		if !(notificationModel.SlackChannelID == types.StringPointerValue(
			notificationResponse.SlackChannelID,
		)) {
			return false
		}
		if !(notificationModel.SlackChannelName == types.StringPointerValue(
			notificationResponse.SlackChannelName,
		)) {
			return false
		}
	case 4:
		// external notification
		if !(notificationModel.ExternalEmail == types.StringPointerValue(
			notificationResponse.ExternalEmail,
		)) {
			return false
		}
	}
	return true
}

func extractModelJobLists(
	data notification.NotificationResourceModel,
) (intOnCancel, intOnFailure, intOnSuccess []int, ok bool) {

	diags := data.OnCancel.ElementsAs(context.Background(), &intOnCancel, false)
	if diags.HasError() {
		return nil, nil, nil, false
	}
	diags = data.OnFailure.ElementsAs(context.Background(), &intOnFailure, false)
	if diags.HasError() {
		return nil, nil, nil, false
	}

	diags = data.OnSuccess.ElementsAs(context.Background(), &intOnSuccess, false)
	if diags.HasError() {
		return nil, nil, nil, false
	}
	return intOnCancel, intOnFailure, intOnSuccess, true
}
