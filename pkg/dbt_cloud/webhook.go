package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type WebhookResponse struct {
	Data   WebhookRead    `json:"data"`
	Status ResponseStatus `json:"status"`
}

type WebhookRead struct {
	WebhookId         string   `json:"id"`
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	ClientUrl         string   `json:"client_url"`
	EventTypes        []string `json:"event_types,omitempty"`
	JobIds            []string `json:"job_ids"`
	Active            bool     `json:"active,omitempty"`
	HmacSecret        *string  `json:"hmac_secret,omitempty"`
	HttpStatusCode    *string  `json:"http_status_code,omitempty"`
	AccountIdentifier *string  `json:"account_identifier,omitempty"`
}

type WebhookWrite struct {
	WebhookId   string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	ClientUrl   string   `json:"client_url"`
	EventTypes  []string `json:"event_types,omitempty"`
	JobIds      []int64  `json:"job_ids"`
	Active      bool     `json:"active,omitempty"`
}

func (c *Client) GetWebhook(webhookID string) (*WebhookRead, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%s/webhooks/subscription/%s",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			webhookID,
		),
		nil,
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	webhookResponse := WebhookResponse{}
	err = json.Unmarshal(body, &webhookResponse)
	if err != nil {
		return nil, err
	}

	return &webhookResponse.Data, nil
}

func (c *Client) CreateWebhook(
	webhookId string,
	name string,
	description string,
	clientUrl string,
	eventTypes []string,
	jobIds []int64,
	active bool,
) (*WebhookRead, error) {

	newWebhook := WebhookWrite{
		WebhookId:   webhookId,
		Name:        name,
		Description: description,
		ClientUrl:   clientUrl,
		EventTypes:  eventTypes,
		JobIds:      jobIds,
		Active:      active,
	}

	newWebhookData, err := json.Marshal(newWebhook)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/webhooks/subscriptions",
			c.HostURL,
			strconv.Itoa(c.AccountID),
		),
		strings.NewReader(string(newWebhookData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	webhookResponse := WebhookResponse{}
	err = json.Unmarshal(body, &webhookResponse)
	if err != nil {
		return nil, err
	}

	return &webhookResponse.Data, nil
}

func (c *Client) UpdateWebhook(webhookId string, webhook WebhookWrite) (*WebhookRead, error) {
	webhookData, err := json.Marshal(webhook)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf(
			"%s/v3/accounts/%s/webhooks/subscription/%s",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			webhookId,
		),
		strings.NewReader(string(webhookData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	webhookResponse := WebhookResponse{}
	err = json.Unmarshal(body, &webhookResponse)
	if err != nil {
		return nil, err
	}

	return &webhookResponse.Data, nil
}

func (c *Client) DeleteWebhook(webhookId string) (string, error) {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf(
			"%s/v3/accounts/%s/webhooks/subscription/%s",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			webhookId,
		),
		nil,
	)
	if err != nil {
		return "", err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return "", err
	}

	return "", err
}
