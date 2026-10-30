package project_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var projectName = "testproject1"
var projectDescription = "testprojectdescription"
var projectName2 = "testproject2"

var createStep = resource.TestStep{
	Config: testAccDbtCloudProjectResourceBasicConfig(projectName),
	Check: resource.ComposeTestCheckFunc(
		testAccCheckDbtCloudProjectExists("dbtcloud_project.test_project"),
		resource.TestCheckResourceAttr(
			"dbtcloud_project.test_project",
			"name",
			projectName,
		),
	),
}

var renameStep = resource.TestStep{
	Config: testAccDbtCloudProjectResourceBasicConfig(projectName2),
	Check: resource.ComposeTestCheckFunc(
		testAccCheckDbtCloudProjectExists("dbtcloud_project.test_project"),
		resource.TestCheckResourceAttr(
			"dbtcloud_project.test_project",
			"name",
			projectName2,
		),
	),
}

var modifyStep = resource.TestStep{
	Config: testAccDbtCloudProjectResourceFullConfig(projectName, projectDescription),
	Check: resource.ComposeTestCheckFunc(
		testAccCheckDbtCloudProjectExists("dbtcloud_project.test_project"),
		resource.TestCheckResourceAttr(
			"dbtcloud_project.test_project",
			"name",
			projectName,
		),
		resource.TestCheckResourceAttr(
			"dbtcloud_project.test_project",
			"dbt_project_subdirectory",
			"project/subdirectory_where/dbt-is",
		),
		resource.TestCheckResourceAttr(
			"dbtcloud_project.test_project",
			"type",
			"0",
		),
	),
}

var emptyStringStep = resource.TestStep{
	Config: testAccDbtCloudProjectResourceEmptyStringConfig(projectName, projectDescription),
	Check: resource.ComposeTestCheckFunc(
		testAccCheckDbtCloudProjectExists("dbtcloud_project.test_project"),
		resource.TestCheckResourceAttr(
			"dbtcloud_project.test_project",
			"name",
			projectName,
		),
		resource.TestCheckResourceAttr(
			"dbtcloud_project.test_project",
			"dbt_project_subdirectory",
			"",
		),
		resource.TestCheckResourceAttr(
			"dbtcloud_project.test_project",
			"type",
			"0",
		),
	),
}

func TestAccDbtCloudProjectResource(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudProjectDestroy,
		Steps: []resource.TestStep{
			createStep,
			renameStep,
			modifyStep,
			emptyStringStep,
			{
				ResourceName:            "dbtcloud_project.test_project",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccDbtCloudProjectResourceBasicConfig(projectName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}
`, projectName)
}

func testAccDbtCloudProjectResourceFullConfig(
	projectName string,
	projectDescription string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
  description = "%s"
  dbt_project_subdirectory = "project/subdirectory_where/dbt-is"
  type = 0
}
`, projectName, projectDescription)
}

func testAccDbtCloudProjectResourceEmptyStringConfig(
	projectName string,
	projectDescription string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
  description = "%s"
  dbt_project_subdirectory = ""
  type = 0
}
`, projectName, projectDescription)
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
		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
		_, err = apiClient.GetProject(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudProjectDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_project" {
			continue
		}
		_, err := apiClient.GetProject(rs.Primary.ID)
		if err == nil {
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

// Rule: a project can use a connection as its analytics connection, and keeps it on unrelated updates.
// given: a Snowflake connection and a project that uses it as analytics connection
// when: the project is created, updated on another attribute, imported and read by the data source
// then: the analytics connection is set each time, and removing the attribute clears it.
func TestAccDbtCloudProjectAnalyticsConnection(t *testing.T) {
	if os.Getenv("DBT_ACCEPTANCE_TEST_ANALYTICS_CONNECTION_ENABLED") != "true" {
		t.Skip("Skipping: set DBT_ACCEPTANCE_TEST_ANALYTICS_CONNECTION_ENABLED=true when the account supports an analytics connection on projects.")
	}
	name := "tf-acc-analytics-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	connection := fmt.Sprintf(`
resource "dbtcloud_global_connection" "snowflake" {
  name = "%s"

  snowflake = {
    account   = "analytics-acc"
    database  = "ANALYTICS_DB"
    warehouse = "ANALYTICS_WH"
    allow_sso = false
  }
}
`, name)
	project := func(attributes string) string {
		return fmt.Sprintf(`
resource "dbtcloud_project" "analytics" {
  name = "%s"
  %s
}

data "dbtcloud_project" "analytics" {
  id = dbtcloud_project.analytics.id
}
`, name, attributes)
	}
	attached := "analytics_connection_id = dbtcloud_global_connection.snowflake.id"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: connection + project(attached),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("dbtcloud_project.analytics", "analytics_connection_id", "dbtcloud_global_connection.snowflake", "id"),
					resource.TestCheckResourceAttrPair("data.dbtcloud_project.analytics", "analytics_connection_id", "dbtcloud_global_connection.snowflake", "id"),
				),
			},
			{
				Config: connection + project(attached+"\n  description = \"updated\""),
				Check:  resource.TestCheckResourceAttrPair("dbtcloud_project.analytics", "analytics_connection_id", "dbtcloud_global_connection.snowflake", "id"),
			},
			{
				Config:            connection + project(attached),
				ResourceName:      "dbtcloud_project.analytics",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: connection + project(""),
				Check:  resource.TestCheckNoResourceAttr("dbtcloud_project.analytics", "analytics_connection_id"),
			},
		},
	})
}
