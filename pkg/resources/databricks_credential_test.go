package resources_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDbtCloudDatabricksCredentialResource(t *testing.T) {

    testDatabricks, exists := os.LookupEnv("TEST_DATABRICKS")
    if (exists && testDatabricks == true){
    projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	targetName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	token := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudDatabricksCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudDatabricksCredentialResourceBasicConfig(projectName, targetName, token),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudDatabricksCredentialExists("dbt_cloud_databricks_credential.test_credential"),
					resource.TestCheckResourceAttr("dbt_cloud_databricks_credential.test_credential", "target_name", targetName),
				),
			},
			// RENAME
			// MODIFY
			// IMPORT
			{
				ResourceName:            "dbt_cloud_databricks_credential.test_credential",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
    }
}

func testAccDbtCloudDatabricksCredentialResourceBasicConfig(projectName, targetName, token string) string {
	return fmt.Sprintf(`
resource "dbt_cloud_project" "test_project" {
  name        = "%s"
}
resource "dbt_cloud_databricks_credential" "test_credential" {
    project_id = dbt_cloud_project.test_project.id
    adapter_id = 123
    target_name = "%s"
    token = "%s"
    num_threads = 3
}
`, projectName, targetName, token)
}

func testAccCheckDbtCloudDatabricksCredentialExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		projectId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0])
		if err != nil {
			return fmt.Errorf("Can't get projectId")
		}
		credentialId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1])
		if err != nil {
			return fmt.Errorf("Can't get projectId")
		}

		apiClient := testAccProvider.Meta().(*dbt_cloud.Client)
		_, err = apiClient.GetDatabricksCredential(projectId, credentialId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudDatabricksCredentialDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbt_cloud_databricks_credential" {
			continue
		}
		projectId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0])
		if err != nil {
			return fmt.Errorf("Can't get projectId")
		}
		credentialId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1])
		if err != nil {
			return fmt.Errorf("Can't get credentialId")
		}

		_, err = apiClient.GetDatabricksCredential(projectId, credentialId)
		if err == nil {
			return fmt.Errorf("Databricks credential still exists")
		}
		notFoundErr := "Did not find"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
