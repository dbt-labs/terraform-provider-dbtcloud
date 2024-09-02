package dbt_cloud

import (
	"encoding/json"
	"fmt"
)

type GlobalConnectionSummary struct {
	ID                    int64  `json:"id"`
	CreatedAt             string `json:"created_at"`
	UpdatedAt             string `json:"updated_at"`
	AccountID             int64  `json:"account_id"`
	Name                  string `json:"name"`
	AdapterVersion        string `json:"adapter_version"`
	PrivateLinkEndpointID *int64 `json:"private_link_endpoint_id"`
	IsSSHTunnelEnabled    bool   `json:"is_ssh_tunnel_enabled"`
	OauthConfigurationID  *int64 `json:"oauth_configuration_id"`
	EnvironmentCount      int64  `json:"environment__count"`
}

func (c *Client) GetAllConnections() ([]GlobalConnectionSummary, error) {

	url := fmt.Sprintf(
		`%s/v3/accounts/%d/connections/`,
		c.HostURL,
		c.AccountID,
	)

	allConnectionsRaw := c.GetData(url)

	allConnections := []GlobalConnectionSummary{}
	for _, connection := range allConnectionsRaw {

		data, _ := json.Marshal(connection)
		currentConnection := GlobalConnectionSummary{}
		err := json.Unmarshal(data, &currentConnection)
		if err != nil {
			return nil, err
		}
		allConnections = append(allConnections, currentConnection)
	}
	return allConnections, nil
}
