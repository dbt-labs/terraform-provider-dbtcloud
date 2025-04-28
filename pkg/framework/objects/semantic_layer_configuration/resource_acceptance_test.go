package semantic_layer_configuration_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudSemanticLayerConfigurationResource(t *testing.T) {

	environmentID, environmentID2, projectID := acctest_helper.GetSemanticLayerConfigTestingConfigurations()
	if environmentID == 0 || environmentID2 == 0 || projectID == 0 {
		t.Skip("Skipping test because config is not set")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudSemanticLayerConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudSemanticLayerConfigurationResourceBasicConfig(
					projectID,
					environmentID,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_semantic_layer_configuration.test_semantic_layer_configuration",
						"environment_id",
						strconv.Itoa(environmentID),
					),
				),
			},
			// MODIFY
			{
				Config: testAccDbtCloudSemanticLayerConfigurationResourceBasicConfig(
					projectID,
					environmentID2,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_semantic_layer_configuration.test_semantic_layer_configuration",
						"environment_id",
						strconv.Itoa(environmentID2),
					),
				),
			},
		},
	})
}

func testAccDbtCloudSemanticLayerConfigurationResourceBasicConfig(
	projectID,
	environmentID int,

) string {

	return fmt.Sprintf(`
	
resource "dbtcloud_semantic_layer_configuration" "test_semantic_layer_configuration" {
    project_id = %s
	environment_id = "%s"
}`, strconv.Itoa(projectID), strconv.Itoa(environmentID))
}

func testAccCheckDbtCloudSemanticLayerConfigurationDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_semantic_layer_configuration" {
			continue
		}
		projectIDStr := rs.Primary.Attributes["project_id"]
		projectID, err := strconv.Atoi(projectIDStr)
		if err != nil {
			return fmt.Errorf("failed to convert project_id to int: %s", err)
		}

		id, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert resource ID to int64: %s", err)
		}
		_, err = apiClient.GetSemanticLayerConfiguration(int64(projectID), id)
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
