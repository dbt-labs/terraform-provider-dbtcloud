package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ClickhouseCredentialResponse struct {
	Data   ClickhouseCredential `json:"data"`
	Status ResponseStatus       `json:"status"`
}

type ClickhouseUnencryptedCredentialDetails struct {
	User       string `json:"user"`
	Schema     string `json:"schema"`
	TargetName string `json:"target_name"`
	Threads    *int   `json:"threads"`
}

type ClickhouseCredential struct {
	ID                           *int                                   `json:"id"`
	Account_Id                   int64                                  `json:"account_id"`
	Project_Id                   int                                    `json:"project_id"`
	Type                         string                                 `json:"type"`
	State                        int                                    `json:"state"`
	Threads                      int                                    `json:"threads"`
	Target_Name                  string                                 `json:"target_name"`
	AdapterVersion               string                                 `json:"adapter_version,omitempty"`
	Credential_Details           AdapterCredentialDetails               `json:"credential_details"`
	UnencryptedCredentialDetails ClickhouseUnencryptedCredentialDetails `json:"unencrypted_credential_details"`
}

type ClickhouseCredentialGlobConn struct {
	ID                *int                     `json:"id"`
	AccountID         int64                    `json:"account_id"`
	ProjectID         int                      `json:"project_id"`
	Type              string                   `json:"type"`
	State             int                      `json:"state"`
	Threads           int                      `json:"threads"`
	AdapterVersion    string                   `json:"adapter_version"`
	CredentialDetails AdapterCredentialDetails `json:"credential_details"`
}

type ClickhouseCredentialGlobConnPatch struct {
	ID                int                      `json:"id"`
	CredentialDetails AdapterCredentialDetails `json:"credential_details"`
}

func (c *Client) GetClickhouseCredential(
	projectID int,
	credentialID int,
) (*ClickhouseCredential, error) {
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

	credentialResponse := ClickhouseCredentialResponse{}
	err = json.Unmarshal(body, &credentialResponse)
	if err != nil {
		return nil, err
	}

	return &credentialResponse.Data, nil
}

func (c *Client) CreateClickhouseCredential(
	projectID int,
	user string,
	password string,
	schema string,
	targetName string,
	threads *int,
) (*ClickhouseCredential, error) {

	credentialDetails, err := GenerateClickhouseCredentialDetails(
		user,
		password,
		schema,
		targetName,
		threads,
	)
	if err != nil {
		return nil, err
	}

	// The API requires the top-level "threads" property to be a real integer
	// on create (it rejects both a missing key and an explicit null)
	topLevelThreads := 4
	if threads != nil {
		topLevelThreads = *threads
	}

	newClickhouseCredential := ClickhouseCredentialGlobConn{
		AccountID:         c.AccountID,
		ProjectID:         projectID,
		Type:              "adapter",
		AdapterVersion:    "clickhouse_v0",
		State:             STATE_ACTIVE,
		Threads:           topLevelThreads,
		CredentialDetails: credentialDetails,
	}

	newClickhouseCredentialData, err := json.Marshal(newClickhouseCredential)
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
		strings.NewReader(string(newClickhouseCredentialData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	clickhouseCredentialResponse := ClickhouseCredentialResponse{}
	err = json.Unmarshal(body, &clickhouseCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &clickhouseCredentialResponse.Data, nil
}

func (c *Client) UpdateClickhouseCredential(
	projectID int,
	credentialID int,
	clickhouseCredential ClickhouseCredentialGlobConnPatch,
) (*ClickhouseCredential, error) {
	clickhouseCredentialData, err := json.Marshal(clickhouseCredential)
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
		strings.NewReader(string(clickhouseCredentialData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	clickhouseCredentialResponse := ClickhouseCredentialResponse{}
	err = json.Unmarshal(body, &clickhouseCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &clickhouseCredentialResponse.Data, nil
}

func GenerateClickhouseCredentialDetails(
	user string,
	password string,
	schema string,
	targetName string,
	threads *int,
) (AdapterCredentialDetails, error) {
	// we load the raw JSON to make it easier to update if the schema changes in the future
	defaultConfig := `{
	"fields": {
      "user": {
        "metadata": {
          "label": "User",
          "description": "The ClickHouse username.",
          "field_type": "text",
          "encrypt": false,
          "overrideable": false,
          "validation": {
            "required": true
          }
        },
        "value": ""
      },
      "password": {
        "metadata": {
          "label": "Password",
          "description": "The ClickHouse password.",
          "field_type": "text",
          "encrypt": true,
          "overrideable": false,
          "validation": {
            "required": true
          }
        },
        "value": ""
      },
      "schema": {
        "metadata": {
          "label": "Schema",
          "description": "The schema where to create the dbt models.",
          "field_type": "text",
          "encrypt": false,
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
            "required": false
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
        "value": null
      }
    }
	}
`
	var clickhouseCredentialDetailsDefault AdapterCredentialDetails
	err := json.Unmarshal([]byte(defaultConfig), &clickhouseCredentialDetailsDefault)
	if err != nil {
		return clickhouseCredentialDetailsDefault, err
	}

	var threadsValue interface{}
	if threads != nil {
		threadsValue = *threads
	}

	fieldMapping := map[string]interface{}{
		"user":        user,
		"password":    password,
		"schema":      schema,
		"target_name": targetName,
		"threads":     threadsValue,
	}

	clickhouseCredentialFields := map[string]AdapterCredentialField{}
	for key, value := range clickhouseCredentialDetailsDefault.Fields {
		value.Value = fieldMapping[key]
		clickhouseCredentialFields[key] = value
	}

	credentialDetails := AdapterCredentialDetails{
		Fields:      clickhouseCredentialFields,
		Field_Order: []string{"user", "password", "schema", "target_name", "threads"},
	}
	return credentialDetails, nil
}
