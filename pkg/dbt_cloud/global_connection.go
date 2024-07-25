package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/oapi-codegen/nullable"
)

// TODO: Do we need this? Or do we just remove it?
type GlobalConnectionType int

// Declare enum values using iota
const (
	Snowflake GlobalConnectionType = iota
	BigQuery
)

type GlobalConnectionInterface interface {
	GeneratePostPayload() (*strings.Reader, error)
	// UpdatePayload() (*strings.Reader, error)
	ReadAnswer([]byte) (any, error)
}

type GlobalConnectionPointerInterface interface {
	GeneratePatchPayload() (*strings.Reader, error)
}

type GlobalConnection struct {
	ID                    *int64 `json:"id,omitempty"`
	AccountID             int64  `json:"account_id"`
	AdapterVersion        string `json:"adapter_version"`
	Name                  string `json:"name"`
	IsSshTunnelEnabled    bool   `json:"is_ssh_tunnel_enabled"`
	PrivateLinkEndpointId *int64 `json:"private_link_endpoint_id"`
	OauthConfigurationId  *int64 `json:"oauth_configuration_id"`
	// OauthRedirectUri           *string `json:"oauth_redirect_uri"` //those are read-only fields, we could maybe get them as Computed but never send them
	// IsConfiguredForNativeOauth bool    `json:"is_configured_for_native_oauth"`
}

type GlobalConnectionPointers struct {
	ID                    *int64  `json:"id,omitempty"`
	AccountID             *int64  `json:"account_id,omitempty"`
	AdapterVersion        *string `json:"adapter_version,omitempty"`
	Name                  *string `json:"name,omitempty"`
	IsSshTunnelEnabled    *bool   `json:"is_ssh_tunnel_enabled,omitempty"`
	PrivateLinkEndpointId *int64  `json:"private_link_endpoint_id"` // those seem to be sent all the time when we modify a connection so I didn't add an omitempty for now
	OauthConfigurationId  *int64  `json:"oauth_configuration_id"`
}

type GlobalConnectionResponse[T GlobalConnectionInterface] struct {
	Data   T              `json:"data"`
	Status ResponseStatus `json:"status"`
}

type SnowflakeGlobalConnection struct {
	GlobalConnection
	Config SnowflakeConfig `json:"config"`
}

type SnowflakeConfig struct {
	Account                string `json:"account"`
	Database               string `json:"database"`
	Warehouse              string `json:"warehouse"`
	ClientSessionKeepAlive bool   `json:"client_session_keep_alive"`
	Role                   string `json:"role"`
	AllowSso               bool   `json:"allow_sso"`
	OauthClientID          string `json:"oauth_client_id"`
	OauthClientSecret      string `json:"oauth_client_secret"`
}

// we create a pointer version so that we can PATCH parts of the connection
type SnowflakeGlobalConnectionPointers struct {
	GlobalConnectionPointers
	Config *SnowflakeConfigPointers `json:"config,omitempty"`
}

// I originally put pointers everywhere and used omitempty but it prevented us from sending null values
// and differentiating between null and not sending the field at all
// TODO: rename to something else than Pointers now that those fields are not Pointers :-)
// TODO: check if we could reuse the same structure as the other one, with the nullable fields and the omitempty
type SnowflakeConfigPointers struct {
	Account                string                    `json:"account,omitempty"`
	Database               string                    `json:"database,omitempty"`
	Warehouse              string                    `json:"warehouse,omitempty"`
	ClientSessionKeepAlive bool                      `json:"client_session_keep_alive,omitempty"`
	Role                   nullable.Nullable[string] `json:"role,omitempty"`
	AllowSso               bool                      `json:"allow_sso,omitempty"`
	OauthClientID          string                    `json:"oauth_client_id,omitempty"`
	OauthClientSecret      string                    `json:"oauth_client_secret,omitempty"`
}

// This is how it was before
// TODO: delete when we get the pattern working
// type SnowflakeConfigPointers struct {
// 	Account                *string `json:"account,omitempty"`
// 	Database               *string `json:"database,omitempty"`
// 	Warehouse              *string `json:"warehouse,omitempty"`
// 	ClientSessionKeepAlive *bool   `json:"client_session_keep_alive,omitempty"`
// 	Role                   *string `json:"role,omitempty"`
// 	AllowSso               *bool   `json:"allow_sso,omitempty"`
// 	OauthClientID          *string `json:"oauth_client_id,omitempty"`
// 	OauthClientSecret      *string `json:"oauth_client_secret,omitempty"`
// }

type SnowflakeGlobalConnectionResponse struct {
	Data   SnowflakeGlobalConnection `json:"data"`
	Status ResponseStatus            `json:"status"`
}

func (gc SnowflakeGlobalConnection) GeneratePostPayload() (*strings.Reader, error) {
	payload, err := json.Marshal(gc)

	if err != nil {
		return nil, err
	}

	return strings.NewReader(string(payload)), nil
}

