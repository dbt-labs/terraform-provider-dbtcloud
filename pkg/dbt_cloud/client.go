package dbt_cloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var versionString = "dev"

// API path constants for consistent URL construction
const (
	APIVersionV3 = "v3"
	APIVersionV2 = "v2"
)

// Resource name constants
const (
	ResourceAccounts             = "accounts"
	ResourcePrivatelinkEndpoints = "private-endpoints"
	ResourceGroups               = "groups"
	ResourceEnvironments         = "environments"
	ResourceNotifications        = "notifications"
	ResourceServiceTokens        = "service-tokens"
	ResourceLicenseMaps          = "license-maps"
)

type Client struct {
	HostURL              *url.URL
	HTTPClient           *http.Client
	Token                string
	AccountURL           string
	AccountID            int64
	RetryIntervalSeconds int
	MaxRetries           int
	RetriableStatusCodes []string
	DisableRetry         bool
	TimeoutSeconds       int
}

type ResponseStatus struct {
	Code              int    `json:"code"`
	Is_Success        bool   `json:"is_success"`
	User_Message      string `json:"user_message"`
	Developer_Message string `json:"developer_message"`
}

type ResponseExtraFilters struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type ResponseExtraPagination struct {
	Count      int `json:"count"`
	TotalCount int `json:"total_count"`
}

type ResponseExtra struct {
	Pagination ResponseExtraPagination `json:"pagination"`
	Filters    ResponseExtraFilters    `json:"filters"`
}

type AuthResponseData struct {
	DocsJobId                      int    `json:"docs_job_id"`
	FreshnessJobId                 int    `json:"freshness_job_id"`
	LockReason                     string `json:"lock_reason"`
	UnlockIfSubscriptionRenewed    bool   `json:"unlock_if_subscription_renewed"`
	ReadOnlySeats                  int    `json:"read_only_seats"`
	Id                             int64  `json:"id"`
	Name                           string `json:"name"`
	State                          int    `json:"state"`
	Plan                           string `json:"plan"`
	PendingCancel                  bool   `json:"pending_cancel"`
	RunSlots                       int    `json:"run_slots"`
	DeveloperSeats                 int    `json:"developer_seats"`
	QueueLimit                     int    `json:"queue_limit"`
	PodMemoryRequestMebibytes      int    `json:"pod_memory_request_mebibytes"`
	RunDurationLimitSeconds        int    `json:"run_duration_limit_seconds"`
	EnterpriseAuthenticationMethod string `json:"enterprise_authentication_method"`
	EnterpriseLoginSlug            string `json:"enterprise_login_slug"`
	EnterpriseUniqueIdentifier     string `json:"enterprise_unique_identifier"`
	BillingEmailAddress            string `json:"billing_email_address"`
	Locked                         bool   `json:"locked"`
	DevelopFileSystem              bool   `json:"develop_file_system"`
	UnlockedAt                     string `json:"unlocked_at"`
	CreatedAt                      string `json:"created_at"`
	UpdatedAt                      string `json:"updated_at"`
	StarterRepoUrl                 string `json:"starter_repo_url"`
	SsoReauth                      bool   `json:"sso_reauth"`
	GitAuthLevel                   string `json:"git_auth_level"`
	DocsJob                        string `json:"docs_job"`
	FreshnessJob                   string `json:"freshness_job"`
	EnterpriseLoginUrl             string `json:"enterprise_login_url"`
}

// AuthResponse -
type AuthResponse struct {
	Status ResponseStatus     `json:"status"`
	Data   []AuthResponseData `json:"data"`
}

// Parses the error we get to see if it is a 404 for a missing resource
type APIError struct {
	Data   interface{} `json:"data"`
	Status struct {
		Code             int    `json:"code"`
		DeveloperMessage string `json:"developer_message"`
		IsSuccess        bool   `json:"is_success"`
		UserMessage      string `json:"user_message"`
	} `json:"status"`
}

// DefaultTimeoutSeconds is the HTTP client timeout used when the timeout_seconds
// provider attribute is not set.
const DefaultTimeoutSeconds = 30

// maxIdleConns caps the connections kept for reuse. http.DefaultTransport keeps only
// 2 idle connections per host, so a Terraform run issuing API calls in parallel has to
// dial and complete a new TLS handshake for most of them. Keeping them instead removes
// that churn.
const maxIdleConns = 100

