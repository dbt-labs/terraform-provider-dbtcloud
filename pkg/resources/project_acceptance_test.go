package resources_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("DBT_CLOUD_ACCOUNT_ID"); v == "" {
		t.Fatal("DBT_CLOUD_ACCOUNT_ID must be set for acceptance tests")
	}
	if v := os.Getenv("DBT_CLOUD_TOKEN"); v == "" {
		t.Fatal("DBT_CLOUD_TOKEN must be set for acceptance tests")
	}
}

func TestAccDbtCloudProjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudProjectResourceBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudProjectExists("dbt_cloud_project.test_project"),
					resource.TestCheckResourceAttr("dbt_cloud_project.test_project", "name", "TEST"),
				),
			},
		},
	})
}

func TestAccDbtCloudProjectResource_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudProjectResourceBefore(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudProjectExists("dbt_cloud_project.test_project"),
					resource.TestCheckResourceAttr("dbt_cloud_project.test_project", "name", "TEST_BEFORE"),
				),
			},
			{
				Config: testAccDbtCloudProjectResourceAfter(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudProjectExists("dbt_cloud_project.test_project"),
					resource.TestCheckResourceAttr("dbt_cloud_project.test_project", "name", "TEST_AFTER"),
				),
			},
		},
	})
}

func testAccDbtCloudProjectResourceBasic() string {
	return fmt.Sprintf(`
resource "dbt_cloud_project" "test_project" {
  name        = "TEST"
}
`)
}

func testAccDbtCloudProjectResourceBefore() string {
	return fmt.Sprintf(`
resource "dbt_cloud_project" "test_project" {
  name        = "TEST_BEFORE"
}
`)
}

func testAccDbtCloudProjectResourceAfter() string {
	return fmt.Sprintf(`
resource "dbt_cloud_project" "test_project" {
  name        = "TEST_AFTER"
}
`)
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
		projectID := rs.Primary.ID
		apiClient := testAccProvider.Meta().(*dbt_cloud.Client)
		_, err := apiClient.GetProject(projectID)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudProjectDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbt_cloud_project" {
			continue
		}

		_, err := apiClient.GetProject(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Alert still exists")
		}
		notFoundErr := "not found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
