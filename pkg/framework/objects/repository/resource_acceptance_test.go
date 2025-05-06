package repository_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_config"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var repoUrlGithub = "git@github.com:dbt-labs/terraform-provider-dbtcloud.git"
var projectName = strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

var repoUrlGithubApplication = "git://github.com/victorasu/jaffle-shop.git"
var githubAppInstallationId = acctest_config.AcceptanceTestConfig.GitHubAppInstallationId
var projectNameGithubApplication = strings.ToUpper(
	acctest.RandStringFromCharSet(10, acctest.CharSetAlpha),
)

var createByDeployKeyTestStep = resource.TestStep{
	// CREATE Github repository
	Config: testAccDbtCloudRepositoryResourceGithubConfig(repoUrlGithub, projectName),
	Check: resource.ComposeTestCheckFunc(
		testAccCheckDbtCloudRepositoryExists(
			"dbtcloud_repository.test_repository_github_app",
		),
		resource.TestCheckResourceAttr(
			"dbtcloud_repository.test_repository_github_app",
			"remote_url",
			repoUrlGithub,
		),
		resource.TestCheckResourceAttrSet(
			"dbtcloud_repository.test_repository_github_app",
			"deploy_key",
		),
	),
}

var createByCloneTestStep = resource.TestStep{
	// CREATE Github repository via clone
	Config: testAccDbtCloudRepositoryResourceGithubApplicationConfig(
		repoUrlGithubApplication,
		projectNameGithubApplication,
		githubAppInstallationId,
	),
	Check: resource.ComposeTestCheckFunc(
		testAccCheckDbtCloudRepositoryExists(
			"dbtcloud_repository.test_repository_github",
		),
		resource.TestCheckResourceAttr(
			"dbtcloud_repository.test_repository_github",
			"remote_url",
			repoUrlGithubApplication,
		),
		resource.TestCheckResourceAttr(
			"dbtcloud_repository.test_repository_github",
			"git_clone_strategy",
			"github_app",
		),
	),
}

// TODO: Add more tests
// we are currently testing in CI the SSH cloning and GH native cloning but not GitLab native and ADO native
// this would require having the GitLab and ADO native integrations set up in the dbt Cloud account used for CI
func TestAccDbtCloudRepositoryResource(t *testing.T) {

	var importDeployTestStep = resource.TestStep{
		// IMPORT
		ResourceName:            "dbtcloud_repository.test_repository_github_app",
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"fetch_deploy_key"},
	}

	var importCloneTestStep = resource.TestStep{
		// IMPORT
		ResourceName:            "dbtcloud_repository.test_repository_github",
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"fetch_deploy_key"},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudRepositoryDestroy,
		Steps: []resource.TestStep{
			createByDeployKeyTestStep,
			importDeployTestStep,
		},
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudRepositoryDestroy,
		Steps: []resource.TestStep{
			createByCloneTestStep,
			importCloneTestStep,
		},
	})

}

func testAccDbtCloudRepositoryResourceGithubConfig(repoUrl, projectName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_repository" "test_repository_github_app" {
  remote_url = "%s"
  project_id = dbtcloud_project.test_project.id
  git_clone_strategy = "deploy_key"
  pull_request_url_template = "https://github.com/my-org/my-repo/compare/qa...{{source}}"
}
`, projectName, repoUrl)
}

func testAccDbtCloudRepositoryResourceGithubApplicationConfig(
	repoUrl string,
	projectName string,
	githubAppInstallationId int,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_repository" "test_repository_github" {
  remote_url = "%s"
  project_id = dbtcloud_project.test_project.id
  github_installation_id = %d
  git_clone_strategy = "github_app"
  pull_request_url_template = "https://github.com/my-org/my-repo/compare/qa...{{source}}"
}
`, projectName, repoUrl, githubAppInstallationId)
}

func testAccCheckDbtCloudRepositoryExists(resource string) resource.TestCheckFunc {
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
		parts := strings.Split(rs.Primary.ID, ":")
		if len(parts) != 2 {
			return fmt.Errorf("Unexpected format of ID (%s), expected project_id:repository_id", rs.Primary.ID)
		}
		projectId := parts[0]
		repositoryId := parts[1]
		_, err = apiClient.GetRepository(repositoryId, projectId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudRepositoryDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_repository" {
			continue
		}
		parts := strings.Split(rs.Primary.ID, ":")
		if len(parts) != 2 {
			return fmt.Errorf("Unexpected format of ID (%s), expected project_id:repository_id", rs.Primary.ID)
		}
		projectId := parts[0]
		repositoryId := parts[1]
		_, err = apiClient.GetRepository(repositoryId, projectId)
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
