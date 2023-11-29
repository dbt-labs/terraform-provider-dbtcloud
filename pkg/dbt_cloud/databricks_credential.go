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

type DatabricksUnencryptedCredentialDetails struct {
	Catalog    string `json:"catalog"`
	Schema     string `json:"schema"`
	TargetName string `json:"target_name"`
	Threads    int    `json:"threads"`
	Token      string `json:"token,omitempty"`
}

type DatabricksCredential struct {
	ID                           *int                                   `json:"id"`
	Account_Id                   int                                    `json:"account_id"`
	Project_Id                   int                                    `json:"project_id"`
	Type                         string                                 `json:"type"`
	State                        int                                    `json:"state"`
	Threads                      int                                    `json:"threads"`
	Target_Name                  string                                 `json:"target_name"`
	Adapter_Id                   int                                    `json:"adapter_id"`
	Credential_Details           AdapterCredentialDetails               `json:"credential_details"`
	UnencryptedCredentialDetails DatabricksUnencryptedCredentialDetails `json:"unencrypted_credential_details"`
}

func (c *Client) GetDatabricksCredential(
	projectId int,
	credentialId int,
) (*DatabricksCredential, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/credentials/%d/?include_related=[adapter]",
			c.HostURL,
			c.AccountID,
			projectId,
			credentialId,
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

	credentialResponse := DatabricksCredentialResponse{}
	err = json.Unmarshal(body, &credentialResponse)
	if err != nil {
		return nil, err
	}

	return &credentialResponse.Data, nil
}

func (c *Client) CreateDatabricksCredential(
	projectId int,
	type_ string,
	targetName string,
	adapterId int,
	token string,
	catalog string,
	schema string,
	adapterType string,
) (*DatabricksCredential, error) {
	validation := AdapterCredentialFieldMetadataValidation{
		Required: false,
	}
	tokenMetadata := AdapterCredentialFieldMetadata{
		Label:       "Token",
		Description: "Personalized user token.",
		Field_Type:  "text",
		Encrypt:     true,
		Validation:  validation,
	}
	catalogMetadata := AdapterCredentialFieldMetadata{
		Label:       "Catalog",
		Description: "Catalog name if Unity Catalog is enabled in your Databricks workspace.  Only available in dbt version 1.1 and later.",
		Field_Type:  "text",
		Encrypt:     false,
		Validation:  validation,
	}
	schemaMetadata := AdapterCredentialFieldMetadata{
		Label:       "Schema",
		Description: "User schema.",
		Field_Type:  "text",
		Encrypt:     false,
		Validation:  validation,
	}
	threadsMetadata := AdapterCredentialFieldMetadata{
		Label:       "Threads",
		Description: "The number of threads to use for your jobs.",
		Field_Type:  "number",
		Encrypt:     false,
		Validation:  validation,
	}

	credentialsFieldToken := AdapterCredentialField{
		Metadata: tokenMetadata,
		Value:    token,
	}
	credentialsFieldCatalog := AdapterCredentialField{
		Metadata: catalogMetadata,
		Value:    catalog,
	}
	credentialsFieldSchema := AdapterCredentialField{
		Metadata: schemaMetadata,
		Value:    schema,
	}
	credentialsFieldThreads := AdapterCredentialField{
		Metadata: threadsMetadata,
		Value:    NUM_THREADS_CREDENTIAL,
	}

	credentialFields := map[string]AdapterCredentialField{}

	// the catalog field is only available for databricks adapter type
	// there is an issue if we provide the number of threads at the creation
	if adapterType == "databricks" {
		credentialFields["catalog"] = credentialsFieldCatalog
		credentialFields["token"] = credentialsFieldToken
		credentialFields["schema"] = credentialsFieldSchema
	}

	// for spark, we use all except the catalog
	if adapterType == "spark" {
		credentialFields["token"] = credentialsFieldToken
		credentialFields["schema"] = credentialsFieldSchema
		credentialFields["threads"] = credentialsFieldThreads
	}

	credentialDetails := AdapterCredentialDetails{
		Fields:      credentialFields,
		Field_Order: []string{},
	}
	newDatabricksCredential := DatabricksCredential{
		Account_Id:         c.AccountID,
		Project_Id:         projectId,
		Type:               type_,
		State:              STATE_ACTIVE,
		Threads:            NUM_THREADS_CREDENTIAL,
		Target_Name:        targetName,
		Adapter_Id:         adapterId,
		Credential_Details: credentialDetails,
	}

	newDatabricksCredentialData, err := json.Marshal(newDatabricksCredential)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/credentials/",
			c.HostURL,
			c.AccountID,
			projectId,
		),
		strings.NewReader(string(newDatabricksCredentialData)),
	)
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

func (c *Client) UpdateDatabricksCredential(
	projectId int,
	credentialId int,
	databricksCredential DatabricksCredential,
) (*DatabricksCredential, error) {
	databricksCredentialData, err := json.Marshal(databricksCredential)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/credentials/%d/",
			c.HostURL,
			c.AccountID,
			projectId,
			credentialId,
		),
		strings.NewReader(string(databricksCredentialData)),
	)
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
