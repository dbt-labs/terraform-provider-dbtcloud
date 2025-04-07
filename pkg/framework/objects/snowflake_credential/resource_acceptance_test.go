package snowflake_credential_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_config"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func getBasicConfigTestStep(projectName, database, role, warehouse, schema, user, password string) resource.TestStep {
	return resource.TestStep{
		Config: testAccDbtCloudSnowflakeCredentialResourceBasicConfig(
			projectName,
			database,
			role,
			warehouse,
			schema,
			user,
			password,
		),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckDbtCloudSnowflakeCredentialExists(
				"dbtcloud_snowflake_credential.test_credential",
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_snowflake_credential.test_credential",
				"database",
				database,
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_snowflake_credential.test_credential",
				"role",
				role,
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_snowflake_credential.test_credential",
				"warehouse",
				warehouse,
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_snowflake_credential.test_credential",
				"schema",
				schema,
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_snowflake_credential.test_credential",
				"user",
				user,
			),
		),
	}
}

func getModifyConfigTestStep(projectName, database, role, warehouse, schema, user, password string) resource.TestStep {
	return resource.TestStep{
		Config: testAccDbtCloudSnowflakeCredentialResourceBasicConfig(
			projectName,
			database,
			role,
			warehouse,
			schema,
			user,
			password,
		),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckDbtCloudSnowflakeCredentialExists(
				"dbtcloud_snowflake_credential.test_credential",
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_snowflake_credential.test_credential",
				"database",
				database,
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_snowflake_credential.test_credential",
				"role",
				role,
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_snowflake_credential.test_credential",
				"warehouse",
				warehouse,
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_snowflake_credential.test_credential",
				"schema",
				schema,
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_snowflake_credential.test_credential",
				"user",
				user,
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_snowflake_credential.test_credential",
				"password",
				password,
			),
		),
	}
}

func TestBasicConfigConformance(t *testing.T) {
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	database := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	role := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	warehouse := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	schema := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	user := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	password := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest_helper.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckDbtCloudSnowflakeCredentialDestroy,
		Steps: []resource.TestStep{
			acctest_helper.MakeExternalProviderTestStep(getBasicConfigTestStep(projectName, database, role, warehouse, schema, user, password), acctest_config.LAST_VERSION_BEFORE_FRAMEWORK_MIGRATION),
			acctest_helper.MakeCurrentProviderNoOpTestStep(getBasicConfigTestStep(projectName, database, role, warehouse, schema, user, password)),
		},
	})
}

func TestModifyConfigConformance(t *testing.T) {
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	database2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	role2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	warehouse2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	schema2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	user2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	password2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest_helper.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckDbtCloudSnowflakeCredentialDestroy,
		Steps: []resource.TestStep{
			acctest_helper.MakeExternalProviderTestStep(getModifyConfigTestStep(projectName, database2, role2, warehouse2, schema2, user2, password2), acctest_config.LAST_VERSION_BEFORE_FRAMEWORK_MIGRATION),
			acctest_helper.MakeCurrentProviderNoOpTestStep(getModifyConfigTestStep(projectName, database2, role2, warehouse2, schema2, user2, password2)),
		},
	})
}

func TestAccDbtCloudSnowflakeCredentialResource(t *testing.T) {

	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	database := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	role := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	warehouse := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	schema := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	user := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	password := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	database2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	role2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	warehouse2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	schema2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	user2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	password2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	privateKey := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	privateKeyPassphrase := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudSnowflakeCredentialDestroy,
		Steps: []resource.TestStep{
			getBasicConfigTestStep(projectName, database, role, warehouse, schema, user, password),
			// RENAME
			// MODIFY
			getModifyConfigTestStep(projectName, database2, role2, warehouse2, schema2, user2, password2),
			// IMPORT
			{
				ResourceName:            "dbtcloud_snowflake_credential.test_credential",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "private_key", "private_key_passphrase"},
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudSnowflakeCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudSnowflakeCredentialResourceBasicPrivateKeyConfig(
					projectName,
					database,
					role,
					warehouse,
					schema,
					user,
					privateKey,
					privateKeyPassphrase,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudSnowflakeCredentialExists(
						"dbtcloud_snowflake_credential.test_credential_p",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_credential.test_credential_p",
						"database",
						database,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_credential.test_credential_p",
						"role",
						role,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_credential.test_credential_p",
						"warehouse",
						warehouse,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_credential.test_credential_p",
						"schema",
						schema,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_credential.test_credential_p",
						"user",
						user,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_credential.test_credential_p",
						"private_key",
						privateKey,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_credential.test_credential_p",
						"private_key_passphrase",
						privateKeyPassphrase,
					),
				),
			},
			// RENAME
			// MODIFY
			// IMPORT
			{
				ResourceName:            "dbtcloud_snowflake_credential.test_credential_p",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_key", "private_key_passphrase"},
			},
		},
	})

}

func testAccDbtCloudSnowflakeCredentialResourceBasicConfig(
	projectName, database, role, warehouse, schema, user, password string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}
resource "dbtcloud_snowflake_credential" "test_credential" {
    is_active = true
    project_id = dbtcloud_project.test_project.id
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

func testAccDbtCloudSnowflakeCredentialResourceBasicPrivateKeyConfig(
	projectName, database, role, warehouse, schema, user, private_key, private_key_passphrase string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}
resource "dbtcloud_snowflake_credential" "test_credential_p" {
    is_active = true
    project_id = dbtcloud_project.test_project.id
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

		projectId, credentialId, err := helper.SplitIDToInts(
			rs.Primary.ID,
			"dbtcloud_snowflake_credential",
		)
		if err != nil {
			return err
		}

		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
		_, err = apiClient.GetSnowflakeCredential(projectId, credentialId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudSnowflakeCredentialDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_snowflake_credential" {
			continue
		}
		projectId, credentialId, err := helper.SplitIDToInts(
			rs.Primary.ID,
			"dbtcloud_snowflake_credential",
		)
		if err != nil {
			return err
		}

		_, err = apiClient.GetSnowflakeCredential(projectId, credentialId)
		if err == nil {
			return fmt.Errorf("Snowflake credential still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
