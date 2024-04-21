package data_sources_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDbtCloudEnvironmentVariableDataSource(t *testing.T) {

	projectName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	environmentName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	environmentVariableName := strings.ToUpper(acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum))

	config := environmentVariable(projectName, environmentName, environmentVariableName)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("data.dbtcloud_environment_variable.test_env_var_read", "name", fmt.Sprintf("DBT_%s", environmentVariableName)),
		resource.TestCheckResourceAttrSet("data.dbtcloud_environment_variable.test_env_var_read", "project_id"),
		resource.TestCheckResourceAttr("data.dbtcloud_environment_variable.test_env_var_read", "environment_values.%", "2"),
		resource.TestCheckResourceAttr("data.dbtcloud_environment_variable.test_env_var_read", "environment_values.project", "Baa"),
		resource.TestCheckResourceAttr("data.dbtcloud_environment_variable.test_env_var_read", fmt.Sprintf("environment_values.%s", environmentName), "Moo"),
	)

	resource.ParallelTest(t, resource.TestCase{
		Providers: providers(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  check,
			},
		},
	})
}

func environmentVariable(projectName, environmentName, environmentVariableName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_environment" "test_env" {
  name        = "%s"
  type = "deployment"
  dbt_version = "%s"
  project_id = dbtcloud_project.test_project.id
}

resource "dbtcloud_environment_variable" "test_env_var" {
  name        = "DBT_%s"
  project_id = dbtcloud_project.test_project.id
  environment_values = {
    "project": "Baa",
    "%s": "Moo"
  }
  depends_on = [
    dbtcloud_project.test_project,
    dbtcloud_environment.test_env
  ]
}

data "dbtcloud_environment_variable" "test_env_var_read" {
  name = dbtcloud_environment_variable.test_env_var.name
  project_id = dbtcloud_environment_variable.test_env_var.project_id
}
`, projectName, environmentName, DBT_CLOUD_VERSION, environmentVariableName, environmentName)
}
