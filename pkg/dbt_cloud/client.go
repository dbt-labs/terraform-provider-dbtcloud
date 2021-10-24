package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const HostURL string = "https://cloud.getdbt.com/api/"

// Client -
type Client struct {
	HostURL    string
	HTTPClient *http.Client
	Token      string
	AccountURL string
	AccountID  int
}

type ResponseStatus struct {
	Code              int    `json:"code"`
	Is_Success        bool   `json:"is_success"`
	User_Message      string `json:"user_message"`
	Developer_Message string `json:"developer_message"`
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
	EnterpriseUniqueIdentifier     int    `json:"enterprise_unique_identifier"`
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
	Status ResponseStatus   `json:"status`
	Data   AuthResponseData `json:"data`
}

// NewClient -
func NewClient(account_id *int, token *string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		HostURL:    HostURL,
		Token:      *token,
		AccountID:  *account_id,
	}

	if (account_id != nil) && (token != nil) {
		url := fmt.Sprintf("%s/accounts/%s", HostURL, strconv.Itoa(*account_id))

		// authenticate
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		body, err := c.doRequest(req)

		// parse response body
		ar := AuthResponse{}
		err = json.Unmarshal(body, &ar)
		if err != nil {
			return nil, err
		}

		c.AccountURL = url
	}

	return &c, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.Token))

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if (res.StatusCode != http.StatusOK) && (res.StatusCode != 201) {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}
