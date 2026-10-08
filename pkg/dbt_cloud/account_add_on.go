package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// AccountAddOn is the status of one add-on product on an account.
type AccountAddOn struct {
	Product        string  `json:"product"`
	State          *string `json:"state"`
	CanActivate    bool    `json:"can_activate"`
	CanTrial       bool    `json:"can_trial"`
	TrialStartedAt *string `json:"trial_started_at"`
	TrialEndsAt    *string `json:"trial_ends_at"`
	TrialConsumed  bool    `json:"trial_consumed"`
}

// AccountAddOnSpendLimit is the spend limit configured for one add-on product.
// A nil limit means that no limit is set.
type AccountAddOnSpendLimit struct {
	Product               string `json:"product"`
	SpendLimitNanodollars *int64 `json:"spend_limit_nanodollars"`
}

type accountAddOnListData struct {
	AddOns []AccountAddOn `json:"add_ons"`
}

type accountAddOnListResponse struct {
	Data   accountAddOnListData `json:"data"`
	Status ResponseStatus       `json:"status"`
}

type accountAddOnSpendLimitResponse struct {
	Data   AccountAddOnSpendLimit `json:"data"`
	Status ResponseStatus         `json:"status"`
}

// addOnURL builds the add-on URL for the account. The add-on paths carry no
// trailing slash, unlike most of the other account paths.
func (c *Client) addOnURL(parts ...string) string {
	url := fmt.Sprintf("%s/private/accounts/%d/add-ons", c.HostURL, c.AccountID)
	if len(parts) > 0 {
		url = url + "/" + strings.Join(parts, "/")
	}
	return url
}

// GetAccountAddOns returns one entry for every add-on product the account can use.
func (c *Client) GetAccountAddOns() ([]AccountAddOn, error) {
	req, err := http.NewRequest("GET", c.addOnURL(), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	response := accountAddOnListResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Data.AddOns, nil
}

// GetAccountAddOn returns the entry for one product, or nil when the account
// cannot use that product. A product the account has simply never turned on is
// still listed, with no state, so a nil result means something different: the
// product is turned off for the account.
func (c *Client) GetAccountAddOn(product string) (*AccountAddOn, error) {
	addOns, err := c.GetAccountAddOns()
	if err != nil {
		return nil, err
	}

	for _, addOn := range addOns {
		if strings.EqualFold(addOn.Product, product) {
			match := addOn
			return &match, nil
		}
	}

	return nil, nil
}

// GetAccountAddOnSpendLimit returns the spend limit in nanodollars, or nil when
// no limit is set. The endpoint answers 404 when the product has no trial and no
// active subscription to configure, which is reported as no limit.
func (c *Client) GetAccountAddOnSpendLimit(product string) (*int64, error) {
	req, err := http.NewRequest("GET", c.addOnURL(product, "spend-limit"), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			return nil, nil
		}
		return nil, err
	}

	response := accountAddOnSpendLimitResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Data.SpendLimitNanodollars, nil
}
