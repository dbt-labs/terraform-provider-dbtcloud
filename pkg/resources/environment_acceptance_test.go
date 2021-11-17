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

func TestAccDbtCloudEnvironmentResource(t *testing.T) {

	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudEnvironmentResourceBasicConfig(projectName, environmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbt_cloud_environment.test_env"),
					resource.TestCheckResourceAttr("dbt_cloud_environment.test_env", "name", environmentName),
				),
			},
			// RENAME
			{
				Config: testAccDbtCloudEnvironmentResourceBasicConfig(projectName, environmentName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbt_cloud_environment.test_env"),
					resource.TestCheckResourceAttr("dbt_cloud_environment.test_env", "name", environmentName2),
				),
			},
			// 			// MODIFY
			// 			{
			// 				Config: testAccDbtCloudProjectResourceFullConfig(projectName2),
			// 				Check: resource.ComposeTestCheckFunc(
			// 					testAccCheckDbtCloudProjectExists("dbt_cloud_project.test_project"),
			// 					resource.TestCheckResourceAttr("dbt_cloud_project.test_project", "name", projectName2),
			// 					resource.TestCheckResourceAttr("dbt_cloud_project.test_project", "dbt_project_subdirectory", "/project/subdirectory_where/dbt-is"),
			// 				),
			// 			},
			// IMPORT
			{
				ResourceName:            "dbt_cloud_environment.test_env",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccDbtCloudEnvironmentResourceBasicConfig(projectName, environmentName string) string {
	return fmt.Sprintf(`
resource "dbt_cloud_project" "test_project" {
  name        = "%s"
}

resource "dbt_cloud_environment" "test_env" {
  name        = "%s"
  type = "deployment"
  dbt_version = "0.21.0"
  project_id = dbt_cloud_project.test_project.id
}
`, projectName, environmentName)
}

func testAccCheckDbtCloudEnvironmentExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		apiClient := testAccProvider.Meta().(*dbt_cloud.Client)
		projectId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0])
		if err != nil {
			return fmt.Errorf("Can't get projectId")
		}

		environmentId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1])
		if err != nil {
			return fmt.Errorf("Can't get environmentId")
		}

		_, err = apiClient.GetEnvironment(projectId, environmentId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudEnvironmentDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbt_cloud_environment" {
			continue
		}
		projectId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0])
		if err != nil {
			return fmt.Errorf("Can't get projectId")
		}

		environmentId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1])
		if err != nil {
			return fmt.Errorf("Can't get environmentId")
		}
		_, err = apiClient.GetEnvironment(projectId, environmentId)
		if err == nil {
			return fmt.Errorf("Environment still exists")
		}
		notFoundErr := "not found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
