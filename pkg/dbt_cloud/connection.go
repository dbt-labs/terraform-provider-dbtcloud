package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type ConnectionDetails struct {
	Account                string `json:"account,omitempty"`
	Database               string `json:"database,omitempty"`
	DBName                 string `json:"dbname,omitempty"`
	Warehouse              string `json:"warehouse,omitempty"`
	AllowSSO               *bool  `json:"allow_sso,omitempty"`
	ClientSessionKeepAlive *bool  `json:"client_session_keep_alive,omitempty"`
	Role                   string `json:"role,omitempty"`
	OAuthClientID          string `json:"oauth_client_id,omitempty"`
	OAuthClientSecret      string `json:"oauth_client_secret,omitempty"`
	Host                   string `json:"hostname,omitempty"`
	Port                   int    `json:"port,omitempty"`
	TunnelEnabled          *bool  `json:"tunnel_enabled,omitempty"`
}

type Connection struct {
	ID                      *int              `json:"id,omitempty"`
	AccountID               int               `json:"account_id"`
	ProjectID               int               `json:"project_id"`
	Name                    string            `json:"name"`
	Type                    string            `json:"type"`
	CreatedByID             *int              `json:"created_by_id,omitempty"`
	CreatedByServiceTokenID *int              `json:"created_by_service_token_id,omitempty"`
	State                   int               `json:"state"`
	Created_At              *string           `json:"created_at,omitempty"`
	Updated_At              *string           `json:"updated_at,omitempty"`
	Details                 ConnectionDetails `json:"details"`
}

type ConnectionListResponse struct {
	Data   []Connection   `json:"data"`
	Status ResponseStatus `json:"status"`
}

type ConnectionResponse struct {
	Data   Connection     `json:"data"`
	Status ResponseStatus `json:"status"`
}

func (c *Client) GetConnection(connectionID, projectID string) (*Connection, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/accounts/%s/projects/%s/connections/%s/", c.HostURL, strconv.Itoa(c.AccountID), projectID, connectionID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	connectionResponse := ConnectionResponse{}
	err = json.Unmarshal(body, &connectionResponse)
	if err != nil {
		return nil, err
	}

	return &connectionResponse.Data, nil
}

func (c *Client) CreateConnection(projectID int, name string, connectionType string, isActive bool, account string, database string, warehouse string, role string, allowSSO *bool, clientSessionKeepAlive *bool, oAuthClientID string, oAuthClientSecret string, hostName string, port int, tunnelEnabled *bool) (*Connection, error) {
	state := STATE_ACTIVE
	if !isActive {
		state = STATE_DELETED
	}

	connectionDetails := ConnectionDetails{
		Account:           account,
		Warehouse:         warehouse,
		Role:              role,
		OAuthClientID:     oAuthClientID,
		OAuthClientSecret: oAuthClientSecret,
		Host:              hostName,
		Port:              port,
	}
	if connectionType == "snowflake" {
		connectionDetails.Database = database
		connectionDetails.AllowSSO = allowSSO
		connectionDetails.ClientSessionKeepAlive = clientSessionKeepAlive
	} else if connectionType == "redshift" {
		connectionDetails.TunnelEnabled = tunnelEnabled
		connectionDetails.DBName = database
	} else {
		connectionDetails.TunnelEnabled = tunnelEnabled
		connectionDetails.DBName = database
	}

	newConnection := Connection{
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

	connectionResponse := ConnectionResponse{}
	err = json.Unmarshal(body, &connectionResponse)
	if err != nil {
		return nil, err
	}

	if (oAuthClientID != "") && (oAuthClientSecret != "") {
		connectionResponse.Data.Details.OAuthClientID = oAuthClientID
		connectionResponse.Data.Details.OAuthClientSecret = oAuthClientSecret
	}

	return &connectionResponse.Data, nil
}

func (c *Client) UpdateConnection(connectionID, projectID string, connection Connection) (*Connection, error) {
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

	connectionResponse := ConnectionResponse{}
	err = json.Unmarshal(body, &connectionResponse)
	if err != nil {
		return nil, err
	}

	return &connectionResponse.Data, nil
}

func (c *Client) DeleteConnection(connectionID, projectID string) (string, error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v3/accounts/%s/projects/%s/connections/%s/", c.HostURL, strconv.Itoa(c.AccountID), projectID, connectionID), nil)
	if err != nil {
		return "", err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return "", err
	}

	return "", err
}
