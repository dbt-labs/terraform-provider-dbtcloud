package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type BigQueryCredentialResponse struct {
	Data   BigQueryCredential `json:"data"`
	Status ResponseStatus     `json:"status"`
}

// BigQueryUnencryptedCredentialDetails contains the readable credential values for v1 credentials
type BigQueryUnencryptedCredentialDetails struct {
	Schema                         string `json:"schema"`
	Threads                        int    `json:"threads"`
	AuthType                       string `json:"auth_type,omitempty"`
	WorkloadPoolProviderPath       string `json:"workload_pool_provider_path,omitempty"`
	ServiceAccountImpersonationURL string `json:"service_account_impersonation_url,omitempty"`
}

type BigQueryCredential struct {
	ID                           *int                                  `json:"id"`
	Account_Id                   int64                                 `json:"account_id"`
	Project_Id                   int                                   `json:"project_id"`
	Type                         string                                `json:"type"`
	State                        int                                   `json:"state"`
	Threads                      int                                   `json:"threads"`
	Dataset                      string                                `json:"schema"`
	AdapterVersion               string                                `json:"adapter_version,omitempty"`
	CredentialDetails            *AdapterCredentialDetails             `json:"credential_details,omitempty"`
	UnencryptedCredentialDetails *BigQueryUnencryptedCredentialDetails `json:"unencrypted_credential_details,omitempty"`
}

// GetDataset returns the dataset/schema from either the top-level field (v0) or unencrypted_credential_details (v1)
func (c *BigQueryCredential) GetDataset() string {
	// For v1 credentials, dataset is in unencrypted_credential_details
	if c.UnencryptedCredentialDetails != nil && c.UnencryptedCredentialDetails.Schema != "" {
		return c.UnencryptedCredentialDetails.Schema
	}
	// For v0 credentials, dataset is in the top-level schema field
	return c.Dataset
}

// GetThreads returns the threads from either the top-level field (v0) or unencrypted_credential_details (v1)
func (c *BigQueryCredential) GetThreads() int {
	// For v1 credentials, threads is in unencrypted_credential_details
	if c.UnencryptedCredentialDetails != nil && c.UnencryptedCredentialDetails.Threads > 0 {
		return c.UnencryptedCredentialDetails.Threads
	}
	// For v0 credentials or if not found
	return c.Threads
}

// GetAuthType returns the auth_type from unencrypted_credential_details (v1 only)
func (c *BigQueryCredential) GetAuthType() string {
	if c.UnencryptedCredentialDetails != nil {
		return c.UnencryptedCredentialDetails.AuthType
	}
	return ""
}

// GetWorkloadPoolProviderPath returns the workload_pool_provider_path from unencrypted_credential_details (v1 only)
func (c *BigQueryCredential) GetWorkloadPoolProviderPath() string {
	if c.UnencryptedCredentialDetails != nil {
		return c.UnencryptedCredentialDetails.WorkloadPoolProviderPath
	}
	return ""
}

// GetServiceAccountImpersonationURL returns the service_account_impersonation_url from unencrypted_credential_details (v1 only)
func (c *BigQueryCredential) GetServiceAccountImpersonationURL() string {
	if c.UnencryptedCredentialDetails != nil {
		return c.UnencryptedCredentialDetails.ServiceAccountImpersonationURL
	}
	return ""
}

// BigQueryCredentialGlobConn is used for creating credentials with the new adapter format (bigquery_v1)
type BigQueryCredentialGlobConn struct {
	ID                *int                     `json:"id,omitempty"`
	AccountID         int64                    `json:"account_id"`
	ProjectID         int                      `json:"project_id"`
	Type              string                   `json:"type"`
	State             int                      `json:"state"`
	Threads           int                      `json:"threads"`
	AdapterVersion    string                   `json:"adapter_version"`
	CredentialDetails AdapterCredentialDetails `json:"credential_details"`
}

