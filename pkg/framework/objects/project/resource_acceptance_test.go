package project_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudProjectResource(t *testing.T) {
	projectName := "testproject1"
	projectDescription := "testprojectdescription"
	// projectName2 := "testproject2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudProjectDestroy,
		Steps: []resource.TestStep{
			// Initial creation
			{
				Config: testAccDbtCloudProjectResourceBasicConfig(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudProjectExists("dbtcloud_project.test_project"),
					resource.TestCheckResourceAttr(
						"dbtcloud_project.test_project",
						"name",
						projectName,
					),
				),
			},
			// RENAME
			{
				Config: testAccDbtCloudProjectResourceBasicConfig(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudProjectExists("dbtcloud_project.test_project"),
					resource.TestCheckResourceAttr(
						"dbtcloud_project.test_project",
						"name",
						projectName,
					),
				),
			},
			// MODIFY
			{
				Config: testAccDbtCloudProjectResourceFullConfig(projectName, projectDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudProjectExists("dbtcloud_project.test_project"),
					resource.TestCheckResourceAttr(
						"dbtcloud_project.test_project",
						"name",
						projectName,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_project.test_project",
						"dbt_project_subdirectory",
						"/project/subdirectory_where/dbt-is",
					),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_project.test_project",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccDbtCloudProjectResourceBasicConfig(projectName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}
`, projectName)
}

func testAccDbtCloudProjectResourceFullConfig(
	projectName string,
	projectDescription string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
  description = "%s"
  dbt_project_subdirectory = "/project/subdirectory_where/dbt-is"
}
`, projectName, projectDescription)
}

func testAccCheckDbtCloudProjectExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
		_, err = apiClient.GetProject(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudProjectDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_project" {
			continue
		}
		_, err := apiClient.GetProject(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Project still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
