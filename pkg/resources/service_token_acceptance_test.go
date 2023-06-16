package resources_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDbtCloudServiceTokenResource(t *testing.T) {

	serviceTokenName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	serviceTokenName2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	projectName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudServiceTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudServiceTokenResourceBasicConfig(serviceTokenName, projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudServiceTokenExists("dbt_cloud_service_token.test_service_token"),
					resource.TestCheckResourceAttr("dbt_cloud_service_token.test_service_token", "name", serviceTokenName),
					resource.TestCheckResourceAttr("dbt_cloud_service_token.test_service_token", "service_token_permissions.#", "2"),
					resource.TestCheckResourceAttr("dbt_cloud_service_token.test_service_token", "service_token_permissions.1.permission_set", "git_admin"),
					resource.TestCheckResourceAttr("dbt_cloud_service_token.test_service_token", "service_token_permissions.1.all_projects", "true"),
					resource.TestCheckResourceAttr("dbt_cloud_service_token.test_service_token", "service_token_permissions.0.permission_set", "job_admin"),
					resource.TestCheckResourceAttr("dbt_cloud_service_token.test_service_token", "service_token_permissions.0.all_projects", "false"),
					resource.TestCheckResourceAttrSet("dbt_cloud_service_token.test_service_token", "service_token_permissions.0.project_id"),
					resource.TestCheckResourceAttrSet("dbt_cloud_service_token.test_service_token", "token_string"),
				),
			},
			// MODIFY
			{
				Config: testAccDbtCloudServiceTokenResourceFullConfig(serviceTokenName2, projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudServiceTokenExists("dbt_cloud_service_token.test_service_token"),
					resource.TestCheckResourceAttr("dbt_cloud_service_token.test_service_token", "name", serviceTokenName2),
					resource.TestCheckResourceAttr("dbt_cloud_service_token.test_service_token", "service_token_permissions.#", "2"),
					resource.TestCheckResourceAttr("dbt_cloud_service_token.test_service_token", "service_token_permissions.0.permission_set", "git_admin"),
					resource.TestCheckResourceAttr("dbt_cloud_service_token.test_service_token", "service_token_permissions.0.all_projects", "false"),
					resource.TestCheckResourceAttrSet("dbt_cloud_service_token.test_service_token", "service_token_permissions.0.project_id"),
					resource.TestCheckResourceAttr("dbt_cloud_service_token.test_service_token", "service_token_permissions.1.all_projects", "true"),
					resource.TestCheckResourceAttr("dbt_cloud_service_token.test_service_token", "service_token_permissions.1.permission_set", "job_admin"),
					resource.TestCheckResourceAttrSet("dbt_cloud_service_token.test_service_token", "token_string"),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbt_cloud_service_token.test_service_token",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"token_string",
				},
			},
		},
	})
}

func testAccDbtCloudServiceTokenResourceBasicConfig(serviceTokenName, projectName string) string {
	return fmt.Sprintf(`
resource "dbt_cloud_project" "test_project" {
  name        = "%s"
}
resource "dbt_cloud_service_token" "test_service_token" {
    name = "%s"
    service_token_permissions {
        permission_set = "git_admin"
        all_projects = true
    }
    service_token_permissions {
        permission_set = "job_admin"
        all_projects = false
        project_id = dbt_cloud_project.test_project.id
    }
}
`, projectName, serviceTokenName)
}

func testAccDbtCloudServiceTokenResourceFullConfig(serviceTokenName, projectName string) string {
	return fmt.Sprintf(`
resource "dbt_cloud_project" "test_project" {
  name        = "%s"
}
resource "dbt_cloud_service_token" "test_service_token" {
    name = "%s"
    service_token_permissions {
        permission_set = "git_admin"
        all_projects = false
        project_id = dbt_cloud_project.test_project.id
    }
    service_token_permissions {
        permission_set = "job_admin"
        all_projects = true
    }
}
`, projectName, serviceTokenName)
}

func testAccCheckDbtCloudServiceTokenExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		apiClient := testAccProvider.Meta().(*dbt_cloud.Client)
		ServiceTokenID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Can't get ServiceTokenID")
		}
		_, err = apiClient.GetServiceToken(ServiceTokenID)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudServiceTokenDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbt_cloud_service_token" {
			continue
		}
		ServiceTokenID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Can't get ServiceTokenID")
		}
		_, err = apiClient.GetServiceToken(ServiceTokenID)
		if err == nil {
			return fmt.Errorf("ServiceToken still exists")
		}
		notFoundErr := "not found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
