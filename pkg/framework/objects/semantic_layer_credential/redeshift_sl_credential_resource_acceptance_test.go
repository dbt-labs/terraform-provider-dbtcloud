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

func TestDbtCloudSemanticLayerConfigurationRedshiftResource(t *testing.T) {

	_, _, projectID := acctest_helper.GetSemanticLayerConfigTestingConfigurations()
	if projectID == 0 {
		t.Skip("Skipping test because config is not set")
	}

	name := acctest.RandomWithPrefix("redshift_tf_test")
	name2 := acctest.RandomWithPrefix("redshift_tf_test_2")

	username := acctest.RandString(10)
	password := acctest.RandString(10)

	username2 := acctest.RandString(10)
	password2 := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudSemanticLayerCredentialRedshiftDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudRedshiftSemanticLayerCredentialResource(
					projectID,
					name,
					username,
					password,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_redshift_semantic_layer_credential.test_redshift_semantic_layer_credential",
						"configuration.project_id",
						strconv.Itoa(projectID),
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_redshift_semantic_layer_credential.test_redshift_semantic_layer_credential",
						"configuration.name",
						name,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_redshift_semantic_layer_credential.test_redshift_semantic_layer_credential",
						"project_id",
						strconv.Itoa(projectID),
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_redshift_semantic_layer_credential.test_redshift_semantic_layer_credential",
						"username",
						username,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_redshift_semantic_layer_credential.test_redshift_semantic_layer_credential",
						"password",
						password,
					),
				),
			},

			// Update name, username, password
			{
				Config: testAccDbtCloudRedshiftSemanticLayerCredentialResource(
					projectID,
					name2,
					username2,
					password2,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_redshift_semantic_layer_credential.test_redshift_semantic_layer_credential",
						"configuration.project_id",
						strconv.Itoa(projectID),
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_redshift_semantic_layer_credential.test_redshift_semantic_layer_credential",
						"configuration.name",
						name2,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_redshift_semantic_layer_credential.test_redshift_semantic_layer_credential",
						"project_id",
						strconv.Itoa(projectID),
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_redshift_semantic_layer_credential.test_redshift_semantic_layer_credential",
						"username",
						username2,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_redshift_semantic_layer_credential.test_redshift_semantic_layer_credential",
						"password",
						password2,
					),
				),
			},
		},
	})
}

// builds a terraform config for dbtcloud_redshift_semantic_layer_credential resource
func testAccDbtCloudRedshiftSemanticLayerCredentialResource(
	projectID int,
	name string,
	username string,
	password string,

) string {

	return fmt.Sprintf(`
	
resource "dbtcloud_redshift_semantic_layer_credential" "test_redshift_semantic_layer_credential" {
  configuration = {
    project_id = %s
	name = "%s"
	adapter_version = "redshift_v0"
  }
  credential = {
  	project_id = %s
	username = "%s"
	dataset = "test"
	is_active = true
	password = "%s"
	num_threads = 3
	default_schema = "test"
  }
	project_id = %d
	username = "%s"
	password = "%s"
  
}`, strconv.Itoa(projectID), name, strconv.Itoa(projectID), username, password, projectID, username, password)
}

func testAccCheckDbtCloudSemanticLayerCredentialRedshiftDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_redshift_semantic_layer_credential" {
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
