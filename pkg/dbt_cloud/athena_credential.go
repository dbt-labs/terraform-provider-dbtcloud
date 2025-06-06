package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	DEFAULT_THREADS     = 4
	DEFAULT_TARGET_NAME = "default"
	DEFAULT_ATHENA_AUTH = "iam_user"
)

type AthenaCredentialResponse struct {
	Data   AthenaCredentialData `json:"data"`
	Status ResponseStatus       `json:"status"`
}

type AthenaUnencryptedCredentialDetails struct {
	Schema     string `json:"schema"`
	TargetName string `json:"target_name"`
	Threads    int    `json:"threads"`
}

// AthenaCredentialData represents the data returned by the API for an Athena credential
type AthenaCredentialData struct {
	ID                           *int                               `json:"id"`
	AccountID                    int                                `json:"account_id"`
	ProjectID                    int                                `json:"project_id"`
	Type                         string                             `json:"type"`
	State                        int                                `json:"state"`
	Threads                      int                                `json:"threads"`
	TargetName                   string                             `json:"target_name"`
	AdapterVersion               string                             `json:"adapter_version,omitempty"`
	UnencryptedCredentialDetails AthenaUnencryptedCredentialDetails `json:"unencrypted_credential_details"`
}

// AthenaCredentialRequest is used for creating and updating Athena credentials
// It doesn't include the UnencryptedCredentialDetails field which is only returned by the API
type AthenaCredentialRequest struct {
	ID                *int                     `json:"id,omitempty"`
	AccountID         int                      `json:"account_id"`
	ProjectID         int                      `json:"project_id"`
	Type              string                   `json:"type"`
	State             int                      `json:"state"`
	Threads           int                      `json:"threads"`
	TargetName        string                   `json:"target_name"`
	AdapterVersion    string                   `json:"adapter_version,omitempty"`
	CredentialDetails AdapterCredentialDetails `json:"credential_details"`
}

func (c *Client) GetAthenaCredential(
	projectId int,
	credentialId int,
) (*AthenaCredentialData, error) {
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

	credentialResponse := AthenaCredentialResponse{}
	err = json.Unmarshal(body, &credentialResponse)
	if err != nil {
		return nil, err
	}

	return &credentialResponse.Data, nil
}

func (c *Client) CreateAthenaCredential(
	projectId int,
	awsAccessKeyId string,
	awsSecretAccessKey string,
	schema string,
	adapterVersion string,
) (*AthenaCredentialData, error) {
	credentialDetails, err := GenerateAthenaCredentialDetails(
		awsAccessKeyId,
		awsSecretAccessKey,
		schema,
	)
	if err != nil {
		return nil, err
	}

	credential := AthenaCredentialRequest{
		ID:                nil,
		AccountID:         c.AccountID,
		ProjectID:         projectId,
		Type:              "adapter",
		State:             STATE_ACTIVE,
		TargetName:        DEFAULT_TARGET_NAME,
		Threads:           DEFAULT_THREADS,
		CredentialDetails: credentialDetails,
		AdapterVersion:    adapterVersion,
	}

	rb, err := json.Marshal(credential)
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
		strings.NewReader(string(rb)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	credentialResponse := AthenaCredentialResponse{}
	err = json.Unmarshal(body, &credentialResponse)
	if err != nil {
		return nil, err
	}

	return &credentialResponse.Data, nil
}

func (c *Client) UpdateAthenaCredential(
	projectId int,
	credentialId int,
	athenaCredential AthenaCredentialRequest,
) (*AthenaCredentialData, error) {
	rb, err := json.Marshal(athenaCredential)
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
		strings.NewReader(string(rb)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	credentialResponse := AthenaCredentialResponse{}
	err = json.Unmarshal(body, &credentialResponse)
	if err != nil {
		return nil, err
	}

	return &credentialResponse.Data, nil
}

func GenerateAthenaCredentialDetails(
	awsAccessKeyId string,
	awsSecretAccessKey string,
	schema string,
) (AdapterCredentialDetails, error) {
	// Create the credential details structure based on the payload example
	fields := map[string]AdapterCredentialField{
		"auth_type": {
			Metadata: AdapterCredentialFieldMetadata{
				Label:        "Authentication method",
				Description:  "",
				Field_Type:   "hidden",
				Encrypt:      false,
				Overrideable: false,
				Validation:   AdapterCredentialFieldMetadataValidation{Required: false},
			},
			Value: DEFAULT_ATHENA_AUTH,
		},
		"aws_access_key_id": {
			Metadata: AdapterCredentialFieldMetadata{
				Label:        "AWS access key ID",
				Description:  "Access key ID of the user performing requests",
				Field_Type:   "text",
				Encrypt:      true,
				Overrideable: false,
				Validation:   AdapterCredentialFieldMetadataValidation{Required: true},
			},
			Value: awsAccessKeyId,
		},
		"aws_secret_access_key": {
			Metadata: AdapterCredentialFieldMetadata{
				Label:        "AWS secret access key",
				Description:  "Secret access key of the user performing requests",
				Field_Type:   "text",
				Encrypt:      true,
				Overrideable: false,
				Validation:   AdapterCredentialFieldMetadataValidation{Required: true},
			},
			Value: awsSecretAccessKey,
		},
		"schema": {
			Metadata: AdapterCredentialFieldMetadata{
				Label:        "Schema",
				Description:  "Specify the schema (Athena database) to build models into (lowercase only)",
				Field_Type:   "text",
				Encrypt:      false,
				Overrideable: false,
				Validation:   AdapterCredentialFieldMetadataValidation{Required: true},
			},
			Value: schema,
		},
		"threads": {
			Metadata: AdapterCredentialFieldMetadata{
				Label:        "Threads",
				Description:  "The number of threads to use for dbt operations.",
				Field_Type:   "number",
				Encrypt:      false,
				Overrideable: false,
				Validation:   AdapterCredentialFieldMetadataValidation{Required: false},
			},
			Value: nil,
		},
	}

	return AdapterCredentialDetails{Fields: fields}, nil
}