// NewHTTPClient builds the HTTP client used for every dbt Cloud API call.
func NewHTTPClient(timeoutSeconds int) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = maxIdleConns
	transport.MaxIdleConnsPerHost = maxIdleConns

	return &http.Client{
		Timeout:   time.Duration(timeoutSeconds) * time.Second,
		Transport: transport,
	}
}

// NewClient -
func NewClient(account_id *int64, token *string, host_url *string, maxRetries *int, retryIntervalSeconds *int, retriableStatusCodes []string, skipCredentialsValidation bool, timeoutSeconds *int) (*Client, error) {

	if (token == nil) || (*token == "") {
		return nil, fmt.Errorf("token is set but it is empty")
	}

	// Parse and validate the host URL
	parsedURL, err := url.Parse(*host_url)
	if err != nil {
		return nil, fmt.Errorf("invalid host URL '%s': %w", *host_url, err)
	}

	c := Client{
		HTTPClient:           NewHTTPClient(*timeoutSeconds),
		HostURL:              parsedURL,
		Token:                *token,
		AccountID:            *account_id,
		RetryIntervalSeconds: *retryIntervalSeconds,
		MaxRetries:           *maxRetries,
		RetriableStatusCodes: retriableStatusCodes,
		TimeoutSeconds:       *timeoutSeconds,
	}

	_, runningAcceptanceTests := os.LookupEnv("TF_ACC")
	if !runningAcceptanceTests && !skipCredentialsValidation {
		url := c.BuildV2URL(ResourceAccounts)

		// authenticate
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		body, err := c.doRequestWithRetry(req)
		if err != nil {
			return nil, err
		}

		// parse response body
		ar := AuthResponse{}
		err = json.Unmarshal(body, &ar)
		if err != nil {
			return nil, err
		}

		for _, account := range ar.Data {
			if account.Id == *account_id {
				c.AccountURL = url
				return &c, nil
			}
		}

		return nil, fmt.Errorf(
			"the token is valid but does not have access to the account id %d. This might be due to a lack of permissions or because IP restrictions are in place for the account",
			*account_id,
		)

	}

	return &c, nil
}

// defaultRetriableHTTPCodes lists the HTTP status codes the client retries by
// default when the provider has not supplied an explicit retriable_status_codes
// list. They cover the transient upstream / proxy / rate-limit failures that
// have caused real customer apply errors and post-merge CI flakes:
//
//   - 408 Request Timeout
//   - 429 Too Many Requests
//   - 500 Internal Server Error
//   - 502 Bad Gateway
//   - 503 Service Unavailable
//   - 504 Gateway Timeout
//
// 501 Not Implemented is intentionally excluded because it indicates a
// permanent missing feature rather than a transient failure.
var defaultRetriableHTTPCodes = []int{408, 429, 500, 502, 503, 504}

