package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type SynapseCredentialListResponse struct {
	Data   []SynapseCredential `json:"data"`
	Status ResponseStatus      `json:"status"`
}

type SynapseCredentialResponse struct {
	Data   SynapseCredential `json:"data"`
	Status ResponseStatus    `json:"status"`
}

type SynapseUnencryptedCredentialDetails struct {
	Authentication      string `json:"authentication"`
	User                string `json:"user"`
	ClientId            string `json:"client_id"`
	Schema              string `json:"schema"`
	SchemaAuthorization string `json:"schema_authorization"`
	TenantId            string `json:"tenant_id"`
}

type SynapseCredential struct {
	ID                           *int                                `json:"id"`
	Account_Id                   int                                 `json:"account_id"`
	Project_Id                   int                                 `json:"project_id"`
	Type                         string                              `json:"type"`
	State                        int                                 `json:"state"`
	Threads                      int                                 `json:"threads"`
	AdapterVersion               string                              `json:"adapter_version"`
	CredentialDetails            AdapterCredentialDetails            `json:"credential_details"`
	UnencryptedCredentialDetails SynapseUnencryptedCredentialDetails `json:"unencrypted_credential_details"`
}

func (c *Client) GetSynapseCredential(
	projectId int,
	credentialId int,
) (*SynapseCredential, error) {
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

	credentialResponse := SynapseCredentialResponse{}
	err = json.Unmarshal(body, &credentialResponse)
	if err != nil {
		return nil, err
	}

	return &credentialResponse.Data, nil
}

func (c *Client) CreateSynapseCredential(
	projectId int,
	authentication string,
	user string,
	password string,
	tenantId string,
	clientId string,
	clientSecret string,
	schema string,
	schemaAuthorization string,
) (*SynapseCredential, error) {

	credentialDetails, err := GenerateSynapseCredentialDetails(
		authentication,
		user,
		password,
		tenantId,
		clientId,
		clientSecret,
		schema,
		schemaAuthorization,
	)
	if err != nil {
		return nil, err
	}

	newSynapseCredential := SynapseCredential{
		Account_Id:        c.AccountID,
		Project_Id:        projectId,
		Type:              "adapter",
		State:             STATE_ACTIVE,
		Threads:           NUM_THREADS_CREDENTIAL,
		AdapterVersion:    "synapse_v0",
		CredentialDetails: credentialDetails,
	}

	newSynapseCredentialData, err := json.Marshal(newSynapseCredential)
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
		strings.NewReader(string(newSynapseCredentialData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	synapseCredentialResponse := SynapseCredentialResponse{}
	err = json.Unmarshal(body, &synapseCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &synapseCredentialResponse.Data, nil
}

func GenerateSynapseCredentialDetails(
	authentication string,
	user string,
	password string,
	tenantId string,
	clientId string,
	clientSecret string,
	schema string,
	schemaAuthorization string,
) (AdapterCredentialDetails, error) {
	// the default config is taken from  the calls made to the API
	// we just remove all the different values and set them to ""
	defaultConfig := `{
	"fields": {
		"authentication": {
		"metadata": {
			"label": "Authentication",
			"description": "",
			"field_type": "select",
			"encrypt": false,
			"overrideable": false,
			"options": [
			{
				"label": "SQL",
				"value": "sql"
			},
			{
				"label": "Active Directory Password",
				"value": "ActiveDirectoryPassword"
			},
			{
				"label": "Service Principal",
				"value": "ServicePrincipal"
			}
			],
			"validation": {
			"required": true
			}
		},
		"value": ""
		},
		"user": {
		"metadata": {
			"label": "User",
			"description": "The username of the Synapse account to connect to.",
			"field_type": "text",
			"encrypt": false,
			"depends_on": {
			"authentication": [
				"sql",
				"ActiveDirectoryPassword"
			]
			},
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
			"description": "The password for the account to connect to.",
			"field_type": "text",
			"encrypt": true,
			"depends_on": {
			"authentication": [
				"sql",
				"ActiveDirectoryPassword"
			]
			},
			"overrideable": false,
			"validation": {
			"required": true
			}
		},
		"value": ""
		},
		"tenant_id": {
		"metadata": {
			"label": "Tenant ID",
			"description": "The tenant ID of the Azure Active Directory instance. This is only used when connecting to Azure SQL with a service principal.",
			"field_type": "text",
			"encrypt": false,
			"depends_on": {
			"authentication": [
				"ServicePrincipal"
			]
			},
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
			"description": "The client ID of the Azure Active Directory service principal. This is only used when connecting to Azure SQL with an AAD service principal.",
			"field_type": "text",
			"encrypt": false,
			"depends_on": {
			"authentication": [
				"ServicePrincipal"
			]
			},
			"overrideable": false,
			"validation": {
			"required": true
			}
		},
		"value": ""
		},
		"client_secret": {
		"metadata": {
			"label": "Client secret",
			"description": "The client secret of the Azure Active Directory service principal. This is only used when connecting to Azure SQL with an AAD service principal.",
			"field_type": "text",
			"encrypt": true,
			"depends_on": {
			"authentication": [
				"ServicePrincipal"
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
		"schema_authorization": {
		"metadata": {
			"label": "Schema authorization",
			"description": "Optionally set this to the principal who should own the schemas created by dbt.",
			"field_type": "text",
			"encrypt": false,
			"overrideable": false,
			"validation": {
			"required": false
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
		"threads": {
		"metadata": {
			"label": "Threads",
			"description": "The number of threads to use for dbt operations.",
			"field_type": "number",
			"encrypt": false,
			"overrideable": false,
			"validation": {
			"required": true
			}
		},
		"value": 6
		}
	}
	}
`
	// we load the raw JSON to make it easier to update if the schema changes in the future
	var synapseCredentialDetailsDefault AdapterCredentialDetails
	err := json.Unmarshal([]byte(defaultConfig), &synapseCredentialDetailsDefault)
	if err != nil {
		return synapseCredentialDetailsDefault, err
	}

	fieldMapping := map[string]interface{}{
		"authentication":       authentication,
		"user":                 user,
		"password":             password,
		"tenant_id":            tenantId,
		"client_id":            clientId,
		"client_secret":        clientSecret,
		"schema":               schema,
		"schema_authorization": schemaAuthorization,
		"target_name":          "default",
		"threads":              NUM_THREADS_CREDENTIAL,
	}

	synapseCredentialFields := map[string]AdapterCredentialField{}
	for key, value := range synapseCredentialDetailsDefault.Fields {
		value.Value = fieldMapping[key]
		synapseCredentialFields[key] = value
	}

	credentialDetails := AdapterCredentialDetails{
		Fields:      synapseCredentialFields,
		Field_Order: []string{},
	}
	return credentialDetails, nil
}

func (c *Client) UpdateSynapseCredential(
	projectId int,
	credentialId int,
	synapseCredential SynapseCredential,
) (*SynapseCredential, error) {
	synapseCredentialData, err := json.Marshal(synapseCredential)
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
		strings.NewReader(string(synapseCredentialData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	synapseCredentialResponse := SynapseCredentialResponse{}
	err = json.Unmarshal(body, &synapseCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &synapseCredentialResponse.Data, nil
}
