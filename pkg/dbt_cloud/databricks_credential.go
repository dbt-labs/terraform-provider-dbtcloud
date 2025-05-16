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
	AdapterVersion               string                                 `json:"adapter_version,omitempty"`
	Credential_Details           AdapterCredentialDetails               `json:"credential_details"`
	UnencryptedCredentialDetails DatabricksUnencryptedCredentialDetails `json:"unencrypted_credential_details"`
}

type DatabricksCredentialGlobConn struct {
	ID                *int                     `json:"id"`
	AccountID         int                      `json:"account_id"`
	ProjectID         int                      `json:"project_id"`
	Type              string                   `json:"type"`
	State             int                      `json:"state"`
	Threads           int                      `json:"threads"`
	AdapterVersion    string                   `json:"adapter_version"`
	CredentialDetails AdapterCredentialDetails `json:"credential_details"`
}

type DatabricksCredentialGLobConnPatch struct {
	ID                int                      `json:"id"`
	CredentialDetails AdapterCredentialDetails `json:"credential_details"`
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
	token string,
	schema string,
	targetName string,
	catalog string,

) (*DatabricksCredential, error) {

	credentialDetails, err := GenerateDatabricksCredentialDetails(
		token,
		schema,
		targetName,
		catalog,
	)
	if err != nil {
		return nil, err
	}

	newDatabricksCredential := DatabricksCredentialGlobConn{
		AccountID:         c.AccountID,
		ProjectID:         projectId,
		Type:              "adapter",
		AdapterVersion:    "databricks_v0",
		State:             STATE_ACTIVE,
		Threads:           NUM_THREADS_CREDENTIAL,
		CredentialDetails: credentialDetails,
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

func (c *Client) UpdateDatabricksCredentialGlobConn(
	projectId int,
	credentialId int,
	databricksCredential DatabricksCredentialGLobConnPatch,
) (*DatabricksCredential, error) {
	databricksCredentialData, err := json.Marshal(databricksCredential)
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

func GenerateDatabricksCredentialDetails(
	token string,
	schema string,
	targetName string,
	catalog string,

) (AdapterCredentialDetails, error) {
	// the default config is taken from  the calls made to the API
	// we just remove all the different values and set them to ""
	defaultConfig := `{
	"fields": {
      "auth_type": {
        "metadata": {
          "label": "Auth method",
          "description": "",
          "field_type": "select",
          "encrypt": false,
          "overrideable": false,
          "is_searchable": false,
          "options": [
            {
              "label": "Token",
              "value": "token"
            },
            {
              "label": "OAuth",
              "value": "oauth"
            }
          ],
          "validation": {
            "required": true
          }
        },
        "value": "token"
      },
      "token": {
        "metadata": {
          "label": "Token",
          "description": "Personalized user token.",
          "field_type": "text",
          "encrypt": true,
          "depends_on": {
            "auth_type": [
              "token"
            ]
          },
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
      },
      "catalog": {
        "metadata": {
          "label": "Catalog",
          "description": "Catalog name if Unity Catalog is enabled in your Databricks workspace.  Only available in dbt version 1.1 and later.",
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
	var databricksCredentialDetailsDefault AdapterCredentialDetails
	err := json.Unmarshal([]byte(defaultConfig), &databricksCredentialDetailsDefault)
	if err != nil {
		return databricksCredentialDetailsDefault, err
	}

	fieldMapping := map[string]interface{}{
		"token":       token,
		"schema":      schema,
		"target_name": targetName,
		"catalog":     catalog,
		"auth_type":   "token",
	}

	databricksCredentialFields := map[string]AdapterCredentialField{}
	for key, value := range databricksCredentialDetailsDefault.Fields {
		value.Value = fieldMapping[key]
		databricksCredentialFields[key] = value
	}

	credentialDetails := AdapterCredentialDetails{
		Fields:      databricksCredentialFields,
		Field_Order: []string{},
	}
	return credentialDetails, nil
}
