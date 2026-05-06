package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const APIVersionPrivate = "private"

// NotificationSettingChannel represents a delivery channel (Microsoft Teams)
// attached to a notification setting in the v3 notifications API.
type NotificationSettingChannel struct {
	ID             *int64  `json:"id,omitempty"`
	ChannelType    string  `json:"channel_type"`
	TeamsTeamID    *string `json:"teams_team_id,omitempty"`
	TeamsChannelID *string `json:"teams_channel_id,omitempty"`
}

// NotificationSettingRule represents a trigger rule attached to a notification setting.
// A nil JobID means the rule fires for all jobs in the account.
type NotificationSettingRule struct {
	ID        *int64  `json:"id,omitempty"`
	TriggerOn string  `json:"trigger_on"`
	JobID     *int64  `json:"job_id,omitempty"`
	JobName   *string `json:"job_name,omitempty"`
}

type NotificationSetting struct {
	ID          *int64                       `json:"id,omitempty"`
	AccountID   int64                        `json:"account_id,omitempty"`
	Name        string                       `json:"name"`
	Description *string                      `json:"description,omitempty"`
	IsActive    bool                         `json:"is_active"`
	Channels    []NotificationSettingChannel `json:"channels"`
	Rules       []NotificationSettingRule    `json:"rules"`
}

type notificationSettingResponse struct {
	Status ResponseStatus      `json:"status"`
	Data   NotificationSetting `json:"data"`
}

func (c *Client) GetNotificationSetting(id int64) (*NotificationSetting, error) {
	url := c.BuildAPIURL(APIVersionPrivate, fmt.Sprintf("accounts/%d/notification-settings/%d", c.AccountID, id))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	resp := notificationSettingResponse{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *Client) CreateNotificationSetting(setting NotificationSetting) (*NotificationSetting, error) {
	// account_id comes from the URL path, not the body.
	setting.AccountID = 0

	payload, err := json.Marshal(setting)
	if err != nil {
		return nil, err
	}

	url := c.BuildAPIURL(APIVersionPrivate, fmt.Sprintf("accounts/%d/notification-settings", c.AccountID))

	req, err := http.NewRequest("POST", url, strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	resp := notificationSettingResponse{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *Client) UpdateNotificationSetting(id int64, setting NotificationSetting) (*NotificationSetting, error) {
	setting.AccountID = 0

	payload, err := json.Marshal(setting)
	if err != nil {
		return nil, err
	}

	url := c.BuildAPIURL(APIVersionPrivate, fmt.Sprintf("accounts/%d/notification-settings/%d", c.AccountID, id))

	req, err := http.NewRequest("PATCH", url, strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	resp := notificationSettingResponse{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *Client) DeleteNotificationSetting(id int64) error {
	url := c.BuildAPIURL(APIVersionPrivate, fmt.Sprintf("accounts/%d/notification-settings/%d", c.AccountID, id))

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	if _, err := c.doRequestWithRetry(req); err != nil {
		return err
	}
	return nil
}
