package dbt_cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/oapi-codegen/nullable"
)

type GlobalConnectionConfig interface {
	AdapterVersion() string
}

// TODO: Could be improved in the future, maybe creating a client with empty Config
// For now, I couldn't use it as the  AdapterVersion is not returned in the GET response
// To be revisited when we handle different versions for the same adapter
type GlobalConnectionAdapter struct {
	Data struct {
		ID             int64  `json:"id"`
		AdapterVersion string `json:"adapter_version"`
	} `json:"data"`
}

func (c *Client) GetGlobalConnectionAdapter(connectionID int64) (*GlobalConnectionAdapter, error) {
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
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	connectionResponse := GlobalConnectionAdapter{}
	err = json.Unmarshal(body, &connectionResponse)
	if err != nil {
		return nil, err
	}

	return &connectionResponse, nil
}

type GlobalConnectionCommon struct {
	ID                    *int64                    `json:"id,omitempty"`
	Name                  *string                   `json:"name,omitempty"`
	IsSshTunnelEnabled    *bool                     `json:"is_ssh_tunnel_enabled,omitempty"`
	PrivateLinkEndpointId nullable.Nullable[string] `json:"private_link_endpoint_id,omitempty"`
	OauthConfigurationId  nullable.Nullable[int64]  `json:"oauth_configuration_id,omitempty"`
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

	body, err := c.doRequestWithRetry(req)
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

func (c *GlobalConnectionClient[T]) Create(
	common GlobalConnectionCommon,
	config T,
) (*GlobalConnectionCommon, *T, error) {

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

	body, err := c.doRequestWithRetry(req)
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

func (c *GlobalConnectionClient[T]) Update(
	connectionID int64,
	common GlobalConnectionCommon,
	config T,
) (*GlobalConnectionCommon, *T, error) {

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

	body, err := c.doRequestWithRetry(req)
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

	_, err = c.doRequestWithRetry(req)
	if err != nil {
		return "", err
	}

	return "", nil
}

// To manage SSH tunnel connections

type globalConnectionEncryptionByConnectionResponse struct {
	Status ResponseStatus                      `json:"status"`
	Data   []GlobalConnectionEncryptionPayload `json:"data"`
}

type globalConnectionEncryptionResponse struct {
	Status ResponseStatus                    `json:"status"`
	Data   GlobalConnectionEncryptionPayload `json:"data"`
}

type GlobalConnectionEncryptionPayload struct {
	ID           *int64 `json:"id"`
	AccountID    int64  `json:"account_id"`
	ConnectionID int64  `json:"connection_id"`
	Username     string `json:"username,omitempty"`
	Port         int64  `json:"port,omitempty"`
	HostName     string `json:"hostname,omitempty"`
	PublicKey    string `json:"public_key,omitempty"`
	State        int64  `json:"state,omitempty"`
}

func (c *GlobalConnectionClient[T]) GetEncryptionsForConnection(
	connectionID int64,
) (*[]GlobalConnectionEncryptionPayload, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v2/accounts/%d/encryptions/?connection_id=%d&state=1",
			c.HostURL,
			c.AccountID,
			connectionID,
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

	resp := new(globalConnectionEncryptionByConnectionResponse)

	err = json.Unmarshal(body, resp)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) > 1 {
		return nil, fmt.Errorf("more than one SSH tunnel config found for the connection")
	}

	return &resp.Data, nil
}

func (c *GlobalConnectionClient[T]) CreateUpdateEncryption(
	encryptionPayload GlobalConnectionEncryptionPayload,
) (*GlobalConnectionEncryptionPayload, error) {

	buffer := new(bytes.Buffer)
	enc := json.NewEncoder(buffer)

	err := enc.Encode(encryptionPayload)
	if err != nil {
		return nil, err
	}

	encryptionID := encryptionPayload.ID
	var postURL string

	if encryptionID == nil {
		// create
		postURL = fmt.Sprintf(
			"%s/v2/accounts/%d/encryptions/",
			c.HostURL,
			c.AccountID,
		)
	} else {
		// update
		postURL = fmt.Sprintf(
			"%s/v2/accounts/%d/encryptions/%d/",
			c.HostURL,
			c.AccountID,
			*encryptionID,
		)
	}

	req, err := http.NewRequest("POST", postURL, buffer)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	resp := new(globalConnectionEncryptionResponse)
	err = json.Unmarshal(body, resp)
	if err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

type EmptyConfig struct {
}

func (EmptyConfig) AdapterVersion() string {
	return "n/a"
}

type SnowflakeConfig struct {
	Account                *string                   `json:"account,omitempty"`
	Database               *string                   `json:"database,omitempty"`
	Warehouse              *string                   `json:"warehouse,omitempty"`
	ClientSessionKeepAlive *bool                     `json:"client_session_keep_alive,omitempty"`
	Role                   nullable.Nullable[string] `json:"role,omitempty"`
	AllowSso               *bool                     `json:"allow_sso,omitempty"`
	OauthClientID          *string                   `json:"oauth_client_id,omitempty"`
	OauthClientSecret      *string                   `json:"oauth_client_secret,omitempty"`
}

func (SnowflakeConfig) AdapterVersion() string {
	return "snowflake_v0"
}

type BigQueryConfig struct {
	ProjectID                 *string                   `json:"project_id,omitempty"`
	TimeoutSeconds            *int64                    `json:"timeout_seconds,omitempty"`
	PrivateKeyID              *string                   `json:"private_key_id,omitempty"`
	PrivateKey                *string                   `json:"private_key,omitempty"`
	ClientEmail               *string                   `json:"client_email,omitempty"`
	ClientID                  *string                   `json:"client_id,omitempty"`
	AuthURI                   *string                   `json:"auth_uri,omitempty"`
	TokenURI                  *string                   `json:"token_uri,omitempty"`
	AuthProviderX509CertURL   *string                   `json:"auth_provider_x509_cert_url,omitempty"`
	ClientX509CertURL         *string                   `json:"client_x509_cert_url,omitempty"`
	Priority                  nullable.Nullable[string] `json:"priority,omitempty"`
	Retries                   *int64                    `json:"retries,omitempty"` //not nullable because there is a default in the UI
	Location                  nullable.Nullable[string] `json:"location,omitempty"`
	MaximumBytesBilled        nullable.Nullable[int64]  `json:"maximum_bytes_billed,omitempty"`
	ExecutionProject          nullable.Nullable[string] `json:"execution_project,omitempty"`
	ImpersonateServiceAccount nullable.Nullable[string] `json:"impersonate_service_account,omitempty"`
	JobRetryDeadlineSeconds   nullable.Nullable[int64]  `json:"job_retry_deadline_seconds,omitempty"`
	JobCreationTimeoutSeconds nullable.Nullable[int64]  `json:"job_creation_timeout_seconds,omitempty"`
	ApplicationID             nullable.Nullable[string] `json:"application_id,omitempty"`
	ApplicationSecret         nullable.Nullable[string] `json:"application_secret,omitempty"`
	GcsBucket                 nullable.Nullable[string] `json:"gcs_bucket,omitempty"`
	DataprocRegion            nullable.Nullable[string] `json:"dataproc_region,omitempty"`
	DataprocClusterName       nullable.Nullable[string] `json:"dataproc_cluster_name,omitempty"`
	Scopes                    []string                  `json:"scopes,omitempty"` //not nullable because there is a default in the UI
}

func (BigQueryConfig) AdapterVersion() string {
	return "bigquery_v0"
}

type DatabricksConfig struct {
	Host         *string                   `json:"host,omitempty"`
	HTTPPath     *string                   `json:"http_path,omitempty"`
	Catalog      nullable.Nullable[string] `json:"catalog,omitempty"`
	ClientID     nullable.Nullable[string] `json:"client_id,omitempty"`
	ClientSecret nullable.Nullable[string] `json:"client_secret,omitempty"`
}

func (DatabricksConfig) AdapterVersion() string {
	return "databricks_v0"
}

// Redshift and Postgres are the same today but they might diverge in the future to support more authentication methods
type RedshiftConfig struct {
	HostName *string                   `json:"hostname,omitempty"`
	Port     *int64                    `json:"port,omitempty"`
	DBName   nullable.Nullable[string] `json:"dbname,omitempty"`
}

func (RedshiftConfig) AdapterVersion() string {
	return "redshift_v0"
}

type PostgresConfig struct {
	HostName *string                   `json:"hostname,omitempty"`
	Port     *int64                    `json:"port,omitempty"`
	DBName   nullable.Nullable[string] `json:"dbname,omitempty"`
}

func (PostgresConfig) AdapterVersion() string {
	return "postgres_v0"
}

var FabricDriver = "ODBC Driver 18 for SQL Server"

type FabricConfig struct {
	Driver       *string `json:"driver,omitempty"`
	Server       *string `json:"server,omitempty"`
	Port         *int64  `json:"port,omitempty"`
	Database     *string `json:"database,omitempty"`
	Retries      *int64  `json:"retries,omitempty"`
	LoginTimeout *int64  `json:"login_timeout,omitempty"`
	QueryTimeout *int64  `json:"query_timeout,omitempty"`
}

func (FabricConfig) AdapterVersion() string {
	return "fabric_v0"
}

// Right now Synapse and Fabric are the same
// If they diverge in the future, we can update the SynapseConfig struct
var SynapseDriver = FabricDriver

type SynapseConfig struct {
	Driver       *string `json:"driver,omitempty"`
	Host         *string `json:"host,omitempty"`
	Port         *int64  `json:"port,omitempty"`
	Database     *string `json:"database,omitempty"`
	Retries      *int64  `json:"retries,omitempty"`
	LoginTimeout *int64  `json:"login_timeout,omitempty"`
	QueryTimeout *int64  `json:"query_timeout,omitempty"`
}

func (SynapseConfig) AdapterVersion() string {
	return "synapse_v0"
}

type StarburstConfig struct {
	Method *string `json:"method,omitempty"`
	Host   *string `json:"host,omitempty"`
	Port   *int64  `json:"port,omitempty"`
}

func (StarburstConfig) AdapterVersion() string {
	return "trino_v0"
}

type AthenaConfig struct {
	RegionName        *string                   `json:"region_name,omitempty"`
	Database          *string                   `json:"database,omitempty"`
	S3StagingDir      *string                   `json:"s3_staging_dir,omitempty"`
	WorkGroup         nullable.Nullable[string] `json:"work_group,omitempty"`
	SparkWorkGroup    nullable.Nullable[string] `json:"spark_work_group,omitempty"`
	S3DataDir         nullable.Nullable[string] `json:"s3_data_dir,omitempty"`
	S3DataNaming      nullable.Nullable[string] `json:"s3_data_naming,omitempty"`
	S3TmpTableDir     nullable.Nullable[string] `json:"s3_tmp_table_dir,omitempty"`
	PollInterval      nullable.Nullable[int64]  `json:"poll_interval,omitempty"`
	NumRetries        nullable.Nullable[int64]  `json:"num_retries,omitempty"`
	NumBoto3Retries   nullable.Nullable[int64]  `json:"num_boto3_retries,omitempty"`
	NumIcebergRetries nullable.Nullable[int64]  `json:"num_iceberg_retries,omitempty"`
}

func (AthenaConfig) AdapterVersion() string {
	return "athena_v0"
}

type ApacheSparkConfig struct {
	Method         *string                   `json:"method,omitempty"`
	Host           *string                   `json:"host,omitempty"`
	Port           *int64                    `json:"port,omitempty"`
	Cluster        *string                   `json:"cluster,omitempty"`
	ConnectTimeout *int64                    `json:"connect_timeout,omitempty"`
	ConnectRetries *int64                    `json:"connect_retries,omitempty"`
	Organization   nullable.Nullable[string] `json:"organization,omitempty"`
	User           nullable.Nullable[string] `json:"user,omitempty"`
	Auth           nullable.Nullable[string] `json:"auth,omitempty"`
	// KerberosServiceName any    `json:"kerberos_service_name,omitempty"` // This field comes back but can't be set from the UI
}

func (ApacheSparkConfig) AdapterVersion() string {
	return "apache_spark_v0"
}

type TeradataConfig struct {
	Port           *string `json:"port,omitempty"`
	TMode          *string `json:"tmode,omitempty"`
	Host           *string `json:"host,omitempty"`
	Retries        *int64  `json:"retries,omitempty"`
	RequestTimeout *int64  `json:"request_timeout,omitempty"`
}

func (TeradataConfig) AdapterVersion() string {
	return "teradata_v0"
}
