package resources_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudProjectArtefactsResource(t *testing.T) {

	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudProjectArtefactsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudProjectArtefactsResourceBasicConfig(
					projectName,
					environmentName,
					jobName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudProjectArtefactsExists(
						"dbtcloud_project_artefacts.test_project_artefacts",
					),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_project_artefacts.test_project_artefacts",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// EMPTY
			{
				Config: testAccDbtCloudProjectArtefactsResourceEmptyConfig(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudProjectArtefactsEmpty("dbtcloud_project.test_project"),
				),
			},
		},
	})
}

func testAccDbtCloudProjectArtefactsResourceBasicConfig(
	projectName, environmentName, jobName string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_artefacts_project" {
	name = "%s"
}

resource "dbtcloud_environment" "test_job_environment" {
	project_id = dbtcloud_project.test_artefacts_project.id
	name = "%s"
	dbt_version = "%s"
	type = "development"
}

resource "dbtcloud_job" "test_job" {
	name        = "%s"
	project_id = dbtcloud_project.test_artefacts_project.id
	environment_id = dbtcloud_environment.test_job_environment.environment_id
	execute_steps = [
	"dbt test"
	]
	triggers = {
	"github_webhook": false,
	"git_provider_webhook": false,
	"schedule": false,
	}
	run_generate_sources = true
	generate_docs = true
}

resource "dbtcloud_project_artefacts" "test_project_artefacts" {
  project_id = dbtcloud_project.test_artefacts_project.id
  docs_job_id = dbtcloud_job.test_job.id
  freshness_job_id = dbtcloud_job.test_job.id
}
`, projectName, environmentName, DBT_CLOUD_VERSION, jobName)
}

func testAccDbtCloudProjectArtefactsResourceEmptyConfig(projectName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_project_artefacts" "test_project_artefacts" {
	project_id = dbtcloud_project.test_project.id
	docs_job_id = 0
	freshness_job_id = 0
  }
`, projectName)
}

func testAccCheckDbtCloudProjectArtefactsExists(resource string) resource.TestCheckFunc {
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
		project, err := apiClient.GetProject(projectId)
		if err != nil {
			return fmt.Errorf("Can't get project")
		}
		if project.DocsJobId == nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		if project.FreshnessJobId == nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudProjectArtefactsEmpty(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		apiClient := testAccProvider.Meta().(*dbt_cloud.Client)
		project, err := apiClient.GetProject(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Can't get project")
		}
		if project.DocsJobId != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		if project.FreshnessJobId != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudProjectArtefactsDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_project_artefacts" {
			continue
		}
		projectId := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0]
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
