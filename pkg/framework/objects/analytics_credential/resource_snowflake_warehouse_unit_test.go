package analytics_credential_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/testhelpers"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const keypairWithoutWarehouse = `auth_type = "keypair"
    user = "analytics_user"
    private_key_wo = "super-secret-key"
    private_key_wo_version = 1`

// Rule: the Snowflake warehouse is optional, and it is sent only when it is set.
// given: a keypair credential without a warehouse, then the same credential with one
// when: each configuration is applied
// then: the create body has no warehouse and the state has none, and the update body carries the warehouse.
func TestAnalyticsCredential_SnowflakeWarehouseIsOptional(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "snowflake_v0"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	tfresource.Test(t, unitTestCase([]tfresource.TestStep{
		{
			Config: snowflakeConfig(srv.URL, 42, keypairWithoutWarehouse),
			Check: tfresource.ComposeTestCheckFunc(
				func(*terraform.State) error {
					posts := calls(srv, t, "POST", collectionPath)
					if len(posts) != 1 {
						return fmt.Errorf("expected one POST, got %d", len(posts))
					}
					if got, found := configOf(t, posts[0])["warehouse"]; found {
						return fmt.Errorf("expected no warehouse in the POST body, got %v", got)
					}
					return nil
				},
				tfresource.TestCheckNoResourceAttr(resourceAddress, "snowflake.warehouse"),
			),
		},
		{
			Config: snowflakeConfig(srv.URL, 42, keypairWithoutWarehouse+"\n    warehouse = \"analytics_wh\""),
			Check: tfresource.ComposeTestCheckFunc(
				func(*terraform.State) error {
					patches := calls(srv, t, "PATCH", itemPath(credentialID))
					if len(patches) != 1 {
						return fmt.Errorf("expected one PATCH, got %d", len(patches))
					}
					if got := configOf(t, patches[0])["warehouse"]; got != "analytics_wh" {
						return fmt.Errorf("expected the warehouse in the PATCH body, got %v", got)
					}
					return nil
				},
				tfresource.TestCheckResourceAttr(resourceAddress, "snowflake.warehouse", "analytics_wh"),
			),
		},
	}))
}

// Rule: an empty warehouse is rejected, as the API rejects it.
// given: a Snowflake block with warehouse set to an empty string
// when: the plan is made
// then: the plan fails on the length.
func TestAnalyticsCredential_SnowflakeEmptyWarehouseFailsPlan(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "snowflake_v0"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	tfresource.Test(t, unitTestCase([]tfresource.TestStep{{
		Config:      snowflakeConfig(srv.URL, 42, keypairWithoutWarehouse+"\n    warehouse = \"\""),
		PlanOnly:    true,
		ExpectError: regexp.MustCompile(`(?s)Invalid Attribute Value Length.*warehouse`),
	}}))
}

// Rule: an optional field cannot be cleared once set.
// given: a warehouse set in the state
// when: the warehouse is removed from the configuration
// then: the plan fails with the documented error.
func TestAnalyticsCredential_RemovingWarehouseFailsPlan(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + connectionPath(42): connectionHandler(42, "snowflake_v0"),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	withWarehouse := snowflakeModel("password")
	withWarehouse.Warehouse = types.StringValue("ANALYTICS_WH")
	withoutWarehouse := snowflakeModel("password")
	withoutWarehouse.Warehouse = types.StringNull()

	resp := runModifyPlan(t, r,
		buildPlan(t, s, credentialModel(42, withoutWarehouse)),
		buildState(t, s, credentialModel(42, withWarehouse)))

	if !hasErrorSummary(resp, "Optional attribute cannot be removed") {
		t.Fatalf("expected the clearing error, got %v", resp.Diagnostics)
	}
}
