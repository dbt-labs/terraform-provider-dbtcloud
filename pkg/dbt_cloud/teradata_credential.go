package dbt_cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type TeradataCredentialResponse struct {
	Data   TeradataCredentialData `json:"data"`
	Status ResponseStatus         `json:"status"`
}

type TeradataUnencryptedCredentialDetails struct {
	Schema     string `json:"schema"`
	TargetName string `json:"target_name"`
	Threads    int    `json:"threads"`
}

// TeradataCredentialData represents the data returned by the API for an Teradata credential
type TeradataCredentialData struct {
	ID                           *int                                 `json:"id"`
	AccountID                    int                                  `json:"account_id"`
	Threads                      int                                  `json:"threads"`
	TargetName                   string                               `json:"target_name"`
	AdapterVersion               string                               `json:"adapter_version,omitempty"`
	UnencryptedCredentialDetails TeradataUnencryptedCredentialDetails `json:"unencrypted_credential_details"`
}

// TeradataCredentialRequest is used for creating and updating Teradata credentials
// It doesn't include the UnencryptedCredentialDetails field which is only returned by the API
type TeradataCredentialRequest struct {
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

func (c *Client) GetTeradataCredential(
	projectId int,
	credentialId int,
) (*TeradataCredentialData, error) {
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

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	credentialResponse := TeradataCredentialResponse{}
	err = json.Unmarshal(body, &credentialResponse)
	if err != nil {
		return nil, err
	}

	return &credentialResponse.Data, nil
}

func (c *Client) CreateTeradataCredential(
	ctx context.Context,
	projectId int,
	username string,
	password string,
	schema string,
	threads int,
) (*TeradataCredentialData, error) {
	credentialDetails, err := GenerateTeradataCredentialDetails(
		username,
		password,
		schema,
		threads,
	)
	if err != nil {
		return nil, err
	}

	credential := TeradataCredentialRequest{
		ID:                nil,
		AccountID:         int(c.AccountID),
		ProjectID:         projectId,
		Type:              "adapter",
		State:             STATE_ACTIVE,
		TargetName:        DEFAULT_TARGET_NAME,
		Threads:           threads,
		CredentialDetails: credentialDetails,
		AdapterVersion:    "teradata_v0",
	}

	rb, err := json.Marshal(credential)
	if err != nil {
		return nil, err
	}
	tflog.Debug(ctx, fmt.Sprintf("CreateTeradataCredential: %s", string(rb)))

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

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	tflog.Debug(ctx, fmt.Sprintf("CreateTeradataCredentialResponse: %s", string(body)))

	credentialResponse := TeradataCredentialResponse{}
	err = json.Unmarshal(body, &credentialResponse)
	if err != nil {
		return nil, err
	}

	return &credentialResponse.Data, nil
}

func (c *Client) UpdateTeradataCredential(
	projectId int,
	credentialId int,
	teradataCredential TeradataCredentialRequest,
) (*TeradataCredentialData, error) {
	rb, err := json.Marshal(teradataCredential)
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

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	credentialResponse := TeradataCredentialResponse{}
	err = json.Unmarshal(body, &credentialResponse)
	if err != nil {
		return nil, err
	}

	return &credentialResponse.Data, nil
}

func GenerateTeradataCredentialDetails(
	username string,
	password string,
	schema string,
	threads int,
) (AdapterCredentialDetails, error) {
	// Create the credential details structure based on the payload example
	fields := map[string]AdapterCredentialField{
		"user": {
			Metadata: AdapterCredentialFieldMetadata{
				Label:        "Teradata username",
				Description:  "The username",
				Field_Type:   "text",
				Encrypt:      false,
				Overrideable: false,
				Validation:   AdapterCredentialFieldMetadataValidation{Required: true},
			},
			Value: username,
		},
		"password": {
			Metadata: AdapterCredentialFieldMetadata{
				Label:        "Teradata password",
				Description:  "User's password",
				Field_Type:   "text",
				Encrypt:      true,
				Overrideable: false,
				Validation:   AdapterCredentialFieldMetadataValidation{Required: true},
			},
			Value: password,
		},
		"schema": {
			Metadata: AdapterCredentialFieldMetadata{
				Label:        "Schema",
				Description:  "The schema to build models into",
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
				Description:  "The number of threads to use for dbt operations",
				Field_Type:   "number",
				Encrypt:      false,
				Overrideable: false,
				Validation:   AdapterCredentialFieldMetadataValidation{Required: true},
			},
			Value: threads,
		},
	}

	return AdapterCredentialDetails{Fields: fields}, nil
}
