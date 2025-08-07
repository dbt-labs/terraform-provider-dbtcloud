package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

type IPRestrictions []IPRestrictionsRule

type IPRestrictionsRule struct {
	ID                     int64   `json:"id,omitempty"`
	IPRestrictionRuleSetID int64   `json:"ip_restriction_rule_set_id,omitempty"`
	Name                   string  `json:"name"`
	Type                   int64   `json:"type"`
	AccountID              int64   `json:"account_id,omitempty"`
	Description            string  `json:"description"`
	Cidrs                  []Cidrs `json:"cidrs,"`
	RuleSetEnabled         bool    `json:"rule_set_enabled"`
	// not needed for TF
	State                   int64 `json:"state,omitempty"`
	CreatedByID             int64 `json:"created_by_id,omitempty"`
	EnabledForServiceTokens bool  `json:"enabled_for_service_tokens,omitempty"`
}

type Cidrs struct {
	ID                  int64  `json:"id,omitempty"`
	IPRestrictionRuleID int64  `json:"ip_restriction_rule_id,omitempty"`
	Cidr                string `json:"cidr,omitempty"`
	CidrIpv6            string `json:"cidr_ipv6,omitempty"`
	// not needed for TF
	State     int64  `json:"state,omitempty"`
	Enabled   bool   `json:"enabled,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type IPRestrictionsResponse struct {
	Data   IPRestrictions `json:"data"`
	Status ResponseStatus `json:"status"`
}

type IPRestrictionsRuleResponse struct {
	Data   IPRestrictionsRule `json:"data"`
	Status ResponseStatus     `json:"status"`
}

func (c *Client) GetIPRestrictions() (*IPRestrictions, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%s/ip-restrictions/",
			c.HostURL,
			strconv.Itoa(int(c.AccountID)),
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

	ipRestrictionsResponse := IPRestrictionsResponse{}
	err = json.Unmarshal(body, &ipRestrictionsResponse)
	if err != nil {
		return nil, err
	}

	return &ipRestrictionsResponse.Data, nil
}

func (c *Client) GetIPRestrictionsRule(ruleID int64) (*IPRestrictionsRule, error) {
	allIPRestrictions, err := c.GetIPRestrictions()
	if err != nil {
		return nil, err
	}

	foundIPRestriction := lo.Filter(
		*allIPRestrictions,
		func(ipRestrictionsRule IPRestrictionsRule, _ int) bool {
			return ipRestrictionsRule.ID == ruleID
		},
	)

	if len(foundIPRestriction) == 0 {
		return nil, nil
	}
	return &foundIPRestriction[0], nil

}

func (c *Client) CreateIPRestrictionsRule(
	ipRestrictionsRule IPRestrictionsRule,
) (*IPRestrictionsRule, error) {

	newIPRestrictionsData, err := json.Marshal(ipRestrictionsRule)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/v3/accounts/%s/ip-restrictions/", c.HostURL, strconv.Itoa(int(c.AccountID))),
		strings.NewReader(string(newIPRestrictionsData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	ipRestrictionsRuleResponse := IPRestrictionsRuleResponse{}
	err = json.Unmarshal(body, &ipRestrictionsRuleResponse)
	if err != nil {
		return nil, err
	}

	return &ipRestrictionsRuleResponse.Data, nil
}

func (c *Client) UpdateIPRestrictionsRule(
	ipRestrictionsId string,
	ipRestrictions IPRestrictionsRule,
) (*IPRestrictionsRule, error) {
	ipRestrictionsData, err := json.Marshal(ipRestrictions)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf(
			"%s/v3/accounts/%s/ip-restrictions/%s",
			c.HostURL,
			strconv.Itoa(int(c.AccountID)),
			ipRestrictionsId,
		),
		strings.NewReader(string(ipRestrictionsData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	ipRestrictionsRuleResponse := IPRestrictionsRuleResponse{}
	err = json.Unmarshal(body, &ipRestrictionsRuleResponse)
	if err != nil {
		return nil, err
	}

	return &ipRestrictionsRuleResponse.Data, nil
}

func (c *Client) DeleteIPRestrictions(ipRestrictions IPRestrictions) error {
	for _, ipRestrictionsRule := range ipRestrictions {
		err := c.DeleteIPRestrictionsRule(ipRestrictionsRule.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) DeleteIPRestrictionsRule(ipRestrictionsRuleID int64) error {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf(
			"%s/v3/accounts/%s/ip-restrictions/%d",
			c.HostURL,
			strconv.Itoa(int(c.AccountID)),
			ipRestrictionsRuleID,
		),
		nil,
	)
	if err != nil {
		return err
	}

	_, err = c.doRequestWithRetry(req)
	if err != nil {
		return err
	}

	return nil
}
