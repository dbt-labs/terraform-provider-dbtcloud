package job_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_config"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDbtCloudJobResourceCIWithDeferral tests that CI jobs can be created
// with deferring_environment_id set. This validates the fix for the bug where
// the provider incorrectly dropped deferral settings for CI/Merge jobs.
func TestAccDbtCloudJobResourceCIWithDeferral(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudJobResourceCIWithDeferralConfig(
					jobName,
					projectName,
					environmentName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.ci_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "triggers.git_provider_webhook", "true"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.ci_job", "deferring_environment_id"),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_job.ci_job",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"validate_execute_steps",
				},
			},
		},
	})
}

func testAccDbtCloudJobResourceCIWithDeferralConfig(
	jobName, projectName, environmentName string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "project" {
    name = "%s"
}

resource "dbtcloud_environment" "prod_env" {
    project_id = dbtcloud_project.project.id
    name = "PROD %s"
    dbt_version = "%s"
    type = "deployment"
    deployment_type = "production"
}

resource "dbtcloud_environment" "ci_env" {
    project_id = dbtcloud_project.project.id
    name = "CI %s"
    dbt_version = "%s"
    type = "deployment"
}

resource "dbtcloud_job" "ci_job" {
    project_id = dbtcloud_project.project.id
    environment_id = dbtcloud_environment.ci_env.environment_id
    name = "%s"
    execute_steps = [
        "dbt build -s state:modified+ --fail-fast"
    ]
    # CI job with deferral - this should work after the bug fix
    deferring_environment_id = dbtcloud_environment.prod_env.environment_id
    
    triggers = {
        "github_webhook" : true
        "git_provider_webhook" : true
        "schedule" : false
        "on_merge" : false
    }
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName)
}

// TestAccDbtCloudJobResourceMergeWithDeferral tests that Merge jobs can be created
// with deferring_environment_id set. This validates the fix for the bug where
// the provider incorrectly dropped deferral settings for CI/Merge jobs.
func TestAccDbtCloudJobResourceMergeWithDeferral(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudJobResourceMergeWithDeferralConfig(
					jobName,
					projectName,
					environmentName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.merge_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.merge_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.merge_job", "triggers.on_merge", "true"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.merge_job", "deferring_environment_id"),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_job.merge_job",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"validate_execute_steps",
				},
			},
		},
	})
}

func testAccDbtCloudJobResourceMergeWithDeferralConfig(
	jobName, projectName, environmentName string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "project" {
    name = "%s"
}

resource "dbtcloud_environment" "prod_env" {
    project_id = dbtcloud_project.project.id
    name = "PROD %s"
    dbt_version = "%s"
    type = "deployment"
    deployment_type = "production"
}

resource "dbtcloud_environment" "merge_env" {
    project_id = dbtcloud_project.project.id
    name = "MERGE %s"
    dbt_version = "%s"
    type = "deployment"
}

resource "dbtcloud_job" "merge_job" {
    project_id = dbtcloud_project.project.id
    environment_id = dbtcloud_environment.merge_env.environment_id
    name = "%s"
    execute_steps = [
        "dbt build"
    ]
    # Merge job with deferral - this should work after the bug fix
    deferring_environment_id = dbtcloud_environment.prod_env.environment_id
    
    triggers = {
        "github_webhook" : false
        "git_provider_webhook" : false
        "schedule" : false
        "on_merge" : true
    }
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName)
}

// TestAccDbtCloudJobResourceCostOptimizationFeatures tests the new cost_optimization_features
// attribute on scheduled jobs. This validates the new feature works correctly.
// Note: The API may not support removing SAO features once enabled, so we only test create + import.
func TestAccDbtCloudJobResourceCostOptimizationFeatures(t *testing.T) {
	if acctest_config.IsDbtCloudPR() {
		t.Skip("Skipping: cost_optimization_features requires SAO to be enabled in account")
	}

	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			// Create with SAO enabled
			{
				Config: testAccDbtCloudJobResourceCostOptimizationFeaturesConfig(
					jobName,
					projectName,
					environmentName,
					`["state_aware_orchestration"]`,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.scheduled_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.scheduled_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.scheduled_job", "cost_optimization_features.#", "1"),
					resource.TestCheckTypeSetElemAttr("dbtcloud_job.scheduled_job", "cost_optimization_features.*", "state_aware_orchestration"),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_job.scheduled_job",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"validate_execute_steps",
				},
			},
		},
	})
}