func (gc SnowflakeGlobalConnection) ReadAnswer(
	body []byte,
) (any, error) {
	connectionResponse := SnowflakeGlobalConnectionResponse{}
	err := json.Unmarshal(body, &connectionResponse)
	if err != nil {
		return nil, err
	}
	return connectionResponse.Data, nil
}

func (gcp SnowflakeGlobalConnectionPointers) GeneratePatchPayload() (*strings.Reader, error) {
	payload, err := json.Marshal(gcp)

	if err != nil {
		return nil, err
	}

	return strings.NewReader(string(payload)), nil
}

type BigQueryGlobalConnection struct {
	GlobalConnection
	Config BigQueryConfig `json:"config"`
}

type BigQueryConfig struct {
	ProjectID                 string   `json:"project_id"`
	TimeoutSeconds            int64    `json:"timeout_seconds"`
	PrivateKeyID              string   `json:"private_key_id"`
	PrivateKey                string   `json:"private_key"`
	ClientEmail               string   `json:"client_email"`
	ClientID                  string   `json:"client_id"`
	AuthURI                   string   `json:"auth_uri"`
	TokenURI                  string   `json:"token_uri"`
	AuthProviderX509CertURL   string   `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL         string   `json:"client_x509_cert_url"`
	Priority                  string   `json:"priority"`
	Retries                   int64    `json:"retries"`
	Location                  string   `json:"location"`
	MaximumBytesBilled        int64    `json:"maximum_bytes_billed"`
	ExecutionProject          string   `json:"execution_project"`
	ImpersonateServiceAccount string   `json:"impersonate_service_account"`
	JobRetryDeadlineSeconds   int64    `json:"job_retry_deadline_seconds"`
	JobCreationTimeoutSeconds int64    `json:"job_creation_timeout_seconds"`
	ApplicationID             string   `json:"application_id"`
	ApplicationSecret         string   `json:"application_secret"`
	GcsBucket                 string   `json:"gcs_bucket"`
	DataprocRegion            string   `json:"dataproc_region"`
	DataprocClusterName       string   `json:"dataproc_cluster_name"`
	Scopes                    []string `json:"scopes"`
}

type BigQueryGlobalConnectionResponse struct {
	Data   BigQueryGlobalConnection `json:"data"`
	Status ResponseStatus           `json:"status"`
}

func (gc BigQueryGlobalConnection) GetCreatePayload() (*strings.Reader, error) {
	payload, err := json.Marshal(gc)

	if err != nil {
		return nil, err
	}

	return strings.NewReader(string(payload)), nil
}

func GetTypedGlobalConnection[T GlobalConnectionInterface](
	c *Client,
	connectionID int64,
) (*T, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%s/connections/%d/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
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

	var conn T
	data, err := conn.ReadAnswer(body)
	if err != nil {
		return nil, err
	}

	dataTyped := data.(T)
	return &dataTyped, nil
}

func (c *Client) GetSnowflakeGlobalConnection(
	connectionID int64,
) (*SnowflakeGlobalConnection, error) {

	return GetTypedGlobalConnection[SnowflakeGlobalConnection](c, connectionID)
}

func CreateTypedGlobalConnection[T GlobalConnectionInterface](
	c *Client,
	connection T,
) (*T, error) {

	payload, err := connection.GeneratePostPayload()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/connections/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
		),
		payload,
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	data, err := connection.ReadAnswer(body)
	if err != nil {
		return nil, err
	}

	// I could not find a way to do this without doing type inference on any
	dataTyped := data.(T)
	return &dataTyped, nil

}

func (c *Client) CreateSnowflakeGlobalConnection(
	connection SnowflakeGlobalConnection,
) (*SnowflakeGlobalConnection, error) {
	return CreateTypedGlobalConnection(c, connection)
}

func UpdateTypedGlobalConnection[T GlobalConnectionInterface, PT GlobalConnectionPointerInterface](
	c *Client,
	connectionID int64,
	connection PT,
) (*T, error) {

	payload, err := connection.GeneratePatchPayload()
	// connectionData, err := json.Marshal(connection)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"PATCH",
		fmt.Sprintf(
			"%s/v3/accounts/%d/connections/%d/",
			c.HostURL,
			c.AccountID,
			connectionID,
		),
		payload,
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var conn T
	data, err := conn.ReadAnswer(body)
	if err != nil {
		return nil, err
	}

	// I could not find a way to do this without doing type inference on any
	dataTyped := data.(T)
	return &dataTyped, nil
}

func (c *Client) UpdateSnowflakeGlobalConnection(
	connectionID int64,
	connectionPointers SnowflakeGlobalConnectionPointers,
) (*SnowflakeGlobalConnection, error) {
	return UpdateTypedGlobalConnection[SnowflakeGlobalConnection](
		c,
		connectionID,
		connectionPointers,
	)
}

func (c *Client) DeleteGlobalConnection(connectionID int64) (string, error) {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf(
			"%s/v3/accounts/%s/connections/%d/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			connectionID,
		),
		nil,
	)
	if err != nil {
		return "", err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return "", err
	}

	return "", nil
}
