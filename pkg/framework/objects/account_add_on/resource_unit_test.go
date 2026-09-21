package account_add_on_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/testhelpers"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testAccountID = 123

// The account cannot use the product, so turning it on is refused.
var regexpForbidden = regexp.MustCompile("forbidden")

func decodeJSON(r *http.Request, target interface{}) error {
	return json.NewDecoder(r.Body).Decode(target)
}

// fakeAddOns is a stand-in for the add-on endpoints of one account. It holds the
// lifecycle state and the spend limit of each product, so a whole plan/apply cycle
// can run against it.
type fakeAddOns struct {
	mu sync.Mutex
	// products the account can use at all. A product outside this set is left out
	// of the list response, the same way the API hides a product that is turned off.
	available map[string]bool
	state     map[string]*string
	limit     map[string]*int64
	consumed  map[string]bool
}

func newFakeAddOns(available ...string) *fakeAddOns {
	f := &fakeAddOns{
		available: map[string]bool{},
		state:     map[string]*string{},
		limit:     map[string]*int64{},
		consumed:  map[string]bool{},
	}
	for _, product := range available {
		f.available[product] = true
	}
	return f
}

func (f *fakeAddOns) setState(product string, state string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	value := state
	f.state[product] = &value
}

func (f *fakeAddOns) entry(product string) map[string]interface{} {
	var state interface{}
	if current, ok := f.state[product]; ok && current != nil {
		state = *current
	}
	return map[string]interface{}{
		"product":          product,
		"state":            state,
		"can_activate":     true,
		"can_trial":        !f.consumed[product],
		"trial_started_at": nil,
		"trial_ends_at":    nil,
		"trial_consumed":   f.consumed[product],
	}
}

func (f *fakeAddOns) handlers(t *testing.T) map[string]testhelpers.MockEndpointHandler {
	t.Helper()

	basePath := fmt.Sprintf("/private/accounts/%d/add-ons", testAccountID)

	handlers := map[string]testhelpers.MockEndpointHandler{
		"GET " + basePath: func(_ *http.Request) (int, interface{}, error) {
			f.mu.Lock()
			defer f.mu.Unlock()
			entries := []map[string]interface{}{}
			// A stable order keeps the list data source assertions readable.
			for _, product := range []string{"wizard", "state"} {
				if f.available[product] {
					entries = append(entries, f.entry(product))
				}
			}
			return http.StatusOK, map[string]interface{}{
				"data":   map[string]interface{}{"add_ons": entries},
				"status": map[string]interface{}{"code": 200, "is_success": true},
			}, nil
		},
	}

	for _, product := range []string{"wizard", "state"} {
		product := product
		productPath := basePath + "/" + product

		action := func(next string, consumesTrial bool) testhelpers.MockEndpointHandler {
			return func(_ *http.Request) (int, interface{}, error) {
				f.mu.Lock()
				defer f.mu.Unlock()
				if !f.available[product] {
					return http.StatusForbidden, map[string]interface{}{
						"status": map[string]interface{}{
							"code": 403, "is_success": false,
							"user_message": "product unavailable",
						},
					}, nil
				}
				value := next
				f.state[product] = &value
				if consumesTrial {
					f.consumed[product] = true
				}
				return http.StatusOK, map[string]interface{}{
					"data":   f.entry(product),
					"status": map[string]interface{}{"code": 200, "is_success": true},
				}, nil
			}
		}

		handlers["POST "+productPath+"/activate"] = action("ACTIVE", false)
		handlers["POST "+productPath+"/start-trial"] = action("TRIAL", true)
		handlers["POST "+productPath+"/cancel"] = action("CANCELLED", false)

		handlers["GET "+productPath+"/spend-limit"] = func(_ *http.Request) (int, interface{}, error) {
			f.mu.Lock()
			defer f.mu.Unlock()
			// The API has nothing to report without a live add-on.
			if current, ok := f.state[product]; !ok || current == nil {
				return http.StatusNotFound, map[string]interface{}{
					"status": map[string]interface{}{
						"code": 404, "is_success": false,
						"user_message": "no live add-on",
					},
				}, nil
			}
			var limit interface{}
			if f.limit[product] != nil {
				limit = *f.limit[product]
			}
			return http.StatusOK, map[string]interface{}{
				"data": map[string]interface{}{
					"product": product, "spend_limit_nanodollars": limit,
				},
				"status": map[string]interface{}{"code": 200, "is_success": true},
			}, nil
		}

		handlers["PUT "+productPath+"/spend-limit"] = func(r *http.Request) (int, interface{}, error) {
			f.mu.Lock()
			defer f.mu.Unlock()
			body := map[string]interface{}{}
			if err := decodeJSON(r, &body); err != nil {
				return http.StatusBadRequest, nil, err
			}
			raw, present := body["spend_limit_nanodollars"]
			if !present {
				t.Errorf("PUT spend-limit for %s omitted spend_limit_nanodollars", product)
			}
			if raw == nil {
				f.limit[product] = nil
			} else {
				limit := int64(raw.(float64))
				f.limit[product] = &limit
			}
			var echo interface{}
			if f.limit[product] != nil {
				echo = *f.limit[product]
			}
			return http.StatusOK, map[string]interface{}{
				"data": map[string]interface{}{
					"product": product, "spend_limit_nanodollars": echo,
				},
				"status": map[string]interface{}{"code": 200, "is_success": true},
			}, nil
		}
	}

	return handlers
}