func testAccDbtCloudJobResourceCostOptimizationFeaturesConfig(
	jobName, projectName, environmentName, costOptimizationFeatures string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "project" {
    name = "%s"
}

resource "dbtcloud_environment" "prod_env" {
    project_id = dbtcloud_project.project.id
    name = "%s"
    dbt_version = "latest-fusion"
    type = "deployment"
    deployment_type = "production"
}

resource "dbtcloud_job" "scheduled_job" {
    project_id = dbtcloud_project.project.id
    environment_id = dbtcloud_environment.prod_env.environment_id
    name = "%s"
    dbt_version = "latest-fusion"
    execute_steps = [
        "dbt build"
    ]
    
    # Use cost_optimization_features instead of force_node_selection
    cost_optimization_features = %s
    
    triggers = {
        "github_webhook" : false
        "git_provider_webhook" : false
        "schedule" : true
        "on_merge" : false
    }
    
    schedule_type = "every_day"
    schedule_hours = [9]
}
`, projectName, environmentName, jobName, costOptimizationFeatures)
}

// TestAccDbtCloudJobResourceCINoForceNodeSelection tests that CI jobs work
// correctly without specifying force_node_selection. The provider should
// automatically omit this field for CI jobs to avoid SAO validation errors.
func TestAccDbtCloudJobResourceCINoForceNodeSelection(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudJobResourceCINoForceNodeSelectionConfig(
					jobName,
					projectName,
					environmentName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.ci_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "triggers.git_provider_webhook", "true"),
					// force_node_selection should be null/unset for CI jobs
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_job.ci_job",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"validate_execute_steps",
				},
			},
		},
	})
}

func testAccDbtCloudJobResourceCINoForceNodeSelectionConfig(
	jobName, projectName, environmentName string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "project" {
    name = "%s"
}

resource "dbtcloud_environment" "ci_env" {
    project_id = dbtcloud_project.project.id
    name = "%s"
    dbt_version = "%s"
    type = "deployment"
}

resource "dbtcloud_job" "ci_job" {
    project_id = dbtcloud_project.project.id
    environment_id = dbtcloud_environment.ci_env.environment_id
    name = "%s"
    execute_steps = [
        "dbt build -s state:modified+ --fail-fast"
    ]
    # Note: force_node_selection is NOT specified here
    # The provider should automatically omit it for CI jobs
    
    triggers = {
        "github_webhook" : true
        "git_provider_webhook" : true
        "schedule" : false
        "on_merge" : false
    }
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName)
}

// TestAccDbtCloudJobResourceMergeNoForceNodeSelection tests that Merge jobs work
// correctly without specifying force_node_selection. The provider should
// automatically omit this field for Merge jobs to avoid SAO validation errors.
func TestAccDbtCloudJobResourceMergeNoForceNodeSelection(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudJobResourceMergeNoForceNodeSelectionConfig(
					jobName,
					projectName,
					environmentName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.merge_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.merge_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.merge_job", "triggers.on_merge", "true"),
					// force_node_selection should be null/unset for Merge jobs
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_job.merge_job",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"validate_execute_steps",
				},
			},
		},
	})
}

func testAccDbtCloudJobResourceMergeNoForceNodeSelectionConfig(
	jobName, projectName, environmentName string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "project" {
    name = "%s"
}

resource "dbtcloud_environment" "merge_env" {
    project_id = dbtcloud_project.project.id
    name = "%s"
    dbt_version = "%s"
    type = "deployment"
}

resource "dbtcloud_job" "merge_job" {
    project_id = dbtcloud_project.project.id
    environment_id = dbtcloud_environment.merge_env.environment_id
    name = "%s"
    execute_steps = [
        "dbt build"
    ]
    # Note: force_node_selection is NOT specified here
    # The provider should automatically omit it for Merge jobs
    
    triggers = {
        "github_webhook" : false
        "git_provider_webhook" : false
        "schedule" : false
        "on_merge" : true
    }
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName)
}

// TestAccDbtCloudJobResourceCIExplicitNullForceNodeSelection tests that CI jobs work
// correctly when force_node_selection is EXPLICITLY set to null. This is the exact
// scenario that occurs when using Terraform modules that pass null values, and validates
// the fix for the IsUnknown() bug where Computed attributes with explicit null values
// were incorrectly sent as false to the API.
func TestAccDbtCloudJobResourceCIExplicitNullForceNodeSelection(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudJobResourceCIExplicitNullConfig(
					jobName,
					projectName,
					environmentName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.ci_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "triggers.git_provider_webhook", "true"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.ci_job", "deferring_environment_id"),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_job.ci_job",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"validate_execute_steps",
					"force_node_selection", // Explicit null may not round-trip exactly
				},
			},
		},
	})
}

func testAccDbtCloudJobResourceCIExplicitNullConfig(
	jobName, projectName, environmentName string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "project" {
    name = "%s"
}

resource "dbtcloud_environment" "prod_env" {
    project_id = dbtcloud_project.project.id
    name = "PROD %s"
    dbt_version = "%s"
    type = "deployment"
    deployment_type = "production"
}

resource "dbtcloud_environment" "ci_env" {
    project_id = dbtcloud_project.project.id
    name = "CI %s"
    dbt_version = "%s"
    type = "deployment"
}

resource "dbtcloud_job" "ci_job" {
    project_id = dbtcloud_project.project.id
    environment_id = dbtcloud_environment.ci_env.environment_id
    name = "%s"
    execute_steps = [
        "dbt build -s state:modified+ --fail-fast"
    ]
    # CI job with deferral and EXPLICIT null for force_node_selection
    # This mimics what Terraform modules do when passing null values
    deferring_environment_id = dbtcloud_environment.prod_env.environment_id
    force_node_selection = null
    
    triggers = {
        "github_webhook" : true
        "git_provider_webhook" : true
        "schedule" : false
        "on_merge" : false
    }
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName)
}

// TestAccDbtCloudJobResourceMergeExplicitNullForceNodeSelection tests that Merge jobs work
// correctly when force_node_selection is EXPLICITLY set to null.
func TestAccDbtCloudJobResourceMergeExplicitNullForceNodeSelection(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudJobResourceMergeExplicitNullConfig(
					jobName,
					projectName,
					environmentName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.merge_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.merge_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.merge_job", "triggers.on_merge", "true"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.merge_job", "deferring_environment_id"),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_job.merge_job",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"validate_execute_steps",
					"force_node_selection", // Explicit null may not round-trip exactly
				},
			},
		},
	})
}

func testAccDbtCloudJobResourceMergeExplicitNullConfig(
	jobName, projectName, environmentName string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "project" {
    name = "%s"
}

resource "dbtcloud_environment" "prod_env" {
    project_id = dbtcloud_project.project.id
    name = "PROD %s"
    dbt_version = "%s"
    type = "deployment"
    deployment_type = "production"
}

resource "dbtcloud_environment" "merge_env" {
    project_id = dbtcloud_project.project.id
    name = "MERGE %s"
    dbt_version = "%s"
    type = "deployment"
}

resource "dbtcloud_job" "merge_job" {
    project_id = dbtcloud_project.project.id
    environment_id = dbtcloud_environment.merge_env.environment_id
    name = "%s"
    execute_steps = [
        "dbt build"
    ]
    # Merge job with deferral and EXPLICIT null for force_node_selection
    # This mimics what Terraform modules do when passing null values
    deferring_environment_id = dbtcloud_environment.prod_env.environment_id
    force_node_selection = null
    
    triggers = {
        "github_webhook" : false
        "git_provider_webhook" : false
        "schedule" : false
        "on_merge" : true
    }
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName)
}
