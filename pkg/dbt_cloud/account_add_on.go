package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Add-on products that can be activated on an account.
const (
	AddOnProductWizard = "wizard"
	AddOnProductState  = "state"
)

// Lifecycle states an add-on can be in. An account that has never activated or
// trialled a product has no state at all.
const (
	AddOnStateTrial     = "TRIAL"
	AddOnStateActive    = "ACTIVE"
	AddOnStateCancelled = "CANCELLED"
	AddOnStateExpired   = "EXPIRED"
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

type accountAddOnResponse struct {
	Data   AccountAddOn   `json:"data"`
	Status ResponseStatus `json:"status"`
}

type accountAddOnSpendLimitResponse struct {
	Data   AccountAddOnSpendLimit `json:"data"`
	Status ResponseStatus         `json:"status"`
}

// The field carries a meaningful null, so it must not be omitted when empty.
type accountAddOnSpendLimitRequest struct {
	SpendLimitNanodollars *int64 `json:"spend_limit_nanodollars"`
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
// cannot use that product. A product the account has simply never activated is
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

func (c *Client) postAddOnAction(product string, action string) (*AccountAddOn, error) {
	req, err := http.NewRequest(
		"POST",
		c.addOnURL(product, action),
		strings.NewReader("{}"),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	response := accountAddOnResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// ActivateAccountAddOn starts paid use of an add-on product. The call has no
// effect when the product is active already.
func (c *Client) ActivateAccountAddOn(product string) (*AccountAddOn, error) {
	return c.postAddOnAction(product, "activate")
}

// StartAccountAddOnTrial starts the trial of an add-on product. An account gets
// one trial for each product.
func (c *Client) StartAccountAddOnTrial(product string) (*AccountAddOn, error) {
	return c.postAddOnAction(product, "start-trial")
}

// CancelAccountAddOn cancels an add-on product. The call has no effect when the
// product is cancelled already.
func (c *Client) CancelAccountAddOn(product string) (*AccountAddOn, error) {
	return c.postAddOnAction(product, "cancel")
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

// SetAccountAddOnSpendLimit sets the spend limit in nanodollars. A nil limit
// removes the limit.
func (c *Client) SetAccountAddOnSpendLimit(product string, limit *int64) (*int64, error) {
	requestData, err := json.Marshal(accountAddOnSpendLimitRequest{SpendLimitNanodollars: limit})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"PUT",
		c.addOnURL(product, "spend-limit"),
		strings.NewReader(string(requestData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	response := accountAddOnSpendLimitResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Data.SpendLimitNanodollars, nil
}
