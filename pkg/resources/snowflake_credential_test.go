package resources_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/gthesheep/terraform-provider-dbt_cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDbtCloudSnowflakeCredentialResource(t *testing.T) {

	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	database := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	role := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	warehouse := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	schema := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	user := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	password := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	privateKey := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	privateKeyPassphrase := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudSnowflakeCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudSnowflakeCredentialResourceBasicConfig(projectName, database, role, warehouse, schema, user, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudSnowflakeCredentialExists("dbt_cloud_snowflake_credential.test_credential"),
					resource.TestCheckResourceAttr("dbt_cloud_snowflake_credential.test_credential", "database", database),
					resource.TestCheckResourceAttr("dbt_cloud_snowflake_credential.test_credential", "role", role),
					resource.TestCheckResourceAttr("dbt_cloud_snowflake_credential.test_credential", "warehouse", warehouse),
					resource.TestCheckResourceAttr("dbt_cloud_snowflake_credential.test_credential", "schema", schema),
					resource.TestCheckResourceAttr("dbt_cloud_snowflake_credential.test_credential", "user", user),
				),
			},
			// RENAME
			// MODIFY
			// IMPORT
			{
				ResourceName:            "dbt_cloud_snowflake_credential.test_credential",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudSnowflakeCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudSnowflakeCredentialResourceBasicPrivateKeyConfig(projectName, database, role, warehouse, schema, user, privateKey, privateKeyPassphrase),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudSnowflakeCredentialExists("dbt_cloud_snowflake_credential.test_credential_p"),
					resource.TestCheckResourceAttr("dbt_cloud_snowflake_credential.test_credential_p", "database", database),
					resource.TestCheckResourceAttr("dbt_cloud_snowflake_credential.test_credential_p", "role", role),
					resource.TestCheckResourceAttr("dbt_cloud_snowflake_credential.test_credential_p", "warehouse", warehouse),
					resource.TestCheckResourceAttr("dbt_cloud_snowflake_credential.test_credential_p", "schema", schema),
					resource.TestCheckResourceAttr("dbt_cloud_snowflake_credential.test_credential_p", "user", user),
					resource.TestCheckResourceAttr("dbt_cloud_snowflake_credential.test_credential_p", "private_key", privateKey),
					resource.TestCheckResourceAttr("dbt_cloud_snowflake_credential.test_credential_p", "private_key_passphrase", privateKeyPassphrase),
				),
			},
			// RENAME
			// MODIFY
			// IMPORT
			{
				ResourceName:            "dbt_cloud_snowflake_credential.test_credential_p",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_key", "private_key_passphrase"},
			},
		},
	})
}

func testAccDbtCloudSnowflakeCredentialResourceBasicConfig(projectName, database, role, warehouse, schema, user, password string) string {
	return fmt.Sprintf(`
resource "dbt_cloud_project" "test_project" {
  name        = "%s"
}
resource "dbt_cloud_snowflake_credential" "test_credential" {
    is_active = true
    project_id = dbt_cloud_project.test_project.id
    auth_type = "password"
	database = "%s"
	role = "%s"
	warehouse = "%s"
    schema = "%s"
    user = "%s"
    password = "%s"
    num_threads = 3
}
`, projectName, database, role, warehouse, schema, user, password)
}

func testAccDbtCloudSnowflakeCredentialResourceBasicPrivateKeyConfig(projectName, database, role, warehouse, schema, user, private_key, private_key_passphrase string) string {
	return fmt.Sprintf(`
resource "dbt_cloud_project" "test_project" {
  name        = "%s"
}
resource "dbt_cloud_snowflake_credential" "test_credential_p" {
    is_active = true
    project_id = dbt_cloud_project.test_project.id
    auth_type = "keypair"
	database = "%s"
	role = "%s"
	warehouse = "%s"
    schema = "%s"
    user = "%s"
    private_key = "%s"
    private_key_passphrase = "%s"
    num_threads = 3
}
`, projectName, database, role, warehouse, schema, user, private_key, private_key_passphrase)
}

func testAccCheckDbtCloudSnowflakeCredentialExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		projectId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0])
		if err != nil {
			return fmt.Errorf("Can't get projectId")
		}
		credentialId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1])
		if err != nil {
			return fmt.Errorf("Can't get projectId")
		}

		apiClient := testAccProvider.Meta().(*dbt_cloud.Client)
		_, err = apiClient.GetSnowflakeCredential(projectId, credentialId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudSnowflakeCredentialDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbt_cloud_snowflake_credential" {
			continue
		}
		projectId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0])
		if err != nil {
			return fmt.Errorf("Can't get projectId")
		}
		credentialId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1])
		if err != nil {
			return fmt.Errorf("Can't get projectId")
		}

		_, err = apiClient.GetSnowflakeCredential(projectId, credentialId)
		if err == nil {
			return fmt.Errorf("Snowflake credential still exists")
		}
		notFoundErr := "did not find"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
