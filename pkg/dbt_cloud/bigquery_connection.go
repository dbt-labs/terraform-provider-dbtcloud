package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type BigQueryConnectionDetails struct {
	GcpProjectId            string  `json:"project_id"`
	TimeoutSeconds          int     `json:"timeout_seconds"`
	PrivateKeyId            string  `json:"private_key_id"`
	PrivateKey              string  `json:"private_key,omitempty"`
	ClientEmail             string  `json:"client_email"`
	ClientId                string  `json:"client_id"`
	AuthUri                 string  `json:"auth_uri"`
	TokenUri                string  `json:"token_uri"`
	AuthProviderX509CertUrl string  `json:"auth_provider_x509_cert_url"`
	ClientX509CertUrl       string  `json:"client_x509_cert_url"`
	Retries                 *int    `json:"retries,omitempty"`
	Location                *string `json:"location,omitempty"`
	MaximumBytesBilled      *int    `json:"maximum_bytes_billed,omitempty"`
	GcsBucket               *string `json:"gcs_bucket,omitempty"`
	DataprocRegion          *string `json:"dataproc_region,omitempty"`
	DataprocClusterName     *string `json:"dataproc_cluster_name,omitempty"`
}

// maybe try interface here
type BigQueryConnection struct {
	ID                      *int                      `json:"id,omitempty"`
	AccountID               int                       `json:"account_id"`
	ProjectID               int                       `json:"project_id"`
	Name                    string                    `json:"name"`
	Type                    string                    `json:"type"`
	CreatedByID             *int                      `json:"created_by_id,omitempty"`
	CreatedByServiceTokenID *int                      `json:"created_by_service_token_id,omitempty"`
	State                   int                       `json:"state"`
	Created_At              *string                   `json:"created_at,omitempty"`
	Updated_At              *string                   `json:"updated_at,omitempty"`
	Details                 BigQueryConnectionDetails `json:"details"`
}

type BigQueryConnectionResponse struct {
	Data   BigQueryConnection `json:"data"`
	Status ResponseStatus     `json:"status"`
}

func (c *Client) GetBigQueryConnection(connectionID, projectID string) (*BigQueryConnection, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/accounts/%s/projects/%s/connections/%s/", c.HostURL, strconv.Itoa(c.AccountID), projectID, connectionID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	connectionResponse := BigQueryConnectionResponse{}
	err = json.Unmarshal(body, &connectionResponse)
	if err != nil {
		return nil, err
	}

	return &connectionResponse.Data, nil
}

func (c *Client) CreateBigQueryConnection(
	projectID int,
	name string,
	connectionType string,
	isActive bool,
	gcpProjectId string,
	timeoutSeconds int,
	privateKeyId string,
	privateKey string,
	clientEmail string,
	clientId string,
	authUri string,
	tokenUri string,
	authProviderX509CertUrl string,
	clientX509CertUrl string,
	retries *int,
	location *string,
	maximumBytesBilled *int,
	gcsBucket *string,
	dataprocRegion *string,
	dataprocClusterName *string) (*BigQueryConnection, error) {
	state := STATE_ACTIVE
	if !isActive {
		state = STATE_DELETED
	}

	connectionDetails := BigQueryConnectionDetails{
		GcpProjectId:            gcpProjectId,
		TimeoutSeconds:          timeoutSeconds,
		PrivateKeyId:            privateKeyId,
		PrivateKey:              privateKey,
		ClientEmail:             clientEmail,
		ClientId:                clientId,
		AuthUri:                 authUri,
		TokenUri:                tokenUri,
		AuthProviderX509CertUrl: authProviderX509CertUrl,
		ClientX509CertUrl:       clientX509CertUrl,
		Retries:                 retries,
		Location:                location,
		MaximumBytesBilled:      maximumBytesBilled,
		GcsBucket:               gcsBucket,
		DataprocRegion:          dataprocRegion,
		DataprocClusterName:     dataprocClusterName,
	}

	newConnection := BigQueryConnection{
		AccountID: c.AccountID,
		ProjectID: projectID,
		Name:      name,
		Type:      connectionType,
		State:     state,
		Details:   connectionDetails,
	}

	newConnectionData, err := json.Marshal(newConnection)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%s/projects/%s/connections/", c.HostURL, strconv.Itoa(c.AccountID), strconv.Itoa(projectID)), strings.NewReader(string(newConnectionData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	connectionResponse := BigQueryConnectionResponse{}
	err = json.Unmarshal(body, &connectionResponse)
	if err != nil {
		return nil, err
	}

	return &connectionResponse.Data, nil
}

func (c *Client) UpdateBigQueryConnection(connectionID, projectID string, connection BigQueryConnection) (*BigQueryConnection, error) {
	connectionData, err := json.Marshal(connection)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%s/projects/%s/connections/%s/", c.HostURL, strconv.Itoa(c.AccountID), projectID, connectionID), strings.NewReader(string(connectionData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	connectionResponse := BigQueryConnectionResponse{}
	err = json.Unmarshal(body, &connectionResponse)
	if err != nil {
		return nil, err
	}

	return &connectionResponse.Data, nil
}
