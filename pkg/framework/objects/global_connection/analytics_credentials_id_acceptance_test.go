package global_connection_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// These tests create no analytics credential, so they run without the analytics credentials feature.
// The case of a connection that has one is covered with the credential resource.

var analyticsCredentialsIDPath = tfjsonpath.New("analytics_credentials_id")

// adapterConnection returns the configuration of a connection of one adapter, with the data source that
// reads it. The variant changes one non-secret attribute so that a second step updates the connection in place.
type adapterConnection func(name, variant string) string

func connectionWithDataSource(block string) string {
	return block + `
data dbtcloud_global_connection test {
  id = dbtcloud_global_connection.test.id
}
`
}

func analyticsSnowflakeConnection(name, variant string) string {
	return connectionWithDataSource(fmt.Sprintf(`
resource dbtcloud_global_connection test {
  name = "%s"

  snowflake = {
    account = "account"
    warehouse = "warehouse%s"
    database = "database"
    allow_sso = false
    client_session_keep_alive = false
  }
}
`, name, variant))
}

func analyticsDatabricksConnection(name, variant string) string {
	return connectionWithDataSource(fmt.Sprintf(`
resource dbtcloud_global_connection test {
  name = "%s"

  databricks = {
    host = "databricks.com"
    http_path = "/sql/your/http/path"
    catalog = "dbt_catalog%s"
  }
}
`, name, variant))
}

func analyticsRedshiftConnection(name, variant string) string {
	return connectionWithDataSource(fmt.Sprintf(`
resource dbtcloud_global_connection test {
  name = "%s"

  redshift = {
    hostname = "test.com"
    port = 9876
    dbname = "my_database%s"
  }
}
`, name, variant))
}

// Rule: a connection without an analytics credential reports a null analytics_credentials_id, and the value stays null.
// given: a connection of each adapter and its data source
// when: it is created, then updated in place, then imported
// then: the attribute is null in the resource and the data source after create, it is planned as null (not unknown)
// on the update, the plan after the apply is empty, and the import shows no drift.
func TestAccDbtCloudGlobalConnectionAnalyticsCredentialsIDIsNull(t *testing.T) {
	cases := map[string]adapterConnection{
		"snowflake":  analyticsSnowflakeConnection,
		"databricks": analyticsDatabricksConnection,
		"redshift":   analyticsRedshiftConnection,
	}

	for adapter, config := range cases {
		t.Run(adapter, func(t *testing.T) {
			name := acctest.RandStringFromCharSet(12, acctest.CharSetAlpha)

			isNull := []statecheck.StateCheck{
				statecheck.ExpectKnownValue("dbtcloud_global_connection.test", analyticsCredentialsIDPath, knownvalue.Null()),
				statecheck.ExpectKnownValue("data.dbtcloud_global_connection.test", analyticsCredentialsIDPath, knownvalue.Null()),
			}

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:            config(name, ""),
						ConfigStateChecks: isNull,
					},
					{
						Config: config(name, "2"),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectResourceAction("dbtcloud_global_connection.test", plancheck.ResourceActionUpdate),
								plancheck.ExpectKnownValue("dbtcloud_global_connection.test", analyticsCredentialsIDPath, knownvalue.Null()),
							},
							PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
						},
						ConfigStateChecks: isNull,
					},
					{
						ResourceName:      "dbtcloud_global_connection.test",
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		})
	}
}

// Rule: the connections list reports a null analytics_credentials_id for a connection without a credential.
// given: a connection without an analytics credential
// when: the list data source reads the account's connections
// then: the connection's entry has no analytics_credentials_id.
func TestAccDbtCloudGlobalConnectionsDatasourceAnalyticsCredentialsIDIsNull(t *testing.T) {
	name := acctest.RandStringFromCharSet(12, acctest.CharSetAlpha)

	// Unparallelized because of flakiness, as the other tests of the list data source
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`
resource dbtcloud_global_connection test {
  name = "%s"

  redshift = {
    hostname = "test.com"
    port = 9876
    dbname = "my_database"
  }
}

data dbtcloud_global_connections all {
  depends_on = [dbtcloud_global_connection.test]
}
`, name),
			Check: connectionEntryHasNoAnalyticsCredentialsID("data.dbtcloud_global_connections.all", name),
		}},
	})
}

// connectionEntryHasNoAnalyticsCredentialsID finds the entry of the connection by name, as the list holds
// the other connections of the account, and checks that it has no analytics_credentials_id.
func connectionEntryHasNoAnalyticsCredentialsID(address, connectionName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		listed, ok := state.RootModule().Resources[address]
		if !ok {
			return fmt.Errorf("%s not found in the state", address)
		}
		for key, value := range listed.Primary.Attributes {
			if !strings.HasPrefix(key, "connections.") || !strings.HasSuffix(key, ".name") || value != connectionName {
				continue
			}
			entry := strings.TrimSuffix(key, ".name")
			if _, found := listed.Primary.Attributes[entry+".id"]; !found {
				return fmt.Errorf("%s has no id, so it is not an entry of the connection list", entry)
			}
			if attribute, found := listed.Primary.Attributes[entry+".analytics_credentials_id"]; found {
				return fmt.Errorf("%s.analytics_credentials_id is %q, expected it to be null", entry, attribute)
			}
			return nil
		}
		return fmt.Errorf("connection %q is not in %s", connectionName, address)
	}
}

// Rule: analytics_credentials_id is read-only.
// given: a configuration that sets analytics_credentials_id
// when: it is planned
// then: the plan fails with the framework's read-only attribute error.
func TestAccDbtCloudGlobalConnectionAnalyticsCredentialsIDCannotBeConfigured(t *testing.T) {
	name := acctest.RandStringFromCharSet(12, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`
resource dbtcloud_global_connection test {
  name = "%s"
  analytics_credentials_id = 1

  redshift = {
    hostname = "test.com"
    port = 9876
    dbname = "my_database"
  }
}
`, name),
			PlanOnly:    true,
			ExpectError: regexp.MustCompile(`(?is)Invalid Configuration for Read-Only Attribute.*analytics_credentials_id`),
		}},
	})
}
