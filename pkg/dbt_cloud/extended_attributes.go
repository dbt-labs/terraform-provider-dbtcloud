package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ExtendedAttributesResponse struct {
	Data   ExtendedAttributes `json:"data"`
	Status ResponseStatus     `json:"status"`
}

type ExtendedAttributes struct {
	ID                 *int            `json:"id,omitempty"`
	State              int             `json:"state,omitempty"`
	AccountID          int             `json:"account_id"`
	ProjectID          int             `json:"project_id"`
	ExtendedAttributes json.RawMessage `json:"extended_attributes"`
}

func (c *Client) GetExtendedAttributes(projectId int, extendedAttributesID int) (*ExtendedAttributes, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/extended-attributes/%d/", c.HostURL, c.AccountID, projectId, extendedAttributesID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	extendedAttributesResponse := ExtendedAttributesResponse{}
	err = json.Unmarshal(body, &extendedAttributesResponse)
	if err != nil {
		return nil, err
	}

	return &extendedAttributesResponse.Data, nil
}

func (c *Client) CreateExtendedAttributes(
	state int,
	projectId int,
	extendedAttributes json.RawMessage,
) (*ExtendedAttributes, error) {

	newExtendedAttributes := ExtendedAttributes{
		State:              state,
		AccountID:          c.AccountID,
		ProjectID:          projectId,
		ExtendedAttributes: extendedAttributes,
	}

	newExtendedAttributesData, err := json.Marshal(newExtendedAttributes)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/extended-attributes/", c.HostURL, c.AccountID, projectId), strings.NewReader(string(newExtendedAttributesData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	extendedAttributesResponse := ExtendedAttributesResponse{}
	err = json.Unmarshal(body, &extendedAttributesResponse)
	if err != nil {
		return nil, err
	}

	return &extendedAttributesResponse.Data, nil
}

func (c *Client) UpdateExtendedAttributes(projectId int, extendedAttributesID int, extendedAttributes ExtendedAttributes) (*ExtendedAttributes, error) {

	extendedAttributesData, err := json.Marshal(extendedAttributes)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/extended-attributes/%d/", c.HostURL, c.AccountID, projectId, extendedAttributesID), strings.NewReader(string(extendedAttributesData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	extendedAttributesResponse := ExtendedAttributesResponse{}
	err = json.Unmarshal(body, &extendedAttributesResponse)
	if err != nil {
		return nil, err
	}

	return &extendedAttributesResponse.Data, nil
}

func (c *Client) DeleteExtendedAttributes(projectId, extendedAttributesID int) (string, error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/extended-attributes/%d/", c.HostURL, c.AccountID, projectId, extendedAttributesID), nil)
	if err != nil {
		return "", err
	}

	_, err = c.doRequestWithRetry(req)
	if err != nil {
		return "", err
	}

	return "", err
}
