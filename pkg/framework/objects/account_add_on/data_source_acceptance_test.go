package account_add_on_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/testhelpers"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testAccountID = 123

// addOn describes one product of the fake account.
type addOn struct {
	// state is empty when the account has never turned the product on.
	state string
	// limit is negative when no spend limit is set.
	limit int64
}

// handlers serves the add-on endpoints of one account. Only the GET endpoints are
// registered: the mock answers 404 for anything else, so a write from the provider
// fails the test. These are data sources, and they must never write.
func handlers(products map[string]addOn) map[string]testhelpers.MockEndpointHandler {
	basePath := fmt.Sprintf("/private/accounts/%d/add-ons", testAccountID)

	entry := func(product string, details addOn) map[string]interface{} {
		var state interface{}
		if details.state != "" {
			state = details.state
		}
		return map[string]interface{}{
			"product":          product,
			"state":            state,
			"can_activate":     details.state == "",
			"can_trial":        details.state == "",
			"trial_started_at": nil,
			"trial_ends_at":    nil,
			"trial_consumed":   details.state != "",
		}
	}

	result := map[string]testhelpers.MockEndpointHandler{
		"GET " + basePath: func(_ *http.Request) (int, interface{}, error) {
			entries := []map[string]interface{}{}
			// A stable order keeps the list assertions readable.
			for _, product := range []string{"wizard", "state"} {
				if details, ok := products[product]; ok {
					entries = append(entries, entry(product, details))
				}
			}
			return http.StatusOK, map[string]interface{}{
				"data":   map[string]interface{}{"add_ons": entries},
				"status": map[string]interface{}{"code": 200, "is_success": true},
			}, nil
		},
	}

	for product, details := range products {
		product, details := product, details
		result["GET "+basePath+"/"+product+"/spend-limit"] = func(_ *http.Request) (int, interface{}, error) {
			// The API has nothing to report without a live add-on.
			if details.state == "" {
				return http.StatusNotFound, map[string]interface{}{
					"status": map[string]interface{}{
						"code": 404, "is_success": false, "user_message": "no live add-on",
					},
				}, nil
			}
			var limit interface{}
			if details.limit >= 0 {
				limit = details.limit
			}
			return http.StatusOK, map[string]interface{}{
				"data": map[string]interface{}{
					"product": product, "spend_limit_nanodollars": limit,
				},
				"status": map[string]interface{}{"code": 200, "is_success": true},
			}, nil
		}
	}

	return result
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

// TestAccountAddOnDataSource_EveryState checks that each lifecycle state the API
// can report is mapped, for both products. A product the account has never turned
// on reports no state and no spend limit, rather than an error.
func TestAccountAddOnDataSource_EveryState(t *testing.T) {
	for _, product := range []string{"wizard", "state"} {
		for _, addOnState := range []string{"", "TRIAL", "ACTIVE", "CANCELLED", "EXPIRED"} {
			name := addOnState
			if name == "" {
				name = "never_turned_on"
			}
			t.Run(product+"_"+name, func(t *testing.T) {
				server := testhelpers.SetupMockServer(t, handlers(map[string]addOn{
					"wizard": {state: addOnState, limit: -1},
					"state":  {state: addOnState, limit: -1},
				}))
				defer server.Close()

				check := resource.TestCheckResourceAttr(
					"data.dbtcloud_account_add_on.test", "state", addOnState,
				)
				if addOnState == "" {
					check = resource.TestCheckNoResourceAttr(
						"data.dbtcloud_account_add_on.test", "state",
					)
				}

				resource.Test(t, resource.TestCase{
					ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
					Steps: []resource.TestStep{
						{
							Config: providerConfig(server.URL) + fmt.Sprintf(`
data "dbtcloud_account_add_on" "test" {
  product = "%s"
}
`, product),
							Check: resource.ComposeAggregateTestCheckFunc(
								check,
								resource.TestCheckResourceAttr(
									"data.dbtcloud_account_add_on.test", "product", product,
								),
								resource.TestCheckResourceAttr(
									"data.dbtcloud_account_add_on.test", "id", product,
								),
							),
						},
					},
				})
			})
		}
	}
}

// TestAccountAddOnDataSource_SpendLimit checks the spend limit, including the two
// values that are easy to confuse: a limit of 0, and no limit at all.
func TestAccountAddOnDataSource_SpendLimit(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		limit int64
		want  string
	}{
		{"a limit", 500000000000, "500000000000"},
		{"a limit of zero", 0, "0"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			server := testhelpers.SetupMockServer(t, handlers(map[string]addOn{
				"wizard": {state: "ACTIVE", limit: testCase.limit},
			}))
			defer server.Close()

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: providerConfig(server.URL) + `
data "dbtcloud_account_add_on" "test" {
  product = "wizard"
}
`,
						Check: resource.TestCheckResourceAttr(
							"data.dbtcloud_account_add_on.test",
							"spend_limit_nanodollars",
							testCase.want,
						),
					},
				},
			})
		})
	}

	t.Run("no limit set", func(t *testing.T) {
		server := testhelpers.SetupMockServer(t, handlers(map[string]addOn{
			"wizard": {state: "ACTIVE", limit: -1},
		}))
		defer server.Close()

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: providerConfig(server.URL) + `
data "dbtcloud_account_add_on" "test" {
  product = "wizard"
}
`,
					Check: resource.TestCheckNoResourceAttr(
						"data.dbtcloud_account_add_on.test", "spend_limit_nanodollars",
					),
				},
			},
		})
	})
}

