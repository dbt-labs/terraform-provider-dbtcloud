package project_repository_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var projectName = strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
var repoUrlGithub = "git@github.com:dbt-labs/terraform-provider-dbtcloud.git"

var basicConfigTestStep = resource.TestStep{
	Config: testAccDbtCloudProjectRepositoryResourceBasicConfig(
		projectName,
		repoUrlGithub,
	),
	Check: resource.ComposeTestCheckFunc(
		testAccCheckDbtCloudProjectRepositoryExists(
			"dbtcloud_project_repository.test_project_repository",
		),
	),
}

var emptyConfigTestStep = resource.TestStep{
	Config: testAccDbtCloudProjectRepositoryResourceEmptyConfig(
		projectName,
		repoUrlGithub,
	),
	Check: resource.ComposeTestCheckFunc(
		testAccCheckDbtCloudProjectRepositoryEmpty("dbtcloud_project.test_project"),
	),
}

func TestAccDbtCloudProjectRepositoryResource(t *testing.T) {

	importStateTestStep := resource.TestStep{
		ResourceName:            "dbtcloud_project_repository.test_project_repository",
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudProjectRepositoryDestroy,
		Steps: []resource.TestStep{
			basicConfigTestStep,
			importStateTestStep,
			emptyConfigTestStep,
		},
	})
}

func testAccDbtCloudProjectRepositoryResourceBasicConfig(projectName, repoUrlGithub string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_repository" "test_repository" {
  remote_url = "%s"
  project_id = dbtcloud_project.test_project.id
  depends_on = [dbtcloud_project.test_project]
}

resource "dbtcloud_project_repository" "test_project_repository" {
  project_id = dbtcloud_project.test_project.id
  repository_id = dbtcloud_repository.test_repository.repository_id
}
`, projectName, repoUrlGithub)
}

func testAccDbtCloudProjectRepositoryResourceEmptyConfig(projectName, repoUrlGithub string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_repository" "test_repository" {
  remote_url = "%s"
  project_id = dbtcloud_project.test_project.id
  depends_on = [dbtcloud_project.test_project]
}
`, projectName, repoUrlGithub)
}

func testAccCheckDbtCloudProjectRepositoryExists(resource string) resource.TestCheckFunc {
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
		projectId, _, err := helper.SplitIDToStrings(
			rs.Primary.ID,
			"dbtcloud_project_repository",
		)
		if err != nil {
			return err
		}
		project, err := apiClient.GetProject(projectId)
		if err != nil {
			return fmt.Errorf("Can't get project")
		}
		if project.RepositoryID == nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudProjectRepositoryEmpty(resource string) resource.TestCheckFunc {
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
		project, err := apiClient.GetProject(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Can't get project")
		}
		if project.RepositoryID != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudProjectRepositoryDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_project_repository" {
			continue
		}
		projectId, _, err := helper.SplitIDToStrings(
			rs.Primary.ID,
			"dbtcloud_project_repository",
		)
		if err != nil {
			return err
		}
		project, err := apiClient.GetProject(projectId)
		if project != nil {
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
