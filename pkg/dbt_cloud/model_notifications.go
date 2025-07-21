package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type ModelNotificationsResponse struct {
	Data   ModelNotifications `json:"data"`
	Status ResponseStatus     `json:"status"`
}

type ModelNotifications struct {
	ID            *int   `json:"id,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
	AccountID     int    `json:"account_id,omitempty"`
	EnvironmentID int    `json:"environment_id"`
	Enabled       bool   `json:"enabled"`
	OnSuccess     bool   `json:"on_success"`
	OnFailure     bool   `json:"on_failure"`
	OnWarning     bool   `json:"on_warning"`
	OnSkipped     bool   `json:"on_skipped"`
}

func (c *Client) GetModelNotifications(environmentID string) (*ModelNotifications, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%s/environments/%s/model-notifications/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			environmentID,
		),
		nil,
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	modelNotificationsResponse := ModelNotificationsResponse{}
	err = json.Unmarshal(body, &modelNotificationsResponse)
	if err != nil {
		return nil, err
	}

	return &modelNotificationsResponse.Data, nil
}

func (c *Client) CreateModelNotifications(
	environmentID string,
	enabled bool,
	onSuccess bool,
	onFailure bool,
	onWarning bool,
	onSkipped bool) (*ModelNotifications, error) {

	envID, err := strconv.Atoi(environmentID)
	if err != nil {
		return nil, err
	}

	modelNotifications := ModelNotifications{
		EnvironmentID: envID,
		Enabled:       enabled,
		OnSuccess:     onSuccess,
		OnFailure:     onFailure,
		OnWarning:     onWarning,
		OnSkipped:     onSkipped,
	}

	return c.UpdateModelNotifications(environmentID, modelNotifications)
}

func (c *Client) UpdateModelNotifications(
	environmentID string,
	modelNotifications ModelNotifications,
) (*ModelNotifications, error) {
	// Only include the fields that can be updated in the payload
	updatePayload := struct {
		Enabled   bool `json:"enabled"`
		OnSuccess bool `json:"on_success"`
		OnFailure bool `json:"on_failure"`
		OnWarning bool `json:"on_warning"`
		OnSkipped bool `json:"on_skipped"`
	}{
		Enabled:   modelNotifications.Enabled,
		OnSuccess: modelNotifications.OnSuccess,
		OnFailure: modelNotifications.OnFailure,
		OnWarning: modelNotifications.OnWarning,
		OnSkipped: modelNotifications.OnSkipped,
	}

	modelNotificationsData, err := json.Marshal(updatePayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/environments/%s/model-notifications/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			environmentID,
		),
		strings.NewReader(string(modelNotificationsData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	modelNotificationsResponse := ModelNotificationsResponse{}
	err = json.Unmarshal(body, &modelNotificationsResponse)
	if err != nil {
		return nil, err
	}

	return &modelNotificationsResponse.Data, nil
}
