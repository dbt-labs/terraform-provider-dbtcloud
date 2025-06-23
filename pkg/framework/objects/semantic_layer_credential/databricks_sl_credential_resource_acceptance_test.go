package semantic_layer_credential_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudDatabricksSemanticLayerConfigurationResource(t *testing.T) {

	_, _, projectID := acctest_helper.GetSemanticLayerConfigTestingConfigurations()
	if projectID == 0 {
		t.Skip("Skipping test because config is not set")
	}

	name := acctest.RandomWithPrefix("databricks_cred_name")
	name2 := acctest.RandomWithPrefix("databricks_cred_name")
	adapterVersion := "databricks_v0"
	catalog := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	catalog2 := ""
	token := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	token2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudDatabricksSemanticLayerCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudDatabricksSemanticLayerCredentialResource(
					projectID,
					name,
					adapterVersion,
					catalog,
					token,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_databricks_semantic_layer_credential.test",
						"configuration.name",
						name,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_databricks_semantic_layer_credential.test",
						"configuration.adapter_version",
						adapterVersion,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_databricks_semantic_layer_credential.test",
						"credential.catalog",
						catalog,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_databricks_semantic_layer_credential.test",
						"credential.token",
						token,
					),
				),
			},
			// MODIFY general config fields
			{
				Config: testAccDbtCloudDatabricksSemanticLayerCredentialResource(
					projectID,
					name2,
					adapterVersion,
					catalog2,
					token2,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_databricks_semantic_layer_credential.test",
						"configuration.name",
						name2,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_databricks_semantic_layer_credential.test",
						"credential.catalog",
						catalog2,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_databricks_semantic_layer_credential.test",
						"credential.token",
						token2,
					),
				),
			},
		},
	})
}

func testAccDbtCloudDatabricksSemanticLayerCredentialResource(
	projectID int,
	name string,
	adapterVersion string,
	catalog string,
	token string,
) string {

	return fmt.Sprintf(`
	
resource "dbtcloud_databricks_semantic_layer_credential" "test" {
configuration = {
    project_id = %s
	  name = "%s"
	  adapter_version = "%s"
}
credential = {
  	project_id = "%s"
    catalog = "%s"
    token = "%s"
	semantic_layer_credential = true
  }
}`, strconv.Itoa(projectID), name, adapterVersion, strconv.Itoa(projectID), catalog, token)
}

func testAccCheckDbtCloudDatabricksSemanticLayerCredentialDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_databricks_semantic_layer_credential" {
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
