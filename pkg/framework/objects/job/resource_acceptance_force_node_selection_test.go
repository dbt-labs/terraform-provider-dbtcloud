package job_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_config"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudJobResourceForceNodeSelection(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			// Test 1: Create job with force_node_selection = true and non-fusion version
			{
				Config: testAccDbtCloudJobResourceForceNodeSelectionConfig(
					jobName,
					projectName,
					environmentName,
					acctest_config.DBT_CLOUD_VERSION,
					"true",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "force_node_selection", "true"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "dbt_version", acctest_config.DBT_CLOUD_VERSION),
				),
			},
			// Test 2: Update to force_node_selection = false with latest-fusion (should succeed)
			{
				Config: testAccDbtCloudJobResourceForceNodeSelectionConfig(
					jobName,
					projectName,
					environmentName,
					"latest-fusion",
					"false",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "force_node_selection", "false"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "dbt_version", "latest-fusion"),
				),
			},
			// Test 3: Update back to non-fusion version with force_node_selection = true
			{
				Config: testAccDbtCloudJobResourceForceNodeSelectionConfig(
					jobName,
					projectName,
					environmentName,
					acctest_config.DBT_CLOUD_VERSION,
					"true",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "force_node_selection", "true"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "dbt_version", acctest_config.DBT_CLOUD_VERSION),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_job.test_job",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"triggers.%",
					"triggers.custom_branch_only",
				},
			},
		},
	})
}

func TestAccDbtCloudJobResourceForceNodeSelectionOptional(t *testing.T) {
	if acctest_config.IsDbtCloudPR() {
		t.Skip("Skipping: latest-fusion not supported in CI environment")
	}
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			// Test: Create job without force_node_selection (optional field)
			{
				Config: testAccDbtCloudJobResourceForceNodeSelectionConfig(
					jobName,
					projectName,
					environmentName,
					acctest_config.DBT_CLOUD_VERSION,
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "dbt_version", acctest_config.DBT_CLOUD_VERSION),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_job.test_job",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"triggers.%",
					"triggers.custom_branch_only",
				},
			},
		},
	})
}

func TestAccDbtCloudJobResourceForceNodeSelectionValidation(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			// Test: Try to set force_node_selection = false with non-fusion version (should fail)
			{
				Config: testAccDbtCloudJobResourceForceNodeSelectionConfig(
					jobName,
					projectName,
					environmentName,
					acctest_config.DBT_CLOUD_VERSION,
					"false",
				),
				ExpectError: regexp.MustCompile(`Invalid force_node_selection Configuration`),
			},
		},
	})
}

func TestAccDbtCloudJobResourceForceNodeSelectionWithLatestFusion(t *testing.T) {
	if acctest_config.IsDbtCloudPR() {
		t.Skip("Skipping: latest-fusion not supported in CI environment")
	}

	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			// Test: Create job with latest-fusion and force_node_selection = true
			{
				Config: testAccDbtCloudJobResourceForceNodeSelectionConfig(
					jobName,
					projectName,
					environmentName,
					"latest-fusion",
					"true",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "force_node_selection", "true"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "dbt_version", "latest-fusion"),
				),
			},
			// Test: Update to force_node_selection = false (should succeed with latest-fusion)
			{
				Config: testAccDbtCloudJobResourceForceNodeSelectionConfig(
					jobName,
					projectName,
					environmentName,
					"latest-fusion",
					"false",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "force_node_selection", "false"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "dbt_version", "latest-fusion"),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_job.test_job",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"triggers.%",
					"triggers.custom_branch_only",
				},
			},
		},
	})
}

func testAccDbtCloudJobResourceForceNodeSelectionConfig(
	jobName, projectName, environmentName, dbtVersion, forceNodeSelection string,
) string {
	forceNodeSelectionConfig := ""
	if forceNodeSelection != "" {
		forceNodeSelectionConfig = fmt.Sprintf("force_node_selection = %s", forceNodeSelection)
	}

	return fmt.Sprintf(`
resource "dbtcloud_project" "test_job_project" {
    name = "%s"
}

resource "dbtcloud_environment" "test_job_environment" {
    project_id = dbtcloud_project.test_job_project.id
    name = "%s"
    dbt_version = "%s"
    type = "deployment"
	deployment_type = "staging"
}

resource "dbtcloud_job" "test_job" {
    name = "%s"
    project_id = dbtcloud_project.test_job_project.id
    environment_id = dbtcloud_environment.test_job_environment.environment_id
    dbt_version = "%s"
    execute_steps = [
        "dbt build"
    ]
    triggers = {
        "github_webhook": false,
        "git_provider_webhook": false,
        "schedule": false
    }
    %s
}
`, projectName, environmentName, dbtVersion, jobName, dbtVersion, forceNodeSelectionConfig)
}
