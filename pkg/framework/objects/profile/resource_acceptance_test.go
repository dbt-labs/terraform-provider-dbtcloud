package profile_test

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

func TestAccDbtCloudProfileResource(t *testing.T) {
	projectName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	profileKey := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	profileKeyUpdated := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudProfileDestroy,
		Steps: []resource.TestStep{
			// CREATE
			{
				Config: testAccDbtCloudProfileResourceConfig(projectName, profileKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudProfileExists(
						"dbtcloud_profile.test_profile",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_profile.test_profile",
						"profile_id",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_profile.test_profile",
						"project_id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_profile.test_profile",
						"key",
						profileKey,
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_profile.test_profile",
						"connection_id",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_profile.test_profile",
						"credentials_id",
					),
				),
			},
			// UPDATE
			{
				Config: testAccDbtCloudProfileResourceConfig(projectName, profileKeyUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudProfileExists(
						"dbtcloud_profile.test_profile",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_profile.test_profile",
						"key",
						profileKeyUpdated,
					),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_profile.test_profile",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDbtCloudProfileResourceConfig(projectName, profileKey string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_global_connection" "test_connection" {
  name = "profile_test_connection_%s"

  snowflake = {
    account   = "test-account"
    warehouse = "test-warehouse"
    database  = "test-database"
  }
}

resource "dbtcloud_snowflake_credential" "test_credential" {
  is_active  = true
  project_id = dbtcloud_project.test_project.id
  auth_type  = "password"
  database   = "test-database"
  role       = "test-role"
  warehouse  = "test-warehouse"
  schema     = "test-schema"
  user       = "test-user"
  password   = "test-password"
  num_threads = 3
}

resource "dbtcloud_profile" "test_profile" {
  project_id     = dbtcloud_project.test_project.id
  key            = "%s"
  connection_id  = dbtcloud_global_connection.test_connection.id
  credentials_id = dbtcloud_snowflake_credential.test_credential.credential_id
}
`, projectName, projectName, profileKey)
}

func testAccCheckDbtCloudProfileExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}

		projectID, profileID, err := helper.SplitIDToInts(
			rs.Primary.ID,
			"dbtcloud_profile",
		)
		if err != nil {
			return err
		}

		_, err = apiClient.GetProfile(projectID, profileID)
		if err != nil {
			return fmt.Errorf(
				"error fetching item with resource %s. %s",
				resourceName,
				err,
			)
		}
		return nil
	}
}

func testAccCheckDbtCloudProfileDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_profile" {
			continue
		}

		projectID, profileID, err := helper.SplitIDToInts(
			rs.Primary.ID,
			"dbtcloud_profile",
		)
		if err != nil {
			return err
		}

		_, err = apiClient.GetProfile(projectID, profileID)
		if err == nil {
			return fmt.Errorf("Profile still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