func providerConfig(url string) string {
	return fmt.Sprintf(`
provider "dbtcloud" {
  host_url   = "%s"
  token      = "test-token"
  account_id = %d
}
`, url, testAccountID)
}

// TestAccountAddOnResource_PaidLifecycle covers the whole life of a paid add-on:
// activation, then setting, changing and removing the spend limit. Every step also
// asserts that the plan settles, which is what catches a value that does not round
// trip through the API.
func TestAccountAddOnResource_PaidLifecycle(t *testing.T) {
	fake := newFakeAddOns("wizard", "state")
	server := testhelpers.SetupMockServer(t, fake.handlers(t))
	defer server.Close()

	config := func(extra string) string {
		return providerConfig(server.URL) + fmt.Sprintf(`
resource "dbtcloud_account_add_on" "test" {
  product = "wizard"
  %s
}
`, extra)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config(""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dbtcloud_account_add_on.test", "state", "ACTIVE"),
					resource.TestCheckResourceAttr("dbtcloud_account_add_on.test", "activation", "paid"),
					resource.TestCheckResourceAttr("dbtcloud_account_add_on.test", "id", "wizard"),
					resource.TestCheckNoResourceAttr("dbtcloud_account_add_on.test", "spend_limit_nanodollars"),
				),
			},
			{
				Config: config(`spend_limit_nanodollars = 500000000000`),
				Check: resource.TestCheckResourceAttr(
					"dbtcloud_account_add_on.test", "spend_limit_nanodollars", "500000000000",
				),
			},
			{
				Config: config(`spend_limit_nanodollars = 250000000000`),
				Check: resource.TestCheckResourceAttr(
					"dbtcloud_account_add_on.test", "spend_limit_nanodollars", "250000000000",
				),
			},
			// 0 is a real limit, and has to stay distinct from no limit at all.
			{
				Config: config(`spend_limit_nanodollars = 0`),
				Check: resource.TestCheckResourceAttr(
					"dbtcloud_account_add_on.test", "spend_limit_nanodollars", "0",
				),
			},
			{
				Config: config(""),
				Check: resource.TestCheckNoResourceAttr(
					"dbtcloud_account_add_on.test", "spend_limit_nanodollars",
				),
			},
		},
	})
}

