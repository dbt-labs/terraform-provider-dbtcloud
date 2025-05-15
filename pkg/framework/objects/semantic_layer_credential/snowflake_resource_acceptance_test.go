package semantic_layer_credential_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudSemanticLayerConfigurationResource(t *testing.T) {

	_, _, projectID := acctest_helper.GetSemanticLayerConfigTestingConfigurations()
	if projectID == 0 {
		t.Skip("Skipping test because config is not set")
	}

	name := acctest.RandomWithPrefix("tf_test")

	authMethod := "password"
	user := acctest.RandString(10)
	password := acctest.RandString(10)
	role := acctest.RandString(10)

	warehouse := acctest.RandString(10)

	//update config fields
	name2 := acctest.RandomWithPrefix("tf_test2")
	warehouse2 := acctest.RandString(10)

	//update password auth fields
	user2 := acctest.RandString(10)
	password2 := acctest.RandString(10)
	role2 := acctest.RandString(10)

	//update auth method
	authMethod2 := "keypair"
	privateKey := acctest.RandString(10)
	privateKeyPassphrase := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudSemanticLayerCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudSnowflakeSemanticLayerCredentialResourcePasswordAuth(
					projectID,
					name,
					authMethod,
					role,
					warehouse,
					user,
					password,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_semantic_layer_credential.test_snowflake_semantic_layer_credential",
						"configuration.name",
						name,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_semantic_layer_credential.test_snowflake_semantic_layer_credential",
						"credential.auth_type",
						authMethod,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_semantic_layer_credential.test_snowflake_semantic_layer_credential",
						"credential.user",
						user,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_semantic_layer_credential.test_snowflake_semantic_layer_credential",
						"credential.password",
						password,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_semantic_layer_credential.test_snowflake_semantic_layer_credential",
						"credential.role",
						role,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_semantic_layer_credential.test_snowflake_semantic_layer_credential",
						"credential.warehouse",
						warehouse,
					),
				),
			},
			// MODIFY general config fields
			{
				Config: testAccDbtCloudSnowflakeSemanticLayerCredentialResourcePasswordAuth(
					projectID,
					name2,
					authMethod,
					role,
					warehouse2,
					user,
					password,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_semantic_layer_credential.test_snowflake_semantic_layer_credential",
						"configuration.name",
						name2,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_semantic_layer_credential.test_snowflake_semantic_layer_credential",
						"credential.warehouse",
						warehouse2,
					),
				),
			},

			// MODIFY password auth fields
			{
				Config: testAccDbtCloudSnowflakeSemanticLayerCredentialResourcePasswordAuth(
					projectID,
					name2,
					authMethod,
					role2,
					warehouse2,
					user2,
					password2,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_semantic_layer_credential.test_snowflake_semantic_layer_credential",
						"credential.user",
						user2,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_semantic_layer_credential.test_snowflake_semantic_layer_credential",
						"credential.password",
						password2,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_semantic_layer_credential.test_snowflake_semantic_layer_credential",
						"credential.role",
						role2,
					),
				),
			},

			// MODIFY auth method and fields
			{
				Config: testAccDbtCloudSnowflakeSemanticLayerCredentialResourcePrivateKeyAuth(
					projectID,
					name2,
					authMethod2,
					role2,
					warehouse2,
					privateKey,
					privateKeyPassphrase,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_semantic_layer_credential.test_snowflake_semantic_layer_credential",
						"credential.auth_type",
						authMethod2,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_semantic_layer_credential.test_snowflake_semantic_layer_credential",
						"credential.private_key",
						privateKey,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_snowflake_semantic_layer_credential.test_snowflake_semantic_layer_credential",
						"credential.private_key_passphrase",
						privateKeyPassphrase,
					),
				),
			},
		},
	})
}

func testAccDbtCloudSnowflakeSemanticLayerCredentialResourcePasswordAuth(
	projectID int,
	name string,
	auth string,
	role string,
	warehouse string,
	user string,
	password string,
) string {

	return fmt.Sprintf(`
	
resource "dbtcloud_snowflake_semantic_layer_credential" "test_snowflake_semantic_layer_credential" {
  configuration = {
    project_id = %s
	name = "%s"
	adapter_version = "snowflake_v0"
  }
  credential = {
  	project_id = %s
	is_active = true
	auth_type = "%s"
	role = "%s"
	warehouse = "%s"
	user = "%s"
	password = "%s"
	num_threads = 3
	semantic_layer_credential = true
  }
}`, strconv.Itoa(projectID), name, strconv.Itoa(projectID), auth, role, warehouse, user, password)
}

func testAccDbtCloudSnowflakeSemanticLayerCredentialResourcePrivateKeyAuth(
	projectID int,
	name string,
	auth string,
	role string,
	warehouse string,
	privateKey string,
	privateKeyPassphrase string,
) string {

	return fmt.Sprintf(`
	
resource "dbtcloud_snowflake_semantic_layer_credential" "test_snowflake_semantic_layer_credential" {
  configuration = {
    project_id = %s
	name = "%s"
	adapter_version = "snowflake_v0"
  }
  credential = {
  	project_id = %s
	is_active = true
	auth_type = "%s"
	role = "%s"
	warehouse = "%s"
	private_key= "%s"
	private_key_passphrase = "%s"
	num_threads = 3
	semantic_layer_credential = true
  }
}`, strconv.Itoa(projectID), name, strconv.Itoa(projectID), auth, role, warehouse, privateKey, privateKeyPassphrase)
}

func testAccCheckDbtCloudSemanticLayerCredentialDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_snowflake_semantic_layer_credential" {
			continue
		}

		id, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert resource ID to int64: %s", err)
		}
		_, err = apiClient.GetSemanticLayerCredential(id)
		if err == nil {
			return fmt.Errorf("Semantic Layer Configuration still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