func (c *Client) GetBigQueryCredential(
	projectId int,
	credentialId int,
) (*BigQueryCredential, error) {
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

	BigQueryCredentialResponse := BigQueryCredentialResponse{}
	err = json.Unmarshal(body, &BigQueryCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &BigQueryCredentialResponse.Data, nil
}

func (c *Client) CreateBigQueryCredential(
	projectId int,
	type_ string,
	isActive bool,
	dataset string,
	numThreads int,
	adapterVersion string,
	authType string,
	workloadPoolProviderPath string,
	serviceAccountImpersonationURL string,
) (*BigQueryCredential, error) {
	var requestData []byte
	var err error

	// When adapter_version is provided, use the new adapter format with credential_details
	// This is required for connections using use_latest_adapter=true (bigquery_v1)
	if adapterVersion != "" {
		credentialDetails, err := GenerateBigQueryCredentialDetails(dataset, numThreads, authType, workloadPoolProviderPath, serviceAccountImpersonationURL)
		if err != nil {
			return nil, err
		}

		newCredential := BigQueryCredentialGlobConn{
			AccountID:         c.AccountID,
			ProjectID:         projectId,
			Type:              "adapter",
			State:             STATE_ACTIVE,
			Threads:           numThreads,
			AdapterVersion:    adapterVersion,
			CredentialDetails: credentialDetails,
		}
		requestData, err = json.Marshal(newCredential)
		if err != nil {
			return nil, err
		}
	} else {
		// Use legacy format for bigquery_v0
		newBigQueryCredential := BigQueryCredential{
			Account_Id: c.AccountID,
			Project_Id: projectId,
			Type:       type_,
			State:      STATE_ACTIVE,
			Dataset:    dataset,
			Threads:    numThreads,
		}
		requestData, err = json.Marshal(newBigQueryCredential)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/credentials/",
			c.HostURL,
			c.AccountID,
			projectId,
		),
		strings.NewReader(string(requestData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	BigQueryCredentialResponse := BigQueryCredentialResponse{}
	err = json.Unmarshal(body, &BigQueryCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &BigQueryCredentialResponse.Data, nil
}

// GenerateBigQueryCredentialDetails creates the credential_details structure for BigQuery v1 credentials
func GenerateBigQueryCredentialDetails(
	dataset string,
	numThreads int,
	authType string,
	workloadPoolProviderPath string,
	serviceAccountImpersonationURL string,
) (AdapterCredentialDetails, error) {
	// The credential_details structure for BigQuery follows the same pattern as other adapters
	fields := map[string]AdapterCredentialField{
		"schema": {
			Metadata: AdapterCredentialFieldMetadata{
				Label:       "Dataset",
				Description: "The default dataset to use for this credential",
				Field_Type:  "text",
				Encrypt:     false,
				Validation:  AdapterCredentialFieldMetadataValidation{Required: true},
			},
			Value: dataset,
		},
		"threads": {
			Metadata: AdapterCredentialFieldMetadata{
				Label:       "Threads",
				Description: "The number of threads to use",
				Field_Type:  "number",
				Encrypt:     false,
				Validation:  AdapterCredentialFieldMetadataValidation{Required: true},
			},
			Value: numThreads,
		},
	}

	if authType != "" {
		fields["auth_type"] = AdapterCredentialField{
			Metadata: AdapterCredentialFieldMetadata{
				Label:       "Authentication Method",
				Description: "The authentication method to use for this credential.",
				Field_Type:  "select",
				Encrypt:     false,
				Validation:  AdapterCredentialFieldMetadataValidation{Required: true},
				Options: []AdapterCredentialFieldMetadataOptions{
					{Label: "Service Account JSON", Value: "service-account-json"},
					{Label: "Native OAuth", Value: "oauth-secrets"},
					{Label: "Workload Identity Federation", Value: "external-oauth-wif"},
				},
			},
			Value: authType,
		}
	}

	if workloadPoolProviderPath != "" {
		fields["workload_pool_provider_path"] = AdapterCredentialField{
			Metadata: AdapterCredentialFieldMetadata{
				Label:       "Workload Pool Provider Path",
				Description: "The fully specified resource name of the workload pool provider",
				Field_Type:  "text",
				Encrypt:     false,
				Validation:  AdapterCredentialFieldMetadataValidation{Required: true},
			},
			Value: workloadPoolProviderPath,
		}
	}

	if serviceAccountImpersonationURL != "" {
		fields["service_account_impersonation_url"] = AdapterCredentialField{
			Metadata: AdapterCredentialFieldMetadata{
				Label:       "Service Account Impersonation URL",
				Description: "The URL for the service account impersonation request",
				Field_Type:  "text",
				Encrypt:     false,
				Validation:  AdapterCredentialFieldMetadataValidation{Required: false},
			},
			Value: serviceAccountImpersonationURL,
		}
	}

	return AdapterCredentialDetails{
		Fields:      fields,
		Field_Order: []string{},
	}, nil
}

func (c *Client) UpdateBigQueryCredential(
	projectId int,
	credentialId int,
	BigQueryCredential BigQueryCredential,
) (*BigQueryCredential, error) {
	BigQueryCredentialData, err := json.Marshal(BigQueryCredential)
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
		strings.NewReader(string(BigQueryCredentialData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	BigQueryCredentialResponse := BigQueryCredentialResponse{}
	err = json.Unmarshal(body, &BigQueryCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &BigQueryCredentialResponse.Data, nil
}
