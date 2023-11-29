package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type FabricConnectionDetails struct {
	AdapterId      int                      `json:"adapter_id,omitempty"`
	AdapterDetails AdapterCredentialDetails `json:"connection_details,omitempty"`
}

type FabricConnection struct {
	BaseConnection
	Details FabricConnectionDetails `json:"details"`
}

type FabricConnectionListResponse struct {
	Data   []FabricConnection `json:"data"`
	Status ResponseStatus     `json:"status"`
}

type FabricConnectionResponse struct {
	Data   FabricConnection `json:"data"`
	Status ResponseStatus   `json:"status"`
}

func (c *Client) GetFabricConnection(connectionID, projectID string) (*FabricConnection, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%s/projects/%s/connections/%s/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			projectID,
			connectionID,
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

	connectionResponse := FabricConnectionResponse{}
	err = json.Unmarshal(body, &connectionResponse)
	if err != nil {
		return nil, err
	}

	return &connectionResponse.Data, nil
}

func (c *Client) CreateFabricConnection(
	projectID int,
	name string,
	server string,
	port int,
	database string,
	retries int,
	login_timeout int,
	query_timeout int,
) (*FabricConnection, error) {

	connectionDetails := FabricConnectionDetails{}
	adapterId, err := c.createFabricAdapter(projectID)
	if err != nil {
		return nil, err
	}

	connectionDetails.AdapterId = *adapterId

	connectionDetails.AdapterDetails = *GetFabricConnectionDetails(
		server,
		port,
		database,
		retries,
		login_timeout,
		query_timeout,
	)
	newConnection := FabricConnection{
		BaseConnection: BaseConnection{
			AccountID:             c.AccountID,
			ProjectID:             projectID,
			Name:                  name,
			Type:                  "adapter",
			PrivateLinkEndpointID: "",
			State:                 1,
		},
		Details: connectionDetails,
	}

	newConnectionData, err := json.Marshal(newConnection)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/projects/%s/connections/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			strconv.Itoa(projectID),
		),
		strings.NewReader(string(newConnectionData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	connectionResponse := FabricConnectionResponse{}
	err = json.Unmarshal(body, &connectionResponse)
	if err != nil {
		return nil, err
	}

	return &connectionResponse.Data, nil
}

func (c *Client) UpdateFabricConnection(
	connectionID, projectID string,
	connection FabricConnection,
) (*FabricConnection, error) {
	connectionData, err := json.Marshal(connection)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/projects/%s/connections/%s/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			projectID,
			connectionID,
		),
		strings.NewReader(string(connectionData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	connectionResponse := FabricConnectionResponse{}
	err = json.Unmarshal(body, &connectionResponse)
	if err != nil {
		return nil, err
	}

	return &connectionResponse.Data, nil
}

func (c *Client) createFabricAdapter(projectID int) (*int, error) {

	newAdapter := Adapter{
		ID:             nil,
		AdapterVersion: "fabric_v0",
		ProjectID:      projectID,
		AccountID:      c.AccountID,
		State:          1,
		Metadata: AdapterMetadata{
			Title:     "Fabric",
			DocsLink:  "https://docs.getdbt.com/docs/core/connect-data-platform/fabric-setup",
			ImageLink: "https://static.wikia.nocookie.net/logopedia/images/a/aa/Microsoft_Fabric_2023.svg",
		},
	}

	return createGenericAdapter(c, newAdapter, projectID)
}

func GetFabricConnectionDetails(
	server string,
	port int,
	database string,
	retries int,
	loginTimeout int,
	queryTimeout int,
) *AdapterCredentialDetails {
	validation := AdapterCredentialFieldMetadataValidation{
		Required: false,
	}
	noValidation := AdapterCredentialFieldMetadataValidation{
		Required: false,
	}

	typeMetadata := AdapterCredentialFieldMetadata{
		Label:        "Connection type",
		Description:  "",
		Field_Type:   "hidden",
		Encrypt:      false,
		Overrideable: false,
		Validation:   noValidation,
	}
	typeField := AdapterCredentialField{
		Metadata: typeMetadata,
		Value:    "fabric",
	}

	driverMetadata := AdapterCredentialFieldMetadata{
		Label:        "Driver",
		Description:  "The driver to use for this connection.",
		Field_Type:   "hidden",
		Encrypt:      false,
		Overrideable: false,
		Validation:   validation,
	}
	driverField := AdapterCredentialField{
		Metadata: driverMetadata,
		Value:    "ODBC Driver 18 for SQL Server",
	}

	serverMetadata := AdapterCredentialFieldMetadata{
		Label:        "Server",
		Description:  "The server hostname",
		Field_Type:   "text",
		Encrypt:      false,
		Overrideable: false,
		Validation:   validation,
	}
	serverField := AdapterCredentialField{
		Metadata: serverMetadata,
		Value:    server,
	}

	portMetadata := AdapterCredentialFieldMetadata{
		Label:        "Port",
		Description:  "The port to connect to for this connection.",
		Field_Type:   "number",
		Encrypt:      false,
		Overrideable: false,
		Validation:   validation,
	}
	portField := AdapterCredentialField{
		Metadata: portMetadata,
		Value:    port,
	}

	databaseMetadata := AdapterCredentialFieldMetadata{
		Label:        "Database",
		Description:  "The database to connect to for this connection.",
		Field_Type:   "text",
		Encrypt:      false,
		Overrideable: false,
		Validation:   validation,
	}
	databaseField := AdapterCredentialField{
		Metadata: databaseMetadata,
		Value:    database,
	}

	retriesMetadata := AdapterCredentialFieldMetadata{
		Label:        "Retries",
		Description:  "The number of automatic times to retry a query before failing. Defaults to 1. Queries with syntax errors will not be retried. This setting can be used to overcome intermittent network issues.",
		Field_Type:   "number",
		Encrypt:      false,
		Overrideable: false,
		Validation:   noValidation,
	}
	retriesField := AdapterCredentialField{
		Metadata: retriesMetadata,
		Value:    retries,
	}

	loginTimeoutMetadata := AdapterCredentialFieldMetadata{
		Label:        "Login timeout",
		Description:  "The number of seconds used to establish a connection before failing. Defaults to 0, which means that the timeout is disabled or uses the default system settings.",
		Field_Type:   "number",
		Encrypt:      false,
		Overrideable: false,
		Validation:   noValidation,
	}
	loginTimeoutField := AdapterCredentialField{
		Metadata: loginTimeoutMetadata,
		Value:    loginTimeout,
	}

	queryTimeoutMetadata := AdapterCredentialFieldMetadata{
		Label:        "Query timeout",
		Description:  "The number of seconds used to wait for a query before failing. Defaults to 0, which means that the timeout is disabled or uses the default system settings.",
		Field_Type:   "number",
		Encrypt:      false,
		Overrideable: false,
		Validation:   noValidation,
	}
	queryTimeoutField := AdapterCredentialField{
		Metadata: queryTimeoutMetadata,
		Value:    queryTimeout,
	}

	fieldOrder := []string{"type",
		"driver",
		"server",
		"port",
		"database",
		"retries",
		"login_timeout",
		"query_timeout"}
	fields := map[string]AdapterCredentialField{
		"type":          typeField,
		"driver":        driverField,
		"server":        serverField,
		"port":          portField,
		"database":      databaseField,
		"retries":       retriesField,
		"login_timeout": loginTimeoutField,
		"query_timeout": queryTimeoutField,
	}

	fabricCredentialDetails := AdapterCredentialDetails{
		Fields:      fields,
		Field_Order: fieldOrder,
	}

	return &fabricCredentialDetails
}
