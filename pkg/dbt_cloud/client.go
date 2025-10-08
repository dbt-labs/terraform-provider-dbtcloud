package dbt_cloud

import (
	"encoding/json"
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
	ResourcePrivatelinkEndpoints = "private-link-endpoints"
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
	AccountID            int
	RetryIntervalSeconds int
	MaxRetries           int
	RetriableStatusCodes []string
	DisableRetry         bool
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
	Id                             int    `json:"id"`
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

// NewClient -
func NewClient(account_id *int, token *string, host_url *string, maxRetries *int, retryIntervalSeconds *int, retriableStatusCodes []string) (*Client, error) {

	if (token == nil) || (*token == "") {
		return nil, fmt.Errorf("token is set but it is empty")
	}

	// Parse and validate the host URL
	parsedURL, err := url.Parse(*host_url)
	if err != nil {
		return nil, fmt.Errorf("invalid host URL '%s': %w", *host_url, err)
	}

	c := Client{
		HTTPClient:           &http.Client{Timeout: 30 * time.Second},
		HostURL:              parsedURL,
		Token:                *token,
		AccountID:            *account_id,
		RetryIntervalSeconds: *retryIntervalSeconds,
		MaxRetries:           *maxRetries,
		RetriableStatusCodes: retriableStatusCodes,
	}

	_, runningAcceptanceTests := os.LookupEnv("TF_ACC")
	if !runningAcceptanceTests {
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

func (c *Client) doRequestWithRetry(req *http.Request) ([]byte, error) {
	var err error

	// This is needed in case the provider code wants to do retries but the provider config is set to disable retries
	if c.DisableRetry || c.MaxRetries <= 0 {
		c.MaxRetries = 1
	}

	setRequestHeaders(req, c.Token)

	for i := 0; i < c.MaxRetries; i++ {
		res, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)

		// Handle 404 errors - check if it's a true not found or a permissions issue
		if res.StatusCode == 404 {
			isResourceNotFound, apiErr, parseErr := parseAPIError(body)
			if parseErr != nil {
				// If we can't parse the error, return a generic 404
				return nil, fmt.Errorf("resource-not-found (status 404): URL: %s, Response: %s", req.URL, body)
			}

			if isResourceNotFound {
				// Check if the error message mentions permissions - this is a common pattern in dbt Cloud API
				userMsg := strings.ToLower(apiErr.Status.UserMessage)
				if strings.Contains(userMsg, "permission") || strings.Contains(userMsg, "proper permissions") {
					return nil, fmt.Errorf("resource-not-found-permissions: The resource was not found, but this may be due to insufficient permissions. The API token may not have access to this resource or the environment it belongs to.\n\nStatus: 404\nURL: %s\nMessage: %s", req.URL, apiErr.Status.UserMessage)
				}

				// For GET requests, this is typically a legitimate not-found
				if req.Method == "GET" {
					return nil, fmt.Errorf("resource-not-found: %s", req.URL)
				}

				// For POST/PUT/DELETE, a 404 often indicates permissions issues
				return nil, fmt.Errorf("resource-not-found: The resource was not found. If you are updating or deleting a resource, this may indicate insufficient permissions.\n\nStatus: 404\nURL: %s\nMessage: %s", req.URL, apiErr.Status.UserMessage)
			}
		}

		if res.StatusCode == 400 {
			return nil, fmt.Errorf("resource-not-found: %s", body)
		}

		// Handle permission errors (401 Unauthorized, 403 Forbidden)
		if res.StatusCode == 401 {
			return nil, fmt.Errorf("unauthorized: The API token does not have permission to access this resource. Status: 401, URL: %s, Response: %s", req.URL, body)
		}

		if res.StatusCode == 403 {
			return nil, fmt.Errorf("forbidden: The API token does not have permission to perform this action. This may be due to environment-level permissions or other access restrictions. Status: 403, URL: %s, Response: %s", req.URL, body)
		}

		if res.StatusCode == 500 {
			return nil, fmt.Errorf("internal-server-error: %s", body)
		}

		// Check for other non-2xx status codes
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return nil, fmt.Errorf("unexpected status code %d: %s, URL: %s", res.StatusCode, body, req.URL)
		}

		if err == nil {
			return body, nil
		} else {
			if isErrorRetriable(res.StatusCode, c.RetriableStatusCodes) {
				waitDuration := time.Duration(c.RetryIntervalSeconds) * time.Second
				// Exponential backoff
				if i > 0 {
					waitDuration = time.Duration(c.RetryIntervalSeconds) * time.Second * (1 << i) // Exponential backoff
					fmt.Printf("Waiting for %v before retrying...\n", waitDuration)
					time.Sleep(waitDuration)
				} else {
					// Linear backoff for the first retry
					fmt.Printf("Waiting for %d seconds before retrying...\n", c.RetryIntervalSeconds)
					time.Sleep(waitDuration)
				}
				continue
			}

			if strings.Contains(err.Error(), "resource-not-found") {
				return nil, err
			}
		}

		return nil, err
	}

	return nil, fmt.Errorf("max retries reached for request %s: %w", req.URL, err)
}

func isErrorRetriable(statusCode int, retriableStatusCodes []string) bool {
	var retriable bool = false
	for _, code := range retriableStatusCodes {
		if code == fmt.Sprintf("%d", statusCode) {
			retriable = true
			break
		}
	}

	return retriable
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
