package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type SemanticLayerCredentials struct {
	ID             *int                   `json:"id"`
	Name           string                 `json:"name"`
	ProjectID      int                    `json:"project_id"`
	AccountID      int                    `json:"account_id"`
	Values         map[string]interface{} `json:"values"`
	AdapterVersion string                 `json:"adapter_version"`
	SchemaType     string                 `json:"schema_type"`
}

type SemanticLayerCredentialsFilter struct {
	ProjectID int
}

type SemanticLayerCredentialsResponse struct {
	Status ResponseStatus             `json:"status"`
	Data   []SemanticLayerCredentials `json:"data"`
}

type SemanticLayerCredentialResponse struct {
	Status ResponseStatus           `json:"status"`
	Data   SemanticLayerCredentials `json:"data"`
}

func (c *Client) GetSemanticLayerCredential(id int64) (*SemanticLayerCredentials, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%s/semantic-layer-credentials/%d",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			id,
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

	var credentialsResponse SemanticLayerCredentialResponse
	err = json.Unmarshal(body, &credentialsResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %s", err)
	}

	return &credentialsResponse.Data, nil
}

func (c *Client) CreateSemanticLayerCredential(
	projectId int64,
	values map[string]interface{},
	name string,
	adapterVersion string,

) (*SemanticLayerCredentials, error) {

	newCredential := SemanticLayerCredentials{
		SchemaType:     "semantic_layer_credentials",
		AccountID:      c.AccountID,
		ProjectID:      int(projectId),
		Name:           name,
		AdapterVersion: adapterVersion,
		Values:         values,
	}

	newCredentialsData, err := json.Marshal(newCredential)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%d/semantic-layer-credentials/",
			c.HostURL,
			c.AccountID,
		),
		strings.NewReader(string(newCredentialsData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	credentialResponse := SemanticLayerCredentialResponse{}
	err = json.Unmarshal(body, &credentialResponse)
	if err != nil {
		return nil, err
	}

	return &credentialResponse.Data, nil
}

func (c *Client) UpdateSemanticLayerCredential(
	credentialId int64,
	credential SemanticLayerCredentials) (*SemanticLayerCredentials, error) {

	configData, err := json.Marshal(credential)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%d/semantic-layer-credentials/%d/",
			c.HostURL,
			c.AccountID,
			credentialId,
		),
		strings.NewReader(string(configData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	response := SemanticLayerCredentialResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func (c *Client) DeleteSemanticLayerCredential(
	projectId int64,
	credentialId int64,
) error {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf(
			"%s/v3/accounts/%d/semantic-layer-credentials/%d/",
			c.HostURL,
			c.AccountID,
			credentialId,
		),
		nil,
	)
	if err != nil {
		return err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return err
	}

	response := SemanticLayerCredentialResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	return nil
}
