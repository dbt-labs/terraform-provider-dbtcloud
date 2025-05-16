package starburst_credential_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudStarburstCredentialResource(t *testing.T) {
	projectName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	database := "test_catalog"
	schema := "test_schema"
	user := "test_user"
	password := "test_password"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudStarburstCredentialDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDbtCloudStarburstCredentialResourceConfig(
					projectName,
					database,
					schema,
					user,
					password,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDbtCloudStarburstCredentialExists("dbtcloud_starburst_credential.test"),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_starburst_credential.test",
						"id",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_starburst_credential.test",
						"credential_id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_starburst_credential.test",
						"database",
						database,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_starburst_credential.test",
						"schema",
						schema,
					),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dbtcloud_starburst_credential.test",
				ImportState:       true,
				ImportStateVerify: true,
				// These fields can't be read from the API
				ImportStateVerifyIgnore: []string{
					"user",
					"password",
				},
			},
			// Update and Read testing
			{
				Config: testAccDbtCloudStarburstCredentialResourceConfig(
					projectName,
					database,
					"updated_schema",
					user,
					password,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDbtCloudStarburstCredentialExists("dbtcloud_starburst_credential.test"),
					resource.TestCheckResourceAttr(
						"dbtcloud_starburst_credential.test",
						"schema",
						"updated_schema",
					),
				),
			},
		},
	})
}

func testAccDbtCloudStarburstCredentialResourceConfig(
	projectName string,
	database string,
	schema string,
	user string,
	password string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test" {
  name = "%s"
}

resource "dbtcloud_starburst_credential" "test" {
  project_id           = dbtcloud_project.test.id
  database             = "%s"
  schema               = "%s"
  user                 = "%s"
  password             = "%s"
}
`, projectName, database, schema, user, password)
}

func testAccCheckDbtCloudStarburstCredentialExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		projectID, credentialID, err := helper.SplitIDToInts(rs.Primary.ID, "dbtcloud_starburst_credential")
		if err != nil {
			return err
		}

		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
		_, err = apiClient.GetStarburstCredential(projectID, credentialID)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudStarburstCredentialDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_starburst_credential" {
			continue
		}
		projectID, credentialID, err := helper.SplitIDToInts(rs.Primary.ID, "dbtcloud_starburst_credential")
		if err != nil {
			return err
		}

		_, err = apiClient.GetStarburstCredential(projectID, credentialID)
		if err == nil {
			return fmt.Errorf("Starburst credential still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
