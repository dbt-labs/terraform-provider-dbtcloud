package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type SparkCredentialListResponse struct {
	Data   []SparkCredential `json:"data"`
	Status ResponseStatus    `json:"status"`
}

type SparkCredentialResponse struct {
	Data   SparkCredential `json:"data"`
	Status ResponseStatus  `json:"status"`
}

type SparkUnencryptedCredentialDetails struct {
	Schema     string `json:"schema"`
	TargetName string `json:"target_name"`
	Threads    int    `json:"threads"`
	Token      string `json:"token,omitempty"`
}

type SparkCredential struct {
	ID                           *int                              `json:"id"`
	Account_Id                   int                               `json:"account_id"`
	Project_Id                   int                               `json:"project_id"`
	Type                         string                            `json:"type"`
	State                        int                               `json:"state"`
	Threads                      int                               `json:"threads"`
	Target_Name                  string                            `json:"target_name"`
	AdapterVersion               string                            `json:"adapter_version,omitempty"`
	Credential_Details           AdapterCredentialDetails          `json:"credential_details"`
	UnencryptedCredentialDetails SparkUnencryptedCredentialDetails `json:"unencrypted_credential_details"`
}

type SparkCredentialGlobConn struct {
	ID                *int                     `json:"id"`
	AccountID         int                      `json:"account_id"`
	ProjectID         int                      `json:"project_id"`
	Type              string                   `json:"type"`
	State             int                      `json:"state"`
	Threads           int                      `json:"threads"`
	AdapterVersion    string                   `json:"adapter_version"`
	CredentialDetails AdapterCredentialDetails `json:"credential_details"`
}

type SparkCredentialGLobConnPatch struct {
	ID                int                      `json:"id"`
	CredentialDetails AdapterCredentialDetails `json:"credential_details"`
}

func (c *Client) GetSparkCredential(
	projectId int,
	credentialId int,
) (*SparkCredential, error) {
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

	credentialResponse := SparkCredentialResponse{}
	err = json.Unmarshal(body, &credentialResponse)
	if err != nil {
		return nil, err
	}

	return &credentialResponse.Data, nil
}

func (c *Client) CreateSparkCredential(
	projectId int,
	token string,
	schema string,
	targetName string,
) (*SparkCredential, error) {

	credentialDetails, err := GenerateSparkCredentialDetails(
		token,
		schema,
		targetName,
	)
	if err != nil {
		return nil, err
	}

	newSparkCredential := SparkCredentialGlobConn{
		AccountID:         c.AccountID,
		ProjectID:         projectId,
		Type:              "adapter",
		AdapterVersion:    "apache_spark_v0",
		State:             STATE_ACTIVE,
		Threads:           NUM_THREADS_CREDENTIAL,
		CredentialDetails: credentialDetails,
	}

	newSparkCredentialData, err := json.Marshal(newSparkCredential)
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
		strings.NewReader(string(newSparkCredentialData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	sparkCredentialResponse := SparkCredentialResponse{}
	err = json.Unmarshal(body, &sparkCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &sparkCredentialResponse.Data, nil
}

func (c *Client) UpdateSparkCredentialGlobConn(
	projectId int,
	credentialId int,
	sparkCredential SparkCredentialGLobConnPatch,
) (*SparkCredential, error) {
	sparkCredentialData, err := json.Marshal(sparkCredential)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"PATCH",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/credentials/%d/",
			c.HostURL,
			c.AccountID,
			projectId,
			credentialId,
		),
		strings.NewReader(string(sparkCredentialData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	sparkCredentialResponse := SparkCredentialResponse{}
	err = json.Unmarshal(body, &sparkCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &sparkCredentialResponse.Data, nil
}

func GenerateSparkCredentialDetails(
	token string,
	schema string,
	targetName string,
) (AdapterCredentialDetails, error) {
	// the default config is taken from the calls made to the API for Spark credentials
	// Note: Spark credentials do NOT include auth_type field (unlike Databricks)
	defaultConfig := `{
	"fields": {
      "token": {
        "metadata": {
          "label": "Token",
          "description": "Personalized user token.",
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
          "description": "User schema.",
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
          "label": "Target Name",
          "description": "",
          "field_type": "text",
          "encrypt": false,
          "overrideable": false,
          "validation": {
            "required": false
          }
        },
        "value": ""
      }
    }
	}
`
	// we load the raw JSON to make it easier to update if the schema changes in the future
	var sparkCredentialDetailsDefault AdapterCredentialDetails
	err := json.Unmarshal([]byte(defaultConfig), &sparkCredentialDetailsDefault)
	if err != nil {
		return sparkCredentialDetailsDefault, err
	}

	fieldMapping := map[string]interface{}{
		"token":       token,
		"schema":      schema,
		"target_name": targetName,
	}

	sparkCredentialFields := map[string]AdapterCredentialField{}
	for key, value := range sparkCredentialDetailsDefault.Fields {
		value.Value = fieldMapping[key]
		sparkCredentialFields[key] = value
	}

	credentialDetails := AdapterCredentialDetails{
		Fields:      sparkCredentialFields,
		Field_Order: []string{},
	}
	return credentialDetails, nil
}
