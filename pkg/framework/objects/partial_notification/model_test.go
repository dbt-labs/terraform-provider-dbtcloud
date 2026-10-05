package partial_notification

import (
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/notification"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMatchPartialSlackV2(t *testing.T) {
	model := notification.NotificationResourceModel{
		NotificationType: types.Int64Value(5),
		SlackChannelID:   types.StringValue("C_CHANNEL_B"),
		SlackChannelName: types.StringValue("channel-b"),
	}

	channelA, channelB := "C_CHANNEL_A", "C_CHANNEL_B"
	nameA, nameB := "channel-a", "#channel-b"

	tests := []struct {
		name     string
		response dbt_cloud.Notification
		want     bool
	}{
		{
			name:     "different channel ID does not match",
			response: dbt_cloud.Notification{NotificationType: 5, SlackChannelID: &channelA, SlackChannelName: &nameA},
			want:     false,
		},
		{
			name:     "same channel ID matches even if the name differs",
			response: dbt_cloud.Notification{NotificationType: 5, SlackChannelID: &channelB, SlackChannelName: &nameB},
			want:     true,
		},
		{
			name:     "legacy slack type does not match",
			response: dbt_cloud.Notification{NotificationType: 2, SlackChannelID: &channelB, SlackChannelName: &nameB},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchPartial(model, tt.response); got != tt.want {
				t.Errorf("matchPartial() = %v, want %v", got, tt.want)
			}
		})
	}
}
