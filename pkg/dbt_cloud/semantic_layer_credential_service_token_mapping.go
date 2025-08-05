package dbt_cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type SemanticLayerCredentialServiceTokenMapping struct {
	ID                        *int `json:"id"`
	AccountID                 int  `json:"account_id"`
	ProjectID                 int  `json:"project_id"`
	SemanticLayerCredentialID int  `json:"semantic_layer_credentials_id"`
	ServiceTokenID            int  `json:"service_token_id"`
	State                     int  `json:"state,omitempty"`
}

// response type for single semantic layer credential service token mapping
type SemanticLayerCredentialServiceTokenMappingResponse struct {
	Data   SemanticLayerCredentialServiceTokenMapping `json:"data"`
	Status ResponseStatus                             `json:"status"`
}

// response type for multiple semantic layer credential service token mappings
type SemanticLayerCredentialServiceTokenMappingArrayResponse struct {
	Data   []SemanticLayerCredentialServiceTokenMapping `json:"data"`
	Status ResponseStatus                               `json:"status"`
}

func (c *Client) CreateSemanticLayerCredentialServiceTokenMapping(
	projectId int,
	semanticLayerCredentialId int,
	serviceTokenId int,
) (*SemanticLayerCredentialServiceTokenMapping, error) {
	newMapping := SemanticLayerCredentialServiceTokenMapping{
		AccountID:                 int(c.AccountID),
		ProjectID:                 projectId,
		SemanticLayerCredentialID: semanticLayerCredentialId,
		ServiceTokenID:            serviceTokenId,
		State:                     1,
	}

	payload, err := json.Marshal(newMapping)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%d/semantic-layer-credential-to-service-token-mapping/",
			c.HostURL,
			c.AccountID,
		),
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	var mappingResponse SemanticLayerCredentialServiceTokenMappingResponse
	err = json.Unmarshal(body, &mappingResponse)
	if err != nil {
		return nil, err
	}

	return &mappingResponse.Data, nil
}

func (c *Client) GetSemanticLayerCredentialServiceTokenMapping(
	sm SemanticLayerCredentialServiceTokenMapping,
) (*SemanticLayerCredentialServiceTokenMapping, error) {
	query := fmt.Sprintf("project_id=%d", sm.ProjectID)
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/v3/accounts/%d/semantic-layer-credential-to-service-token-mapping/?%s", c.HostURL, c.AccountID, query),
		nil,
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	var credentialsResponse SemanticLayerCredentialServiceTokenMappingArrayResponse
	err = json.Unmarshal(body, &credentialsResponse)
	if err != nil {
		return nil, err
	}

	var SemanticCredentialTokenMapping SemanticLayerCredentialServiceTokenMapping

	for _, mapping := range credentialsResponse.Data {
		if *mapping.ID == *sm.ID {
			SemanticCredentialTokenMapping = mapping
			break
		}
	}

	if SemanticCredentialTokenMapping.SemanticLayerCredentialID == 0 {
		return nil, fmt.Errorf("resource-not-found: semantic layer credential service token mapping not found for ID %d", *sm.ID)
	}

	return &SemanticCredentialTokenMapping, nil
}

func (c *Client) DeleteSemanticLayerCredentialServiceTokenMapping(id int) error {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s/v3/accounts/%d/semantic-layer-credential-to-service-token-mapping/%d/", c.HostURL, c.AccountID, id),
		nil,
	)
	if err != nil {
		return err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return err
	}

	var response struct {
		Status ResponseStatus `json:"status"`
		Data   interface{}    `json:"data"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %s", err)
	}

	return nil
}
