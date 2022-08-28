package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type DatabricksCredentialListResponse struct {
	Data   []DatabricksCredential `json:"data"`
	Status ResponseStatus         `json:"status"`
}

type DatabricksCredentialResponse struct {
	Data   DatabricksCredential `json:"data"`
	Status ResponseStatus       `json:"status"`
}

type DatabricksCredentialFieldMetadataValidation struct {
	Required bool `json:"required"`
}

type DatabricksCredentialFieldMetadata struct {
	Label       string                                      `json:"label"`
	Description string                                      `json:"description"`
	Field_Type  string                                      `json:"field_type"`
	Encrypt     bool                                        `json:"encrypt"`
	Validation  DatabricksCredentialFieldMetadataValidation `json:"validation"`
}

type DatabricksCredentialField struct {
	Metadata DatabricksCredentialFieldMetadata `json:"metadata"`
	Value    string                            `json:"value"`
}

type DatabricksCredentialFields struct {
	Token  DatabricksCredentialField `json:"token"`
	Schema DatabricksCredentialField `json:"schema"`
}

type DatabricksCredentialDetails struct {
	Fields      DatabricksCredentialFields `json:"fields"`
	Field_Order []int                      `json:"field_order"`
}

type DatabricksCredential struct {
	ID                 *int                        `json:"id"`
	Account_Id         int                         `json:"account_id"`
	Project_Id         int                         `json:"project_id"`
	Type               string                      `json:"type"`
	State              int                         `json:"state"`
	Threads            int                         `json:"threads"`
	Target_Name        string                      `json:"target_name"`
	Adapter_Id         int                         `json:"adapter_id"`
	Credential_Details DatabricksCredentialDetails `json:"credential_details"`
}

func (c *Client) GetDatabricksCredential(projectId int, credentialId int) (*DatabricksCredential, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/credentials/", c.HostURL, c.AccountID, projectId), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	databricksCredentialListResponse := DatabricksCredentialListResponse{}
	err = json.Unmarshal(body, &databricksCredentialListResponse)
	if err != nil {
		return nil, err
	}

	for i, credential := range databricksCredentialListResponse.Data {
		if *credential.ID == credentialId {
			return &databricksCredentialListResponse.Data[i], nil
		}
	}

	return nil, fmt.Errorf("Did not find credential ID %d in project ID %d", credentialId, projectId)
}

func (c *Client) CreateDatabricksCredential(projectId int, type_ string, targetName string, adapterId int, numThreads int, token string) (*DatabricksCredential, error) {
	validation := DatabricksCredentialFieldMetadataValidation{
		Required: false,
	}
	tokenMetadata := DatabricksCredentialFieldMetadata{
		Label:       "Token",
		Description: "Personalized user token.",
		Field_Type:  "text",
		Encrypt:     true,
		Validation:  validation,
	}
	credentialsFieldToken := DatabricksCredentialField{
		Metadata: tokenMetadata,
		Value:    token,
	}
	credentialFields := DatabricksCredentialFields{
		Token:  credentialsFieldToken,
	}
	credentialDetails := DatabricksCredentialDetails{
		Fields:      credentialFields,
		Field_Order: []int{},
	}
	newDatabricksCredential := DatabricksCredential{
		Account_Id:         c.AccountID,
		Project_Id:         projectId,
		Type:               type_,
		State:              STATE_ACTIVE,
		Threads:            numThreads,
		Target_Name:        targetName,
		Adapter_Id:         adapterId,
		Credential_Details: credentialDetails,
	}

	newDatabricksCredentialData, err := json.Marshal(newDatabricksCredential)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/credentials/", c.HostURL, c.AccountID, projectId), strings.NewReader(string(newDatabricksCredentialData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	databricksCredentialResponse := DatabricksCredentialResponse{}
	err = json.Unmarshal(body, &databricksCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &databricksCredentialResponse.Data, nil
}

func (c *Client) UpdateDatabricksCredential(projectId int, credentialId int, databricksCredential DatabricksCredential) (*DatabricksCredential, error) {
	databricksCredentialData, err := json.Marshal(databricksCredential)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/credentials/%d/", c.HostURL, c.AccountID, projectId, credentialId), strings.NewReader(string(databricksCredentialData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	databricksCredentialResponse := DatabricksCredentialResponse{}
	err = json.Unmarshal(body, &databricksCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &databricksCredentialResponse.Data, nil
}

func (c *Client) DeleteDatabricksCredential(credentialId, projectId string) (string, error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v3/accounts/%d/projects/%s/credentials/%s/", c.HostURL, c.AccountID, projectId, credentialId), nil)
	if err != nil {
		return "", err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return "", err
	}

	return "", err
}
