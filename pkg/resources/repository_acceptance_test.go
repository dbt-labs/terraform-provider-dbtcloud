package resources_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDbtCloudRepositoryResource(t *testing.T) {

	repoUrl := "https://github.com/GtheSheep/baa.git"
	// 	repoUrl2 := "https://github.com/GtheSheep/dbt-core.git"
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudRepositoryDestroy,
		Steps: []resource.TestStep{
			// Create Github repository
			{
				Config: testAccDbtCloudRepositoryResourceGithubConfig(repoUrl, projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudRepositoryExists("dbt_cloud_repository.test_repository_github"),
					resource.TestCheckResourceAttr("dbt_cloud_repository.test_repository_github", "remote_url", repoUrl),
				),
			},
			// MODIFY
			// IMPORT
			{
				ResourceName:            "dbt_cloud_repository.test_repository_github",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})

	// 	resource.Test(t, resource.TestCase{
	// 		PreCheck:     func() { testAccPreCheck(t) },
	// 		Providers:    testAccProviders,
	// 		CheckDestroy: testAccCheckDbtCloudRepositoryDestroy,
	// 		Steps: []resource.TestStep{
	// 			// Create Gitlab repository
	// 			{
	// 				Config: testAccDbtCloudRepositoryResourceGitlabConfig(repoUrl, projectName),
	// 				Check: resource.ComposeTestCheckFunc(
	// 					testAccCheckDbtCloudRepositoryExists("dbt_cloud_repository.test_repository_gitlab"),
	// 					resource.TestCheckResourceAttr("dbt_cloud_repository.test_repository_gitlab", "remote_url", repoUrl),
	// 					resource.TestCheckResourceAttr("dbt_cloud_repository.test_repository_gitlab", "git_clone_strategy", "deploy_token"),
	// 				),
	// 			},
	// 			// 						MODIFY
	// 			// 			IMPORT
	// 			{
	// 				ResourceName:            "dbt_cloud_repository.test_repository_gitlab",
	// 				ImportState:             true,
	// 				ImportStateVerify:       true,
	// 				ImportStateVerifyIgnore: []string{},
	// 			},
	// 		},
	// 	})
}

func testAccDbtCloudRepositoryResourceGithubConfig(repoUrl, projectName string) string {
	return fmt.Sprintf(`
resource "dbt_cloud_project" "test_project" {
  name        = "%s"
}

resource "dbt_cloud_repository" "test_repository_github" {
  remote_url = "%s"
  project_id = dbt_cloud_project.test_project.id
}
`, projectName, repoUrl)
}

// func testAccDbtCloudRepositoryResourceGitlabConfig(repoUrl, projectName string) string {
// 	return fmt.Sprintf(`
// resource "dbt_cloud_project" "test_project" {
//   name        = "%s"
// }
//
// resource "dbt_cloud_repository" "test_repository_gitlab" {
//   remote_url = "%s"
//   project_id = dbt_cloud_project.test_project.id
//   git_clone_strategy = "deploy_token"
//   repository_credentials_id = 10
// }
// `, projectName, repoUrl)
// }

//
// func testAccDbtCloudEnvironmentResourceModifiedConfig(projectName, projectName2, environmentName string) string {
// 	return fmt.Sprintf(`
// resource "dbt_cloud_project" "test_project" {
//   name        = "%s"
// }
//
// resource "dbt_cloud_project" "test_project_2" {
//   name        = "%s"
// }
//
// resource "dbt_cloud_environment" "test_env" {
//   name        = "%s"
//   type = "deployment"
//   dbt_version = "1.0.1"
//   project_id = dbt_cloud_project.test_project_2.id
// }
// `, projectName, projectName2, environmentName)
// }

func testAccCheckDbtCloudRepositoryExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		apiClient := testAccProvider.Meta().(*dbt_cloud.Client)
		projectId := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0]
		repositoryId := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1]

		_, err := apiClient.GetRepository(repositoryId, projectId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudRepositoryDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbt_cloud_repository" {
			continue
		}
		projectId := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0]
		repositoryId := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1]

		_, err := apiClient.GetRepository(repositoryId, projectId)
		if err == nil {
			return fmt.Errorf("Repository still exists")
		}
		notFoundErr := "not found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
