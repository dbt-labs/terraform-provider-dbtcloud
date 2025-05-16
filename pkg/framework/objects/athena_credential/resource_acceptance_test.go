package athena_credential_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudAthenaCredentialResource(t *testing.T) {
	projectName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schema := "test_schema"
	awsAccessKeyID := "test_access_key_id"
	awsSecretAccessKey := "test_secret_access_key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudAthenaCredentialDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDbtCloudAthenaCredentialResourceConfig(
					projectName,
					schema,
					awsAccessKeyID,
					awsSecretAccessKey,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDbtCloudAthenaCredentialExists("dbtcloud_athena_credential.test"),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_athena_credential.test",
						"id",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_athena_credential.test",
						"credential_id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_athena_credential.test",
						"schema",
						schema,
					),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dbtcloud_athena_credential.test",
				ImportState:       true,
				ImportStateVerify: true,
				// These fields can't be read from the API
				ImportStateVerifyIgnore: []string{
					"aws_access_key_id",
					"aws_secret_access_key",
				},
			},
			// Update and Read testing
			{
				Config: testAccDbtCloudAthenaCredentialResourceConfig(
					projectName,
					"updated_schema",
					awsAccessKeyID,
					awsSecretAccessKey,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDbtCloudAthenaCredentialExists("dbtcloud_athena_credential.test"),
					resource.TestCheckResourceAttr(
						"dbtcloud_athena_credential.test",
						"schema",
						"updated_schema",
					),
				),
			},
		},
	})
}

func testAccDbtCloudAthenaCredentialResourceConfig(
	projectName string,
	schema string,
	awsAccessKeyID string,
	awsSecretAccessKey string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test" {
  name = "%s"
}

resource "dbtcloud_athena_credential" "test" {
  project_id           = dbtcloud_project.test.id
  schema               = "%s"
  aws_access_key_id    = "%s"
  aws_secret_access_key = "%s"
}
`, projectName, schema, awsAccessKeyID, awsSecretAccessKey)
}

func testAccCheckDbtCloudAthenaCredentialExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		projectID, credentialID, err := helper.SplitIDToInts(rs.Primary.ID, "dbtcloud_athena_credential")
		if err != nil {
			return err
		}

		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
		_, err = apiClient.GetAthenaCredential(projectID, credentialID)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudAthenaCredentialDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_athena_credential" {
			continue
		}
		projectID, credentialID, err := helper.SplitIDToInts(rs.Primary.ID, "dbtcloud_athena_credential")
		if err != nil {
			return err
		}

		_, err = apiClient.GetAthenaCredential(projectID, credentialID)
		if err == nil {
			return fmt.Errorf("Athena credential still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
