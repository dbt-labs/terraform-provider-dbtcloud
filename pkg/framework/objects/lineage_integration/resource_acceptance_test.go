package lineage_integration_test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudLineageIntegrationResource(t *testing.T) {

	envVarLineageIntegration, exists := os.LookupEnv("DBT_ACCEPTANCE_TEST_LINEAGE_INTEGRATION")

	if !exists {
		t.Skip(
			"Skipping lineage configuration acceptance tests as the env var DBT_ACCEPTANCE_TEST_LINEAGE_INTEGRATION is not set",
		)
	}

	lineageIntegrationConfigs := strings.Split(envVarLineageIntegration, "~")
	if len(lineageIntegrationConfigs) != 4 {
		t.Fatalf(
			"DBT_ACCEPTANCE_TEST_LINEAGE_INTEGRATION env var should be in the format: host~side_id~token_name~token",
		)
	}

	lineageIntegrationHost := lineageIntegrationConfigs[0]
	lineageIntegrationSiteID := lineageIntegrationConfigs[1]
	lineageIntegrationTokenName := lineageIntegrationConfigs[2]
	lineageIntegrationToken := lineageIntegrationConfigs[3]

	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudLineageIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudLineageIntegrationResourceBasicConfig(
					projectName,
					lineageIntegrationHost,
					lineageIntegrationSiteID,
					lineageIntegrationTokenName,
					lineageIntegrationToken,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_lineage_integration.my_lineage",
						"id",
					),
				),
			},
			// MODIFY
			// IMPORT
			{
				ResourceName:            "dbtcloud_lineage_integration.my_lineage",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func testAccDbtCloudLineageIntegrationResourceBasicConfig(
	projectName, host, siteID, tokenName, token string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_lineage_integration" {
	name = "%s"
}

resource "dbtcloud_snowflake_credential" "my_cred" {
  project_id  = dbtcloud_project.test_lineage_integration.id
  auth_type   = "password"
  num_threads = 16
  schema      = "SCHEMA"
  user        = "user"
  password    = "password"
}

resource "dbtcloud_global_connection" "my_connection" {
name = "terraform_snowflake_testing_proj_qa"
	snowflake = {
		account   = "NA"
		database  = "DB"
		warehouse = "WH"
	}
}

resource dbtcloud_environment my_env {
  dbt_version     = "versionless"
  name            = "Prod"
  project_id      = dbtcloud_project.test_lineage_integration.id
  type            = "deployment"
  credential_id   = dbtcloud_snowflake_credential.my_cred.credential_id
  deployment_type = "production"
  connection_id   = dbtcloud_global_connection.my_connection.id
}

resource dbtcloud_lineage_integration my_lineage {
  project_id = dbtcloud_project.test_lineage_integration.id
  host = "%s"
  site_id = "%s"
  token_name = "%s"
  token = "%s"

  depends_on = [dbtcloud_environment.my_env]
}
`, projectName, host, siteID, tokenName, token)
}

func testAccCheckDbtCloudLineageIntegrationDestroy(s *terraform.State) error {

	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_lineage_integration" {
			continue
		}

		projectID, lineageID, err := helper.SplitIDToInts(rs.Primary.ID, "lineage_integration")
		if err != nil {
			return fmt.Errorf("Error splitting ID: %s", err)
		}

		_, err = apiClient.GetLineageIntegration(int64(projectID), int64(lineageID))
		if err == nil {
			return fmt.Errorf("Lineage integration still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