// TestAccountAddOnDataSource_UnavailableProduct checks the error for a product the
// account cannot use. The API leaves such a product out of the list, which must not
// read as a product the account has simply never turned on.
func TestAccountAddOnDataSource_UnavailableProduct(t *testing.T) {
	server := testhelpers.SetupMockServer(t, handlers(map[string]addOn{
		"state": {state: "ACTIVE", limit: -1},
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig(server.URL) + `
data "dbtcloud_account_add_on" "test" {
  product = "wizard"
}
`,
				ExpectError: regexp.MustCompile("Add-on product not available"),
			},
		},
	})
}

// TestAccountAddOnsDataSource checks the list, and that a product the account
// cannot use is left out of it.
func TestAccountAddOnsDataSource(t *testing.T) {
	server := testhelpers.SetupMockServer(t, handlers(map[string]addOn{
		"wizard": {state: "ACTIVE", limit: 500000000000},
		"state":  {state: "TRIAL", limit: -1},
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig(server.URL) + `
data "dbtcloud_account_add_ons" "all" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dbtcloud_account_add_ons.all", "add_ons.#", "2"),
					resource.TestCheckResourceAttr("data.dbtcloud_account_add_ons.all", "add_ons.0.product", "wizard"),
					resource.TestCheckResourceAttr("data.dbtcloud_account_add_ons.all", "add_ons.0.state", "ACTIVE"),
					resource.TestCheckResourceAttr("data.dbtcloud_account_add_ons.all", "add_ons.0.spend_limit_nanodollars", "500000000000"),
					resource.TestCheckResourceAttr("data.dbtcloud_account_add_ons.all", "add_ons.1.product", "state"),
					resource.TestCheckResourceAttr("data.dbtcloud_account_add_ons.all", "add_ons.1.state", "TRIAL"),
					resource.TestCheckNoResourceAttr("data.dbtcloud_account_add_ons.all", "add_ons.1.spend_limit_nanodollars"),
				),
			},
		},
	})
}

// TestAccountAddOnsDataSource_OnlyAvailableProducts checks that the list carries
// only what the account can use.
func TestAccountAddOnsDataSource_OnlyAvailableProducts(t *testing.T) {
	server := testhelpers.SetupMockServer(t, handlers(map[string]addOn{
		"state": {state: "ACTIVE", limit: -1},
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig(server.URL) + `
data "dbtcloud_account_add_ons" "all" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dbtcloud_account_add_ons.all", "add_ons.#", "1"),
					resource.TestCheckResourceAttr("data.dbtcloud_account_add_ons.all", "add_ons.0.product", "state"),
				),
			},
		},
	})
}
