package dbt_cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type StarburstCredentialResponse struct {
	Data   StarburstCredentialData `json:"data"`
	Status ResponseStatus          `json:"status"`
}

type StarburstUnencryptedCredentialDetails struct {
	Database   string `json:"database"`
	Schema     string `json:"schema"`
	TargetName string `json:"target_name"`
	Threads    int    `json:"threads"`
}

// StarburstCredentialData represents the data returned by the API for an Starburst credential
type StarburstCredentialData struct {
	ID                           *int                                  `json:"id"`
	AccountID                    int                                   `json:"account_id"`
	ProjectID                    int                                   `json:"project_id"`
	Type                         string                                `json:"type"`
	State                        int                                   `json:"state"`
	Threads                      int                                   `json:"threads"`
	TargetName                   string                                `json:"target_name"`
	AdapterVersion               string                                `json:"adapter_version,omitempty"`
	UnencryptedCredentialDetails StarburstUnencryptedCredentialDetails `json:"unencrypted_credential_details"`
}

// StarburstCredentialRequest is used for creating and updating Starburst credentials
// It doesn't include the UnencryptedCredentialDetails field which is only returned by the API
type StarburstCredentialRequest struct {
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

func (c *Client) GetStarburstCredential(
	projectId int,
	credentialId int,
) (*StarburstCredentialData, error) {
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

	credentialResponse := StarburstCredentialResponse{}
	err = json.Unmarshal(body, &credentialResponse)
	if err != nil {
		return nil, err
	}

	return &credentialResponse.Data, nil
}

func (c *Client) CreateStarburstCredential(
	ctx context.Context,
	projectId int,
	username string,
	password string,
	database string,
	schema string,
) (*StarburstCredentialData, error) {
	credentialDetails, err := GenerateStarburstCredentialDetails(
		username,
		password,
		database,
		schema,
	)
	if err != nil {
		return nil, err
	}

	credential := StarburstCredentialRequest{
		ID:                nil,
		AccountID:         c.AccountID,
		ProjectID:         projectId,
		Type:              "adapter",
		State:             STATE_ACTIVE,
		TargetName:        DEFAULT_TARGET_NAME,
		Threads:           DEFAULT_THREADS,
		CredentialDetails: credentialDetails,
		AdapterVersion:    "trino_v0",
	}

	rb, err := json.Marshal(credential)
	if err != nil {
		return nil, err
	}
	tflog.Debug(ctx, fmt.Sprintf("CreateStarburstCredential: %s", string(rb)))

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

	tflog.Debug(ctx, fmt.Sprintf("CreateStarburstCredentialResponse: %s", string(body)))

	credentialResponse := StarburstCredentialResponse{}
	err = json.Unmarshal(body, &credentialResponse)
	if err != nil {
		return nil, err
	}

	return &credentialResponse.Data, nil
}

func (c *Client) UpdateStarburstCredential(
	projectId int,
	credentialId int,
	starburstCredential StarburstCredentialRequest,
) (*StarburstCredentialData, error) {
	rb, err := json.Marshal(starburstCredential)
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

	credentialResponse := StarburstCredentialResponse{}
	err = json.Unmarshal(body, &credentialResponse)
	if err != nil {
		return nil, err
	}

	return &credentialResponse.Data, nil
}

func GenerateStarburstCredentialDetails(
	username string,
	password string,
	database string,
	schema string,
) (AdapterCredentialDetails, error) {
	// Create the credential details structure based on the payload example
	fields := map[string]AdapterCredentialField{
		"user": {
			Metadata: AdapterCredentialFieldMetadata{
				Label:        "Starburst/Trino username",
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
				Label:        "Starburst/Trino password",
				Description:  "User's password",
				Field_Type:   "text",
				Encrypt:      true,
				Overrideable: false,
				Validation:   AdapterCredentialFieldMetadataValidation{Required: true},
			},
			Value: password,
		},
		"database": {
			Metadata: AdapterCredentialFieldMetadata{
				Label:        "Database",
				Description:  "The catalog",
				Field_Type:   "text",
				Encrypt:      false,
				Overrideable: false,
				Validation:   AdapterCredentialFieldMetadataValidation{Required: true},
			},
			Value: database,
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
			Value: nil,
		},
	}

	return AdapterCredentialDetails{Fields: fields}, nil
}
