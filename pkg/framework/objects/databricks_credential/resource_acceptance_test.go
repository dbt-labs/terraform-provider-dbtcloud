package databricks_credential_test

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

func TestAccDbtCloudDatabricksCredentialResourceGlobConn(t *testing.T) {

	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	targetName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	catalog := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	token := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	token2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudDatabricksCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudDatabricksCredentialResourceBasicConfigGlobConn(
					projectName,
					catalog,
					targetName,
					token,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudDatabricksCredentialExists(
						"dbtcloud_databricks_credential.test_credential",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_databricks_credential.test_credential",
						"catalog",
						catalog,
					),
				),
			},
			// ERROR schema must be provided
			{
				Config:      testCheckSchemaIsProvided(),
				ExpectError: regexp.MustCompile("`schema` must be provided when `semantic_layer_credential` is false."),
			},
			// RENAME
			// MODIFY
			{
				Config: testAccDbtCloudDatabricksCredentialResourceBasicConfigGlobConn(
					projectName,
					"",
					targetName,
					token2,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudDatabricksCredentialExists(
						"dbtcloud_databricks_credential.test_credential",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_databricks_credential.test_credential",
						"catalog",
						"",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_databricks_credential.test_credential",
						"token",
						token2,
					),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_databricks_credential.test_credential",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token", "adapter_type", "semantic_layer_credential"},
			},
		},
	})
}

func testAccDbtCloudDatabricksCredentialResourceBasicConfigGlobConn(
	projectName, catalogName, targetName, token string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_global_connection" "databricks" {
  name = "My Databricks connection"
  databricks = {
    host      = "my-databricks-host.cloud.databricks.com"
    http_path = "/sql/my/http/path"
    catalog       = "dbt_catalog"
    client_id     = "yourclientid"
    client_secret = "yourclientsecret"
  }
}

resource "dbtcloud_environment" "prod_environment" {
  dbt_version     = "versionless"
  name            = "Prod"
  project_id      = dbtcloud_project.test_project.id
  connection_id   = dbtcloud_global_connection.databricks.id
  type            = "deployment"
  credential_id   = dbtcloud_databricks_credential.test_credential.credential_id
  deployment_type = "production"
}


resource "dbtcloud_databricks_credential" "test_credential" {
    project_id = dbtcloud_project.test_project.id
    catalog = "%s"
	target_name = "%s"
    token   = "%s"
    schema  = "my_schema"
	adapter_type = "databricks"
}
`, projectName, catalogName, targetName, token)
}

func testCheckSchemaIsProvided() string {
	return `
		resource "dbtcloud_project" "test_project" {
  			name        = "test"
		}

		resource "dbtcloud_databricks_credential" "test_credential" {
    		project_id = dbtcloud_project.test_project.id
    		catalog = "test"
			target_name = "test"
    		token   = "test"
			adapter_type = "databricks"
		}
	`
}

func testAccCheckDbtCloudDatabricksCredentialExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		projectId, credentialId, err := helper.SplitIDToInts(
			rs.Primary.ID,
			"dbtcloud_databricks_credential",
		)
		if err != nil {
			return err
		}

		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
		_, err = apiClient.GetDatabricksCredential(projectId, credentialId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudDatabricksCredentialDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_databricks_credential" {
			continue
		}
		projectId, credentialId, err := helper.SplitIDToInts(
			rs.Primary.ID,
			"dbtcloud_databricks_credential",
		)
		if err != nil {
			return err
		}

		_, err = apiClient.GetDatabricksCredential(projectId, credentialId)
		if err == nil {
			return fmt.Errorf("Databricks credential still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}

func getBasicConfigTestStep(projectName, catalogName, targetName, token string) resource.TestStep {
	return resource.TestStep{
		Config: testAccDbtCloudDatabricksCredentialResourceBasicConfigGlobConn(
			projectName,
			catalogName,
			targetName,
			token,
		),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckDbtCloudDatabricksCredentialExists(
				"dbtcloud_databricks_credential.test_credential",
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_databricks_credential.test_credential",
				"target_name",
				targetName,
			),
		),
	}
}

func getModifyConfigTestStep(projectName, catalogName, targetName, targetName2, token, token2 string) resource.TestStep {
	return resource.TestStep{
		Config: testAccDbtCloudDatabricksCredentialResourceBasicConfigGlobConn(
			projectName,
			catalogName,
			targetName2,
			token2,
		),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckDbtCloudDatabricksCredentialExists(
				"dbtcloud_databricks_credential.test_credential",
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_databricks_credential.test_credential",
				"target_name",
				targetName2,
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_databricks_credential.test_credential",
				"token",
				token2,
			),
		),
	}
}
