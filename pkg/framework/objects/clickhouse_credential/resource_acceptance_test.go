package clickhouse_credential_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudClickhouseCredentialResource(t *testing.T) {
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	user := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	user2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	password := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudClickhouseCredentialDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDbtCloudClickhouseCredentialResourceConfig(
					projectName,
					connectionName,
					user,
					password,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudClickhouseCredentialExists(
						"dbtcloud_clickhouse_credential.test_credential",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_clickhouse_credential.test_credential",
						"user",
						user,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_clickhouse_credential.test_credential",
						"schema",
						"my_schema",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_clickhouse_credential.test_credential",
						"target_name",
						"default",
					),
					// threads is left unset in the config, so it should stay
					// null rather than defaulting to a fixed value.
					resource.TestCheckNoResourceAttr(
						"dbtcloud_clickhouse_credential.test_credential",
						"threads",
					),
				),
			},
			// Update testing
			{
				Config: testAccDbtCloudClickhouseCredentialResourceConfig(
					projectName,
					connectionName,
					user2,
					password,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudClickhouseCredentialExists(
						"dbtcloud_clickhouse_credential.test_credential",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_clickhouse_credential.test_credential",
						"user",
						user2,
					),
				),
			},
			// Update testing - explicitly setting threads should round-trip
			{
				Config: testAccDbtCloudClickhouseCredentialResourceConfigWithThreads(
					projectName,
					connectionName,
					user2,
					password,
					8,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudClickhouseCredentialExists(
						"dbtcloud_clickhouse_credential.test_credential",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_clickhouse_credential.test_credential",
						"threads",
						"8",
					),
				),
			},
			// Import testing
			{
				ResourceName:            "dbtcloud_clickhouse_credential.test_credential",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func testAccDbtCloudClickhouseCredentialResourceConfig(
	projectName, connectionName, user, password string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_global_connection" "clickhouse" {
  name = "%s"
  clickhouse = {
    host     = "example.com"
    port     = 8443
    database = "default"
  }
}

resource "dbtcloud_clickhouse_credential" "test_credential" {
  project_id = dbtcloud_project.test_project.id
  schema     = "my_schema"
  user       = "%s"
  password   = "%s"
}
`, projectName, connectionName, user, password)
}

func testAccDbtCloudClickhouseCredentialResourceConfigWithThreads(
	projectName, connectionName, user, password string, threads int,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_global_connection" "clickhouse" {
  name = "%s"
  clickhouse = {
    host     = "example.com"
    port     = 8443
    database = "default"
  }
}

resource "dbtcloud_clickhouse_credential" "test_credential" {
  project_id = dbtcloud_project.test_project.id
  schema     = "my_schema"
  user       = "%s"
  password   = "%s"
  threads    = %d
}
`, projectName, connectionName, user, password, threads)
}

func testAccCheckDbtCloudClickhouseCredentialExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ClickHouse credential ID is set")
		}

		return nil
	}
}

func TestAccDbtCloudClickhouseCredentialResourceWriteOnly(t *testing.T) {

	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	user := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	password := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	password2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudClickhouseCredentialDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with write-only password
			{
				Config: testAccDbtCloudClickhouseCredentialWriteOnlyConfig(
					projectName, connectionName, user, password, 1,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudClickhouseCredentialExists(
						"dbtcloud_clickhouse_credential.test_credential_wo",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_clickhouse_credential.test_credential_wo",
						"user",
						user,
					),
					resource.TestCheckNoResourceAttr(
						"dbtcloud_clickhouse_credential.test_credential_wo",
						"password_wo",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_clickhouse_credential.test_credential_wo",
						"password_wo_version",
						"1",
					),
				),
			},
			// Step 2: Update by incrementing version with new password
			{
				Config: testAccDbtCloudClickhouseCredentialWriteOnlyConfig(
					projectName, connectionName, user, password2, 2,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudClickhouseCredentialExists(
						"dbtcloud_clickhouse_credential.test_credential_wo",
					),
					resource.TestCheckNoResourceAttr(
						"dbtcloud_clickhouse_credential.test_credential_wo",
						"password_wo",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_clickhouse_credential.test_credential_wo",
						"password_wo_version",
						"2",
					),
				),
			},
			// Step 3: Import
			{
				ResourceName:      "dbtcloud_clickhouse_credential.test_credential_wo",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password", "password_wo", "password_wo_version",
				},
			},
		},
	})
}

func testAccDbtCloudClickhouseCredentialWriteOnlyConfig(
	projectName, connectionName, user, passwordWo string, passwordWoVersion int,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_global_connection" "clickhouse" {
  name = "%s"
  clickhouse = {
    host     = "example.com"
    port     = 8443
    database = "default"
  }
}

resource "dbtcloud_clickhouse_credential" "test_credential_wo" {
  project_id           = dbtcloud_project.test_project.id
  schema               = "my_schema"
  user                 = "%s"
  password_wo          = "%s"
  password_wo_version  = %d
}
`, projectName, connectionName, user, passwordWo, passwordWoVersion)
}

func testAccCheckDbtCloudClickhouseCredentialDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_clickhouse_credential" {
			continue
		}

		idParts := strings.Split(rs.Primary.ID, ":")
		if len(idParts) != 2 {
			return fmt.Errorf("Unexpected ID format: %s", rs.Primary.ID)
		}

		var projectID, credentialID int
		fmt.Sscanf(idParts[0], "%d", &projectID)
		fmt.Sscanf(idParts[1], "%d", &credentialID)

		_, err := apiClient.GetClickhouseCredential(projectID, credentialID)
		if err == nil {
			return fmt.Errorf("ClickHouse credential still exists")
		}

		if !strings.HasPrefix(err.Error(), "resource-not-found") {
			return fmt.Errorf("Unexpected error: %s", err.Error())
		}
	}

	return nil
}
