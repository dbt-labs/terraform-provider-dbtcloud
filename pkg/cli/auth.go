package cli

import (
	"fmt"
	"os"
	"strconv"

	dbtcloud "github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
)

type AuthConfig struct {
	Token     string
	AccountID int64
	HostURL   string
}

func LoadAuthConfig() (*AuthConfig, error) {
	token := os.Getenv("DBT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DBT_TOKEN environment variable is required")
	}

	accountIDStr := os.Getenv("DBT_ACCOUNT_ID")
	if accountIDStr == "" {
		return nil, fmt.Errorf("DBT_ACCOUNT_ID environment variable is required")
	}

	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("DBT_ACCOUNT_ID must be a valid integer: %w", err)
	}

	hostURL := os.Getenv("DBT_HOST_URL")
	if hostURL == "" {
		hostURL = "https://cloud.getdbt.com/api"
	}

	return &AuthConfig{
		Token:     token,
		AccountID: accountID,
		HostURL:   hostURL,
	}, nil
}

func NewClientFromAuth(cfg *AuthConfig) (*dbtcloud.Client, error) {
	maxRetries := 3
	retryInterval := 5
	timeout := 30
	return dbtcloud.NewClient(
		&cfg.AccountID,
		&cfg.Token,
		&cfg.HostURL,
		&maxRetries,
		&retryInterval,
		nil,   // retriableStatusCodes
		false, // skipCredentialsValidation
		&timeout,
	)
}
