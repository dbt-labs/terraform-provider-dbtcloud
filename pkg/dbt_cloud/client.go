package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var versionString = "dev"

type Client struct {
	HostURL              string
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

	c := Client{
		HTTPClient:           &http.Client{Timeout: 30 * time.Second},
		HostURL:              *host_url,
		Token:                *token,
		AccountID:            *account_id,
		RetryIntervalSeconds: *retryIntervalSeconds,
		MaxRetries:           *maxRetries,
		RetriableStatusCodes: retriableStatusCodes,
	}

	_, runningAcceptanceTests := os.LookupEnv("TF_ACC")
	if account_id != nil && !runningAcceptanceTests {
		url := fmt.Sprintf("%s/v2/accounts/", *host_url)

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

		// sometimes the err field comes nil but the status code is 404, so we check the status code
		// this happens frequently when checking if the resource was destroyed
		if res.StatusCode == 404 && req.Method == "GET" {
			isResourceNotFound, err := isResourceNotFoundError(body)
			if err != nil {
				return nil, err
			}
			if isResourceNotFound {
				return nil, fmt.Errorf("resource-not-found: %s", req.URL)
			}
		}

		if res.StatusCode == 400 {
			return nil, fmt.Errorf("resource-not-found: %s", body)
		}

		if res.StatusCode == 500 {
			return nil, fmt.Errorf("internal-server-error: %s", body)
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
