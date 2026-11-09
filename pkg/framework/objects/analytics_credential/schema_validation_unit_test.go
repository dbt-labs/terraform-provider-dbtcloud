package analytics_credential_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/testhelpers"
	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Rule: the nested adapter blocks are validated as a whole when the plan is made.
// given: for each adapter, a valid block, a block that misses a required field, and two blocks at once
// when: the plan is made
// then: the valid block plans, the incomplete block fails with a missing-attribute error and two blocks fail on the exactly-one rule.
func TestAnalyticsCredential_NestedSchemaValidatesPlan(t *testing.T) {
	adapters := map[int]string{
		1: "snowflake_v0",
		2: "databricks_v0",
		3: "redshift_v0",
		4: "bigquery_v1",
	}
	missing := regexp.MustCompile(`(?i)missing|required`)
	cases := []struct {
		name        string
		config      func(hostURL string) string
		expectError *regexp.Regexp
	}{
		{
			"snowflake valid",
			func(u string) string { return snowflakeConfig(u, 1, passwordAttributes) },
			nil,
		},
		{
			"snowflake without the password",
			func(u string) string {
				return snowflakeConfig(u, 1, "auth_type = \"password\"\n    user = \"analytics_user\"\n    warehouse = \"analytics_wh\"")
			},
			missing,
		},
		{
			"databricks valid",
			func(u string) string { return databricksConfig(u, 2, databricksAttributes) },
			nil,
		},
		{
			"databricks without the token",
			func(u string) string { return databricksConfig(u, 2, "catalog = \"main\"") },
			missing,
		},
		{
			"redshift valid",
			func(u string) string { return redshiftConfig(u, 3, redshiftAttributes) },
			nil,
		},
		{
			"redshift without the username",
			func(u string) string {
				return redshiftConfig(u, 3, "password_wo = \"super-secret-redshift\"\n    password_wo_version = 1")
			},
			missing,
		},
		{
			"bigquery valid",
			func(u string) string { return bigqueryConfig(u, 4, serviceAccountFields) },
			nil,
		},
		{
			"bigquery without the private key",
			func(u string) string {
				return bigqueryConfig(u, 4, "client_email = \"sa@example.iam.gserviceaccount.com\"")
			},
			missing,
		},
		{
			"two adapter blocks",
			func(u string) string {
				return providerBlock(u) + fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "test" {
  connection_id = 1
  snowflake = {
    %s
  }
  databricks = {
    %s
  }
}
`, passwordAttributes, databricksAttributes)
			},
			regexp.MustCompile(`(?i)invalid attribute combination|only one of`),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setupAcc(t)
			api := newFakeAPI(adapters)
			srv := testhelpers.SetupMockServer(t, api.handlers())
			defer srv.Close()

			step := tfresource.TestStep{Config: tc.config(srv.URL), PlanOnly: true}
			if tc.expectError != nil {
				step.ExpectError = tc.expectError
			} else {
				step.ExpectNonEmptyPlan = true
			}
			tfresource.Test(t, unitTestCase([]tfresource.TestStep{step}))
		})
	}
}
