package bigquery_credential_test

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

var projectName = strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
var dataset = strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

var createCredentialTestStep = resource.TestStep{
	Config: testAccDbtCloudBigQueryCredentialResourceBasicConfig(projectName, dataset),
	Check: resource.ComposeTestCheckFunc(
		testAccCheckDbtCloudBigQueryCredentialExists(
			"dbtcloud_bigquery_credential.test_credential",
		),
		resource.TestCheckResourceAttr(
			"dbtcloud_bigquery_credential.test_credential",
			"dataset",
			dataset,
		),
	),
}

func TestAccDbtCloudBigQueryCredentialResource(t *testing.T) {
	var importStateTestStep = resource.TestStep{
		ResourceName:            "dbtcloud_bigquery_credential.test_credential",
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"password"},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudBigQueryCredentialDestroy,
		Steps: []resource.TestStep{
			createCredentialTestStep,
			// RENAME
			// MODIFY
			importStateTestStep,
		},
	})
}

func testAccDbtCloudBigQueryCredentialResourceBasicConfig(projectName, dataset string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}
resource "dbtcloud_bigquery_credential" "test_credential" {
    is_active = true
    project_id = dbtcloud_project.test_project.id
    dataset = "%s"
    num_threads = 3
}
`, projectName, dataset)
}

func testAccCheckDbtCloudBigQueryCredentialExists(resource string) resource.TestCheckFunc {
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
			"dbtcloud_bigquery_credential",
		)
		if err != nil {
			return err
		}

		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
		_, err = apiClient.GetBigQueryCredential(projectId, credentialId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudBigQueryCredentialDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_bigquery_credential" {
			continue
		}
		projectId, credentialId, err := helper.SplitIDToInts(
			rs.Primary.ID,
			"dbtcloud_bigquery_credential",
		)
		if err != nil {
			return err
		}

		_, err = apiClient.GetBigQueryCredential(projectId, credentialId)
		if err == nil {
			return fmt.Errorf("BigQuery credential still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
