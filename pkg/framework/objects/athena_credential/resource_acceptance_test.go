package athena_credential_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_config"
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

	var step1 = resource.TestStep{
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
	}

	var step2 = resource.TestStep{
		ResourceName:      "dbtcloud_athena_credential.test",
		ImportState:       true,
		ImportStateVerify: true,
		// These fields can't be read from the API
		ImportStateVerifyIgnore: []string{
			"aws_access_key_id",
			"aws_secret_access_key",
		},
	}

	var step3 = resource.TestStep{
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
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudAthenaCredentialDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			step1,
			// ImportState testing
			step2,
			// Update and Read testing
			step3,
		},
	})

	// test the Framework implementation
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudAthenaCredentialDestroy,
		Steps: []resource.TestStep{
			step1,
			step2,
			step3,
		},
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest_helper.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckDbtCloudAthenaCredentialDestroy,
		Steps: []resource.TestStep{
			acctest_helper.MakeExternalProviderTestStep(step1, acctest_config.LAST_VERSION_BEFORE_FRAMEWORK_MIGRATION),
			acctest_helper.MakeCurrentProviderNoOpTestStep(step1),
		},
	})

	// MODIFY: test that running commands in SDKv2 and then the same commands in Framework generates a NoOp plan
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest_helper.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckDbtCloudAthenaCredentialDestroy,
		Steps: []resource.TestStep{
			acctest_helper.MakeExternalProviderTestStep(step3, acctest_config.LAST_VERSION_BEFORE_FRAMEWORK_MIGRATION),
			acctest_helper.MakeCurrentProviderNoOpTestStep(step3),
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