func (c *Client) doRequestWithRetry(req *http.Request) ([]byte, error) {
	// Provider config may want to disable retries even when client code asks
	// for them. Floor MaxRetries at 1 so we always make at least one attempt.
	if c.DisableRetry || c.MaxRetries <= 0 {
		c.MaxRetries = 1
	}

	setRequestHeaders(req, c.Token)

	// The first attempt consumes the request body, so a retry can only send the same
	// payload again if the body can be rewound. http.NewRequest populates GetBody for
	// the in-memory readers this client uses; without it a retry would send an empty
	// body, so those requests are not retried at all.
	bodyCanBeRewound := req.Body == nil || req.GetBody != nil

	var lastErr error
	for attempt := 0; attempt < c.MaxRetries; attempt++ {
		if attempt > 0 && req.Body != nil {
			rewoundBody, rewindErr := req.GetBody()
			if rewindErr != nil {
				return nil, errors.Join(
					lastErr,
					fmt.Errorf("could not rewind the request body to retry: %w", rewindErr),
				)
			}
			req.Body = rewoundBody
		}

		body, statusCode, attemptErr := c.attemptRequest(req)
		if attemptErr == nil {
			return body, nil
		}
		lastErr = attemptErr

		attemptsRemaining := c.MaxRetries - 1 - attempt

		// Transport-level failures (statusCode == 0: DNS, connection reset,
		// timeout) and HTTP responses listed in retriable_status_codes /
		// defaultRetriableHTTPCodes are retried with exponential backoff.
		// Everything else surfaces immediately so callers see the same
		// terminal error they did before this fix.
		retriable := bodyCanBeRewound &&
			(statusCode == 0 ||
				isHTTPCodeRetriable(statusCode, c.RetriableStatusCodes))

		if !retriable || attemptsRemaining == 0 {
			return nil, attemptErr
		}

		waitDuration := time.Duration(c.RetryIntervalSeconds) * time.Second
		if attempt > 0 {
			// Exponential backoff after the first retry.
			waitDuration = time.Duration(c.RetryIntervalSeconds) * time.Second * (1 << attempt)
		}
		fmt.Printf(
			"Request to %s failed with retriable error (attempt %d/%d, status %d): %v. Waiting %v before retrying...\n",
			req.URL, attempt+1, c.MaxRetries, statusCode, attemptErr, waitDuration,
		)
		time.Sleep(waitDuration)
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("no attempt was made for request %s", req.URL)
}

// attemptRequest performs a single HTTP attempt and returns the body on success
// or a fully formatted terminal error on failure. The error message formats
// here are preserved verbatim from the pre-retry-fix implementation so that
// callers matching on substrings such as "internal-server-error:",
// "resource-not-found", "unauthorized:", "forbidden:" or "unexpected status
// code" continue to behave the same once retries are exhausted.
//
// statusCode is 0 when the request failed at the transport layer; the caller
// in doRequestWithRetry treats those as retriable.
func (c *Client) attemptRequest(req *http.Request) ([]byte, int, error) {
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	body, readErr := io.ReadAll(res.Body)
	statusCode := res.StatusCode

	// Handle 404 errors - check if it's a true not found or a permissions issue
	if statusCode == 404 {
		isResourceNotFound, apiErr, parseErr := parseAPIError(body)
		if parseErr != nil {
			// If we can't parse the error, return a generic 404
			return nil, statusCode, fmt.Errorf("resource-not-found (status 404): URL: %s, Response: %s", req.URL, body)
		}

		if isResourceNotFound {
			// Check if the error message mentions permissions - this is a common pattern in dbt Cloud API
			userMsg := strings.ToLower(apiErr.Status.UserMessage)
			if strings.Contains(userMsg, "permission") || strings.Contains(userMsg, "proper permissions") {
				return nil, statusCode, fmt.Errorf("resource-not-found-permissions: The resource was not found, but this may be due to insufficient permissions. The API token may not have access to this resource or the environment it belongs to.\n\nStatus: 404\nURL: %s\nMessage: %s", req.URL, apiErr.Status.UserMessage)
			}

			// For GET requests or DELETE operations, this is typically a legitimate not-found
			// (DELETE gets 404 when resource already deleted, which is fine)
			if req.Method == "GET" || req.Method == "DELETE" {
				return nil, statusCode, fmt.Errorf("resource-not-found: %s", req.URL)
			}

			// For POST/PUT on non-permission 404s, provide additional context
			// This helps with update/create operations that fail due to permissions
			return nil, statusCode, fmt.Errorf("resource-not-found: The resource was not found. If you are updating a resource, this may indicate insufficient permissions.\n\nStatus: 404\nURL: %s\nMessage: %s", req.URL, apiErr.Status.UserMessage)
		}
	}

	if statusCode == 400 {
		return nil, statusCode, fmt.Errorf("resource-not-found: %s", body)
	}

	// Handle permission errors (401 Unauthorized, 403 Forbidden)
	if statusCode == 401 {
		return nil, statusCode, fmt.Errorf("unauthorized: The API token does not have permission to access this resource. Status: 401, URL: %s, Response: %s", req.URL, body)
	}

	if statusCode == 403 {
		return nil, statusCode, fmt.Errorf("forbidden: The API token does not have permission to perform this action. This may be due to environment-level permissions or other access restrictions. Status: 403, URL: %s, Response: %s", req.URL, body)
	}

	if statusCode == 500 {
		return nil, statusCode, fmt.Errorf("internal-server-error: %s", body)
	}

	// Check for other non-2xx status codes
	if statusCode < 200 || statusCode >= 300 {
		return nil, statusCode, fmt.Errorf("unexpected status code %d: %s, URL: %s", statusCode, body, req.URL)
	}

	if readErr != nil {
		return nil, statusCode, readErr
	}
	return body, statusCode, nil
}

// isHTTPCodeRetriable reports whether statusCode should trigger a retry. The
// provider-supplied retriableStatusCodes take precedence; if statusCode is not
// listed there it is also matched against defaultRetriableHTTPCodes so callers
// that never configured retriable_status_codes still benefit from retry on
// canonical transient errors.
func isHTTPCodeRetriable(statusCode int, retriableStatusCodes []string) bool {
	if statusCode == 0 {
		return false
	}
	for _, code := range retriableStatusCodes {
		if code == fmt.Sprintf("%d", statusCode) {
			return true
		}
	}
	for _, code := range defaultRetriableHTTPCodes {
		if code == statusCode {
			return true
		}
	}
	return false
}

// isErrorRetriable is preserved for backwards compatibility with prior call
// sites. It now defers to isHTTPCodeRetriable so the retry decision is
// consistent across the client.
func isErrorRetriable(statusCode int, retriableStatusCodes []string) bool {
	return isHTTPCodeRetriable(statusCode, retriableStatusCodes)
}

func isResourceNotFoundError(body []byte) (bool, error) {
	var apiErr APIError
	if unmarshalErr := json.Unmarshal([]byte(body), &apiErr); unmarshalErr != nil {
		return false, unmarshalErr
	}
	// in this case, the body of the error mentions a 404, this is different from a 404 due to a wrong URL
	if apiErr.Status.Code == 404 {
		return true, nil
	}
	return false, nil
}

// parseAPIError parses the API error response and returns whether it's a 404, the full error details, and any parse error
func parseAPIError(body []byte) (bool, *APIError, error) {
	var apiErr APIError
	if unmarshalErr := json.Unmarshal([]byte(body), &apiErr); unmarshalErr != nil {
		return false, nil, unmarshalErr
	}

	isNotFound := apiErr.Status.Code == 404
	return isNotFound, &apiErr, nil
}

func setRequestHeaders(req *http.Request, token string) {
	userAgentWithVersion := fmt.Sprintf(
		"terraform-provider-dbtcloud/%s",
		versionString,
	)

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))
	req.Header.Set("User-Agent", userAgentWithVersion)
}

