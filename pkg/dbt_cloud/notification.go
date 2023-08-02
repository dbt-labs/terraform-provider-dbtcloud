package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type NotificationResponse struct {
	Data   Notification   `json:"data"`
	Status ResponseStatus `json:"status"`
}

type Notification struct {
	Id               *int    `json:"id,omitempty"`
	AccountId        int     `json:"account_id"`
	UserId           int     `json:"user_id"`
	OnCancel         []int   `json:"on_cancel"`
	OnFailure        []int   `json:"on_failure"`
	OnSuccess        []int   `json:"on_success"`
	State            int     `json:"state"`
	NotificationType int     `json:"type"`
	ExternalEmail    *string `json:"external_email"`
}

func (c *Client) GetNotification(notificationID string) (*Notification, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v2/accounts/%s/notifications/%s/", c.HostURL, strconv.Itoa(c.AccountID), notificationID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	notificationResponse := NotificationResponse{}
	err = json.Unmarshal(body, &notificationResponse)
	if err != nil {
		return nil, err
	}

	return &notificationResponse.Data, nil
}

func (c *Client) CreateNotification(
	userId int,
	onCancel []int,
	onFailure []int,
	onSuccess []int,
	state int,
	notificationType int,
	externalEmail *string) (*Notification, error) {

	newNotification := Notification{
		AccountId:        c.AccountID,
		UserId:           userId,
		OnCancel:         onCancel,
		OnFailure:        onFailure,
		OnSuccess:        onSuccess,
		State:            state,
		NotificationType: notificationType,
		ExternalEmail:    externalEmail,
	}

	newNotificationData, err := json.Marshal(newNotification)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v2/accounts/%s/notifications/", c.HostURL, strconv.Itoa(c.AccountID)), strings.NewReader(string(newNotificationData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	notificationResponse := NotificationResponse{}
	err = json.Unmarshal(body, &notificationResponse)
	if err != nil {
		return nil, err
	}

	return &notificationResponse.Data, nil
}

func (c *Client) UpdateNotification(notificationId string, notification Notification) (*Notification, error) {
	notificationData, err := json.Marshal(notification)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v2/accounts/%s/notifications/%s/", c.HostURL, strconv.Itoa(c.AccountID), notificationId), strings.NewReader(string(notificationData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	notificationResponse := NotificationResponse{}
	err = json.Unmarshal(body, &notificationResponse)
	if err != nil {
		return nil, err
	}

	return &notificationResponse.Data, nil
}