// TestAccountAddOnResource_Trial checks the trial activation path, which uses a
// different endpoint from paid activation.
func TestAccountAddOnResource_Trial(t *testing.T) {
	fake := newFakeAddOns("wizard", "state")
	server := testhelpers.SetupMockServer(t, fake.handlers(t))
	defer server.Close()

	config := providerConfig(server.URL) + `
resource "dbtcloud_account_add_on" "test" {
  product    = "state"
  activation = "trial"
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dbtcloud_account_add_on.test", "state", "TRIAL"),
					resource.TestCheckResourceAttr("dbtcloud_account_add_on.test", "trial_consumed", "true"),
					resource.TestCheckResourceAttr("dbtcloud_account_add_on.test", "can_trial", "false"),
				),
			},
		},
	})
}

// TestAccountAddOnResource_ReadsEveryState checks that each lifecycle state the API
// can report maps into state, and that none of them produces a plan. A trial moves
// to EXPIRED on its own, so a state the configuration never asked for must not read
// as drift.
func TestAccountAddOnResource_ReadsEveryState(t *testing.T) {
	for _, product := range []string{"wizard", "state"} {
		for _, addOnState := range []string{"TRIAL", "ACTIVE", "CANCELLED", "EXPIRED"} {
			t.Run(product+"_"+addOnState, func(t *testing.T) {
				fake := newFakeAddOns("wizard", "state")
				server := testhelpers.SetupMockServer(t, fake.handlers(t))
				defer server.Close()

				config := providerConfig(server.URL) + fmt.Sprintf(`
resource "dbtcloud_account_add_on" "test" {
  product = "%s"
}
`, product)

				resource.Test(t, resource.TestCase{
					ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
					Steps: []resource.TestStep{
						{
							Config: config,
							Check: resource.TestCheckResourceAttr(
								"dbtcloud_account_add_on.test", "state", "ACTIVE",
							),
						},
						{
							// The product moves on its own, and the next plan has to be empty.
							PreConfig: func() { fake.setState(product, addOnState) },
							Config:    config,
							Check: resource.TestCheckResourceAttr(
								"dbtcloud_account_add_on.test", "state", addOnState,
							),
						},
					},
				})
			})
		}
	}
}

// TestAccountAddOnResource_UnavailableProduct checks the error for a product the
// account cannot use. The API leaves such a product out of the list, which must not
// read as a product the account simply never turned on.
func TestAccountAddOnResource_UnavailableProduct(t *testing.T) {
	fake := newFakeAddOns("state")
	server := testhelpers.SetupMockServer(t, fake.handlers(t))
	defer server.Close()

	config := providerConfig(server.URL) + `
resource "dbtcloud_account_add_on" "test" {
  product = "wizard"
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexpForbidden,
			},
		},
	})
}

// TestAccountAddOnDataSources checks both data sources against the same account.
func TestAccountAddOnDataSources(t *testing.T) {
	fake := newFakeAddOns("wizard", "state")
	fake.setState("wizard", "ACTIVE")
	fake.setState("state", "TRIAL")
	server := testhelpers.SetupMockServer(t, fake.handlers(t))
	defer server.Close()

	config := providerConfig(server.URL) + `
data "dbtcloud_account_add_on" "one" {
  product = "wizard"
}

data "dbtcloud_account_add_ons" "all" {}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dbtcloud_account_add_on.one", "state", "ACTIVE"),
					resource.TestCheckResourceAttr("data.dbtcloud_account_add_on.one", "product", "wizard"),
					resource.TestCheckResourceAttr("data.dbtcloud_account_add_ons.all", "add_ons.#", "2"),
					resource.TestCheckResourceAttr("data.dbtcloud_account_add_ons.all", "add_ons.0.product", "wizard"),
					resource.TestCheckResourceAttr("data.dbtcloud_account_add_ons.all", "add_ons.0.state", "ACTIVE"),
					resource.TestCheckResourceAttr("data.dbtcloud_account_add_ons.all", "add_ons.1.product", "state"),
					resource.TestCheckResourceAttr("data.dbtcloud_account_add_ons.all", "add_ons.1.state", "TRIAL"),
				),
			},
		},
	})
}
