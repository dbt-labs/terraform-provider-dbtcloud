package spark_credential_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudSparkCredentialResourceGlobConn(t *testing.T) {

	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	token := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	token2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudSparkCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudSparkCredentialResourceBasicConfigGlobConn(
					projectName,
					token,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudSparkCredentialExists(
						"dbtcloud_spark_credential.test_spark_credential",
					),
				),
			},
			// MODIFY
			{
				Config: testAccDbtCloudSparkCredentialResourceBasicConfigGlobConn(
					projectName,
					token2,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudSparkCredentialExists(
						"dbtcloud_spark_credential.test_spark_credential",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_spark_credential.test_spark_credential",
						"token",
						token2,
					),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_spark_credential.test_spark_credential",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func testAccDbtCloudSparkCredentialResourceBasicConfigGlobConn(
	projectName, token string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_spark_credential" "test_spark_credential" {
  project_id = dbtcloud_project.test_project.id
  token      = "%s"
  schema     = "my_schema"
}
  
resource "dbtcloud_global_connection" "apache_spark" {
  name = "My Awesome Apache Spark connection"
  apache_spark = {
    method          = "http"
    host            = "my-spark-host.com"
    cluster         = "123-12345-example"
    connect_timeout = 100
  }

}

resource "dbtcloud_environment" "spark_environment" {
  dbt_version     = "versionless"
  name            = "Spark Env"
  project_id      = dbtcloud_project.test_project.id
  connection_id   = dbtcloud_global_connection.apache_spark.id
  type            = "deployment"
  credential_id   = dbtcloud_spark_credential.test_spark_credential.credential_id
  deployment_type = "production"
}


`, projectName, token)
}

func TestAccDbtCloudSparkCredentialResourceWriteOnly(t *testing.T) {
	t.Skip("Requires Terraform >= 1.11")

	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	token := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	token2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudSparkCredentialDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with write-only token
			{
				Config: testAccDbtCloudSparkCredentialWriteOnlyConfig(
					projectName,
					token,
					1,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudSparkCredentialExists(
						"dbtcloud_spark_credential.test_spark_credential_wo",
					),
					// token_wo should not be in state
					resource.TestCheckNoResourceAttr(
						"dbtcloud_spark_credential.test_spark_credential_wo",
						"token_wo",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_spark_credential.test_spark_credential_wo",
						"token_wo_version",
						"1",
					),
				),
			},
			// Step 2: Update by incrementing version with new token
			{
				Config: testAccDbtCloudSparkCredentialWriteOnlyConfig(
					projectName,
					token2,
					2,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudSparkCredentialExists(
						"dbtcloud_spark_credential.test_spark_credential_wo",
					),
					resource.TestCheckNoResourceAttr(
						"dbtcloud_spark_credential.test_spark_credential_wo",
						"token_wo",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_spark_credential.test_spark_credential_wo",
						"token_wo_version",
						"2",
					),
				),
			},
			// Step 3: Import
			{
				ResourceName:      "dbtcloud_spark_credential.test_spark_credential_wo",
				ImportState:        true,
				ImportStateVerify:  true,
				ImportStateVerifyIgnore: []string{
					"token", "token_wo", "token_wo_version",
				},
			},
		},
	})
}

func testAccDbtCloudSparkCredentialWriteOnlyConfig(
	projectName, tokenWo string, tokenWoVersion int,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_spark_credential" "test_spark_credential_wo" {
  project_id       = dbtcloud_project.test_project.id
  token_wo         = "%s"
  token_wo_version = %d
  schema           = "my_schema"
}

resource "dbtcloud_global_connection" "apache_spark" {
  name = "My Awesome Apache Spark connection"
  apache_spark = {
    method          = "http"
    host            = "my-spark-host.com"
    cluster         = "123-12345-example"
    connect_timeout = 100
  }

}

resource "dbtcloud_environment" "spark_environment" {
  dbt_version     = "versionless"
  name            = "Spark Env"
  project_id      = dbtcloud_project.test_project.id
  connection_id   = dbtcloud_global_connection.apache_spark.id
  type            = "deployment"
  credential_id   = dbtcloud_spark_credential.test_spark_credential_wo.credential_id
  deployment_type = "production"
}


`, projectName, tokenWo, tokenWoVersion)
}

func testAccCheckDbtCloudSparkCredentialExists(resource string) resource.TestCheckFunc {
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
			"dbtcloud_spark_credential",
		)
		if err != nil {
			return err
		}

		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
		_, err = apiClient.GetSparkCredential(projectId, credentialId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudSparkCredentialDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_spark_credential" {
			continue
		}
		projectId, credentialId, err := helper.SplitIDToInts(
			rs.Primary.ID,
			"dbtcloud_spark_credential",
		)
		if err != nil {
			return err
		}

		_, err = apiClient.GetSparkCredential(projectId, credentialId)
		if err == nil {
			return fmt.Errorf("Apache Spark credential still exists")
		}
	}

	return nil
}
