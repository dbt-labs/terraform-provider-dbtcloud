package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/samber/lo"
)

const (
	STATE_ACTIVE           = 1
	STATE_DELETED          = 2
	ID_DELIMITER           = ":"
	NUM_THREADS_CREDENTIAL = 6
)

var (
	PermissionSets = []string{
		"owner",
		"member",
		"account_admin",
		"security_admin",
		"billing_admin",
		"admin",
		"database_admin",
		"git_admin",
		"team_admin",
		"job_admin",
		"job_runner",
		"job_viewer",
		"analyst",
		"developer",
		"stakeholder",
		"readonly",
		"project_creator",
		"account_viewer",
		"metadata_only",
		"semantic_layer_only",
		"webhooks_only",
		"fusion_admin",
		"cost_management_viewer",
		"cost_management_admin",
		"manage_marketplace_apps",
	}
)

// This is not used now but allows us to find the list of permission sets allowed
// Terraform doesn't allow to run it in the offline validation mode though as we don't have access to the configuration context
type ConstantsResponse struct {
	Data   Constants      `json:"data"`
	Status ResponseStatus `json:"status"`
}

type Constants struct {
	PermissionSets map[string]string `json:"permissions_sets"`
}

func (c *Client) GetConstants() (*Constants, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/v2/constants/", c.HostURL),
		nil,
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	constantsResponse := ConstantsResponse{}
	err = json.Unmarshal(body, &constantsResponse)
	if err != nil {
		return nil, err
	}

	return &constantsResponse.Data, nil
}

func (c *Client) GetPermissionIDs() ([]string, error) {
	constants, err := c.GetConstants()
	if err != nil {
		return nil, err
	}

	return lo.Keys(constants.PermissionSets), nil
}
