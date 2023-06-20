package resources_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDbtCloudRepositoryResource(t *testing.T) {

	repoUrlGithub := "git@github.com:dbt-labs/terraform-provider-dbtcloud.git"
	// 	repoUrlGitlab := "GtheSheep/test"
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudRepositoryDestroy,
		Steps: []resource.TestStep{
			// Create Github repository
			{
				Config: testAccDbtCloudRepositoryResourceGithubConfig(repoUrlGithub, projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudRepositoryExists("dbtcloud_repository.test_repository_github"),
					resource.TestCheckResourceAttr("dbtcloud_repository.test_repository_github", "remote_url", repoUrlGithub),
					resource.TestCheckResourceAttrSet("dbtcloud_repository.test_repository_github", "deploy_key"),
				),
			},
			// MODIFY
			// IMPORT
			{
				ResourceName:            "dbtcloud_repository.test_repository_github",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"fetch_deploy_key"},
			},
		},
	})

	repoUrlGithubApplication := "git://github.com/dbt-labs/jaffle_shop.git"
	projectNameGithubApplication := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudRepositoryDestroy,
		Steps: []resource.TestStep{
			// Create Github repository via the GithUb Application
			{
				Config: testAccDbtCloudRepositoryResourceGithubApplicationConfig(repoUrlGithubApplication, projectNameGithubApplication),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudRepositoryExists("dbtcloud_repository.test_repository_github_application"),
					resource.TestCheckResourceAttr("dbtcloud_repository.test_repository_github_application", "remote_url", repoUrlGithubApplication),
					resource.TestCheckResourceAttr("dbtcloud_repository.test_repository_github_application", "git_clone_strategy", "github_app"),
				),
			},
			// MODIFY
			// IMPORT
			{
				ResourceName:            "dbtcloud_repository.test_repository_github_application",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"fetch_deploy_key"},
			},
		},
	})

	//
	// 		resource.Test(t, resource.TestCase{
	// 			PreCheck:     func() { testAccPreCheck(t) },
	// 			Providers:    testAccProviders,
	// 			CheckDestroy: testAccCheckDbtCloudRepositoryDestroy,
	// 			Steps: []resource.TestStep{
	// 				// Create Gitlab repository
	// 				{
	// 					Config: testAccDbtCloudRepositoryResourceGitlabConfig(repoUrlGitlab, projectName),
	// 					Check: resource.ComposeTestCheckFunc(
	// 						testAccCheckDbtCloudRepositoryExists("dbtcloud_repository.test_repository_gitlab"),
	// 						resource.TestCheckResourceAttr("dbtcloud_repository.test_repository_gitlab", "remote_url", repoUrlGitlab),
	// 						resource.TestCheckResourceAttr("dbtcloud_repository.test_repository_gitlab", "git_clone_strategy", "deploy_token"),
	// 					),
	// 				},
	// 				// 						MODIFY
	// 				// 			IMPORT
	// 				{
	// 					ResourceName:            "dbtcloud_repository.test_repository_gitlab",
	// 					ImportState:             true,
	// 					ImportStateVerify:       true,
	// 					ImportStateVerifyIgnore: []string{},
	// 				},
	// 			},
	// 		})
}

func testAccDbtCloudRepositoryResourceGithubConfig(repoUrl, projectName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_repository" "test_repository_github" {
  remote_url = "%s"
  project_id = dbtcloud_project.test_project.id
  depends_on = [dbtcloud_project.test_project]
}
`, projectName, repoUrl)
}

func testAccDbtCloudRepositoryResourceGithubApplicationConfig(repoUrl, projectName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_repository" "test_repository_github_application" {
  remote_url = "%s"
  project_id = dbtcloud_project.test_project.id
  github_installation_id = 28374841
  git_clone_strategy = "github_app"
}
`, projectName, repoUrl)
}

//
// func testAccDbtCloudRepositoryResourceGitlabConfig(repoUrl, projectName string) string {
// 	return fmt.Sprintf(`
// resource "dbtcloud_project" "test_project" {
//   name        = "%s"
// }
//
// resource "dbtcloud_repository" "test_repository_gitlab" {
//   remote_url = "%s"
//   project_id = dbtcloud_project.test_project.id
//   gitlab_project_id = 34786716
// }
// `, projectName, repoUrl)
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

		_, err := apiClient.GetRepository(repositoryId, projectId, false)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudRepositoryDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_repository" {
			continue
		}
		projectId := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0]
		repositoryId := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1]

		_, err := apiClient.GetRepository(repositoryId, projectId, false)
		if err == nil {
			return fmt.Errorf("Repository still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
