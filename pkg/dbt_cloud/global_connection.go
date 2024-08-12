package dbt_cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// // TODO: Do we need this? Or do we just remove it?
// type GlobalConnectionType int

// // Declare enum values using iota
// const (
// 	Snowflake GlobalConnectionType = iota
// 	BigQuery
// )

type GlobalConnectionConfig interface {
	AdapterVersion() string
}

type GlobalConnectionCommon struct {
	ID                    *int64  `json:"id,omitempty"`
	Name                  *string `json:"name"`
	IsSshTunnelEnabled    *bool   `json:"is_ssh_tunnel_enabled"`
	PrivateLinkEndpointId *int64  `json:"private_link_endpoint_id"`
	OauthConfigurationId  *int64  `json:"oauth_configuration_id"`
	// OauthRedirectUri           *string `json:"oauth_redirect_uri"` //those are read-only fields, we could maybe get them as Computed but never send them
	// IsConfiguredForNativeOauth bool    `json:"is_configured_for_native_oauth"`
}

type globalConnectionPayload[T GlobalConnectionConfig] struct {
	GlobalConnectionCommon
	AccountID      int64   `json:"account_id"`
	AdapterVersion *string `json:"adapter_version,omitempty"`
	Config         T       `json:"config"`
}

type globalConnectionResponse[T GlobalConnectionConfig] struct {
	Status ResponseStatus             `json:"status"`
	Data   globalConnectionPayload[T] `json:"data"`
}

type GlobalConnectionClient[T GlobalConnectionConfig] struct{ *Client }

func NewGlobalConnectionClient[T GlobalConnectionConfig](c *Client) GlobalConnectionClient[T] {
	return GlobalConnectionClient[T]{
		c,
	}
}

func (c *GlobalConnectionClient[T]) Get(connectionID int64) (*GlobalConnectionCommon, *T, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%d/connections/%d/",
			c.HostURL,
			c.AccountID,
			connectionID,
		),
		nil,
	)

	if err != nil {
		return nil, nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, nil, err
	}

	resp := new(globalConnectionResponse[T])

	err = json.Unmarshal(body, resp)
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data.GlobalConnectionCommon, &resp.Data.Config, nil
}

func (c *GlobalConnectionClient[T]) Create(common GlobalConnectionCommon, config T) (*GlobalConnectionCommon, *T, error) {

	buffer := new(bytes.Buffer)
	enc := json.NewEncoder(buffer)

	av := config.AdapterVersion()

	payload := globalConnectionPayload[T]{
		GlobalConnectionCommon: common,
		AccountID:              int64(c.AccountID),
		AdapterVersion:         &av,
		Config:                 config,
	}

	err := enc.Encode(payload)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%d/connections/",
			c.HostURL,
			c.AccountID,
		),
		buffer,
	)
	if err != nil {
		return nil, nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, nil, err
	}

	resp := new(globalConnectionResponse[T])
	err = json.Unmarshal(body, resp)
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data.GlobalConnectionCommon, &resp.Data.Config, nil

}
func (c *GlobalConnectionClient[T]) Update(connectionID int64, common GlobalConnectionCommon, config T) (*GlobalConnectionCommon, *T, error) {

	buffer := new(bytes.Buffer)
	enc := json.NewEncoder(buffer)

	payload := globalConnectionPayload[T]{
		GlobalConnectionCommon: common,
		AccountID:              int64(c.AccountID),
		Config:                 config,
	}

	err := enc.Encode(payload)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequest(
		"PATCH",
		fmt.Sprintf(
			"%s/v3/accounts/%d/connections/%d/",
			c.HostURL,
			c.AccountID,
			connectionID,
		),
		buffer,
	)
	if err != nil {
		return nil, nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, nil, err
	}

	resp := new(globalConnectionResponse[T])
	err = json.Unmarshal(body, resp)
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data.GlobalConnectionCommon, &resp.Data.Config, nil
}

func (c *Client) DeleteGlobalConnection(connectionID int64) (string, error) {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf(
			"%s/v3/accounts/%d/connections/%d/",
			c.HostURL,
			c.AccountID,
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

// I originally put pointers everywhere and used omitempty but it prevented us from sending null values
// and differentiating between null and not sending the field at all
// TODO: rename to something else than Pointers now that those fields are not Pointers :-)
// TODO: check if we could reuse the same structure as the other one, with the nullable fields and the omitempty
type SnowflakeConfig struct {
	Account                *string `json:"account,omitempty"`
	Database               *string `json:"database,omitempty"`
	Warehouse              *string `json:"warehouse,omitempty"`
	ClientSessionKeepAlive *bool   `json:"client_session_keep_alive,omitempty"`
	Role                   *string `json:"role,omitempty"`
	AllowSso               *bool   `json:"allow_sso,omitempty"`
	OauthClientID          *string `json:"oauth_client_id,omitempty"`
	OauthClientSecret      *string `json:"oauth_client_secret,omitempty"`
}

func (SnowflakeConfig) AdapterVersion() string {
	return "snowflake_v0"
}

type BigQueryConfig struct {
	ProjectID                 *string  `json:"project_id,omitempty"`
	TimeoutSeconds            *int64   `json:"timeout_seconds,omitempty"`
	PrivateKeyID              *string  `json:"private_key_id,omitempty"`
	PrivateKey                *string  `json:"private_key,omitempty"`
	ClientEmail               *string  `json:"client_email,omitempty"`
	ClientID                  *string  `json:"client_id,omitempty"`
	AuthURI                   *string  `json:"auth_uri,omitempty"`
	TokenURI                  *string  `json:"token_uri,omitempty"`
	AuthProviderX509CertURL   *string  `json:"auth_provider_x509_cert_url,omitempty"`
	ClientX509CertURL         *string  `json:"client_x509_cert_url,omitempty"`
	Priority                  *string  `json:"priority,omitempty"`
	Retries                   *int64   `json:"retries,omitempty"`
	Location                  *string  `json:"location,omitempty"`
	MaximumBytesBilled        *int64   `json:"maximum_bytes_billed,omitempty"`
	ExecutionProject          *string  `json:"execution_project,omitempty"`
	ImpersonateServiceAccount *string  `json:"impersonate_service_account,omitempty"`
	JobRetryDeadlineSeconds   *int64   `json:"job_retry_deadline_seconds,omitempty"`
	JobCreationTimeoutSeconds *int64   `json:"job_creation_timeout_seconds,omitempty"`
	ApplicationID             *string  `json:"application_id,omitempty"`
	ApplicationSecret         *string  `json:"application_secret,omitempty"`
	GcsBucket                 *string  `json:"gcs_bucket,omitempty"`
	DataprocRegion            *string  `json:"dataproc_region,omitempty"`
	DataprocClusterName       *string  `json:"dataproc_cluster_name,omitempty"`
	Scopes                    []string `json:"scopes,omitempty"`
}

func (BigQueryConfig) AdapterVersion() string {
	return "bigquery_v0"
}
