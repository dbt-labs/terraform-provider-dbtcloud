package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type SalesforceCredentialListResponse struct {
	Data   []SalesforceCredential `json:"data"`
	Status ResponseStatus         `json:"status"`
}

type SalesforceCredentialResponse struct {
	Data   SalesforceCredential `json:"data"`
	Status ResponseStatus       `json:"status"`
}

type SalesforceUnencryptedCredentialDetails struct {
	Username   string `json:"username"`
	ClientID   string `json:"client_id"`
	PrivateKey string `json:"private_key,omitempty"`
	TargetName string `json:"target_name"`
	Threads    int    `json:"threads"`
}

type SalesforceCredential struct {
	ID                           *int                                   `json:"id"`
	Account_Id                   int                                    `json:"account_id"`
	Project_Id                   int                                    `json:"project_id"`
	Type                         string                                 `json:"type"`
	State                        int                                    `json:"state"`
	Threads                      int                                    `json:"threads"`
	Target_Name                  string                                 `json:"target_name"`
	AdapterVersion               string                                 `json:"adapter_version,omitempty"`
	Credential_Details           AdapterCredentialDetails               `json:"credential_details"`
	UnencryptedCredentialDetails SalesforceUnencryptedCredentialDetails `json:"unencrypted_credential_details"`
}

type SalesforceCredentialGlobConn struct {
	ID                *int                     `json:"id"`
	AccountID         int                      `json:"account_id"`
	ProjectID         int                      `json:"project_id"`
	Type              string                   `json:"type"`
	State             int                      `json:"state"`
	Threads           int                      `json:"threads"`
	AdapterVersion    string                   `json:"adapter_version"`
	CredentialDetails AdapterCredentialDetails `json:"credential_details"`
}

type SalesforceCredentialGlobConnPatch struct {
	ID                int                      `json:"id"`
	CredentialDetails AdapterCredentialDetails `json:"credential_details"`
}

func (c *Client) GetSalesforceCredential(
	projectID int,
	credentialID int,
) (*SalesforceCredential, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/credentials/%d/?include_related=[adapter]",
			c.HostURL,
			c.AccountID,
			projectID,
			credentialID,
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

	credentialResponse := SalesforceCredentialResponse{}
	err = json.Unmarshal(body, &credentialResponse)
	if err != nil {
		return nil, err
	}

	return &credentialResponse.Data, nil
}

func (c *Client) CreateSalesforceCredential(
	projectID int,
	username string,
	clientID string,
	privateKey string,
	targetName string,
	threads int,
) (*SalesforceCredential, error) {

	credentialDetails, err := GenerateSalesforceCredentialDetails(
		username,
		clientID,
		privateKey,
		targetName,
		threads,
	)
	if err != nil {
		return nil, err
	}

	newSalesforceCredential := SalesforceCredentialGlobConn{
		AccountID:         c.AccountID,
		ProjectID:         projectID,
		Type:              "adapter",
		AdapterVersion:    "salesforce_v0",
		State:             STATE_ACTIVE,
		Threads:           threads,
		CredentialDetails: credentialDetails,
	}

	newSalesforceCredentialData, err := json.Marshal(newSalesforceCredential)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/credentials/",
			c.HostURL,
			c.AccountID,
			projectID,
		),
		strings.NewReader(string(newSalesforceCredentialData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	salesforceCredentialResponse := SalesforceCredentialResponse{}
	err = json.Unmarshal(body, &salesforceCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &salesforceCredentialResponse.Data, nil
}

func (c *Client) UpdateSalesforceCredential(
	projectID int,
	credentialID int,
	salesforceCredential SalesforceCredentialGlobConnPatch,
) (*SalesforceCredential, error) {
	salesforceCredentialData, err := json.Marshal(salesforceCredential)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"PATCH",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/credentials/%d/",
			c.HostURL,
			c.AccountID,
			projectID,
			credentialID,
		),
		strings.NewReader(string(salesforceCredentialData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	salesforceCredentialResponse := SalesforceCredentialResponse{}
	err = json.Unmarshal(body, &salesforceCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &salesforceCredentialResponse.Data, nil
}

func GenerateSalesforceCredentialDetails(
	username string,
	clientID string,
	privateKey string,
	targetName string,
	threads int,
) (AdapterCredentialDetails, error) {
	// Based on the Salesforce credential schema provided
	defaultConfig := `{
	"fields": {
      "username": {
        "metadata": {
          "label": "Username",
          "description": "The Salesforce username for OAuth JWT bearer flow authentication",
          "field_type": "text",
          "encrypt": false,
          "overrideable": false,
          "validation": {
            "required": true
          }
        },
        "value": ""
      },
      "client_id": {
        "metadata": {
          "label": "Client ID",
          "description": "The OAuth connected app client/consumer ID",
          "field_type": "text",
          "encrypt": true,
          "overrideable": false,
          "validation": {
            "required": true
          }
        },
        "value": ""
      },
      "private_key": {
        "metadata": {
          "label": "Private Key",
          "description": "The private key for JWT bearer flow authentication",
          "field_type": "textarea",
          "encrypt": true,
          "overrideable": false,
          "validation": {
            "required": true
          }
        },
        "value": ""
      },
      "target_name": {
        "metadata": {
          "label": "Target name",
          "description": "",
          "field_type": "text",
          "encrypt": false,
          "overrideable": false,
          "validation": {
            "required": true
          }
        },
        "value": ""
      },
      "threads": {
        "metadata": {
          "label": "Threads",
          "description": "The number of threads to use for dbt operations.",
          "field_type": "number",
          "encrypt": false,
          "overrideable": false,
          "validation": {
            "required": false
          }
        },
        "value": 6
      }
    }
	}
`
	// we load the raw JSON to make it easier to update if the schema changes in the future
	var salesforceCredentialDetailsDefault AdapterCredentialDetails
	err := json.Unmarshal([]byte(defaultConfig), &salesforceCredentialDetailsDefault)
	if err != nil {
		return salesforceCredentialDetailsDefault, err
	}

	fieldMapping := map[string]interface{}{
		"username":    username,
		"client_id":   clientID,
		"private_key": privateKey,
		"target_name": targetName,
		"threads":     threads,
	}

	salesforceCredentialFields := map[string]AdapterCredentialField{}
	for key, value := range salesforceCredentialDetailsDefault.Fields {
		value.Value = fieldMapping[key]
		salesforceCredentialFields[key] = value
	}

	credentialDetails := AdapterCredentialDetails{
		Fields:      salesforceCredentialFields,
		Field_Order: []string{"username", "client_id", "private_key", "target_name", "threads"},
	}
	return credentialDetails, nil
}