// BuildAccountAPIURL constructs API URLs with consistent formatting
func (c *Client) BuildAccountAPIURL(version, resource string, pathParams ...interface{}) string {
	accountResource := fmt.Sprintf("accounts/%d", c.AccountID)
	if resource != "" {
		accountResource = fmt.Sprintf("%s/%s", accountResource, resource)
	}
	return c.BuildAPIURL(version, accountResource, pathParams...)
}

func (c *Client) BuildAPIURL(version, resource string, pathParams ...interface{}) string {
	if c.HostURL == nil {
		return ""
	}

	// Build the base path
	basePath := fmt.Sprintf("/%s", version)

	// Add resource path
	if resource != "" {
		basePath = fmt.Sprintf("%s/%s", basePath, resource)
	}

	// Add any additional path parameters
	if len(pathParams) > 0 {
		for _, param := range pathParams {
			basePath = fmt.Sprintf("%s/%v", basePath, param)
		}
	}

	// Ensure trailing slash
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}

	return fmt.Sprintf("%s%s", c.HostURL, basePath)
}

// BuildV3URL is a convenience method for v3 API endpoints
func (c *Client) BuildV3URL(resource string, pathParams ...interface{}) string {
	return c.BuildAPIURL(APIVersionV3, resource, pathParams...)
}

// BuildV2URL is a convenience method for v2 API endpoints
func (c *Client) BuildV2URL(resource string, pathParams ...interface{}) string {
	return c.BuildAPIURL(APIVersionV2, resource, pathParams...)
}

// BuildAccountV3URL is a convenience method for v3 API endpoints
func (c *Client) BuildAccountV3URL(resource string, pathParams ...interface{}) string {
	return c.BuildAccountAPIURL(APIVersionV3, resource, pathParams...)
}

// BuildAccountV2URL is a convenience method for v2 API endpoints
func (c *Client) BuildAccountV2URL(resource string, pathParams ...interface{}) string {
	return c.BuildAccountAPIURL(APIVersionV2, resource, pathParams...)
}
