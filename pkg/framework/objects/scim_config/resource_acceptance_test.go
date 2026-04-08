package scim_config_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccDbtCloudSCIMConfigResource tests the full lifecycle of the singleton
// dbtcloud_scim_config resource: create, update (toggle flags), import, destroy
// (which resets to defaults rather than deleting).
func TestAccDbtCloudSCIMConfigResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudSCIMConfigDestroy,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccDbtCloudSCIMConfigResourceConfig(true, false, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_scim_config.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_scim_config.test",
						"enabled",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_scim_config.test",
						"manual_updates_allowed",
						"false",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_scim_config.test",
						"scim_controlled_license_type",
						"false",
					),
				),
			},
			// Update — toggle all flags
			{
				Config: testAccDbtCloudSCIMConfigResourceConfig(true, true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_scim_config.test",
						"enabled",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_scim_config.test",
						"manual_updates_allowed",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_scim_config.test",
						"scim_controlled_license_type",
						"true",
					),
				),
			},
			// Import — singleton, import by account ID
			{
				ResourceName:      "dbtcloud_scim_config.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// testAccCheckDbtCloudSCIMConfigDestroy verifies that destroying the resource
// resets the SCIM config to its default state (SCIM disabled).
func testAccCheckDbtCloudSCIMConfigDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("issue getting the client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_scim_config" {
			continue
		}

		cfg, err := apiClient.GetSCIMConfig()
		if err != nil {
			return fmt.Errorf("error reading SCIM config after destroy: %w", err)
		}
		if cfg.Enabled {
			return fmt.Errorf("SCIM config still has enabled=true after destroy; expected reset to defaults")
		}
	}

	return nil
}

func testAccDbtCloudSCIMConfigResourceConfig(enabled, manualUpdatesAllowed, scimControlledLicenseType bool) string {
	return fmt.Sprintf(`
resource "dbtcloud_scim_config" "test" {
  enabled                      = %t
  manual_updates_allowed       = %t
  scim_controlled_license_type = %t
}
`, enabled, manualUpdatesAllowed, scimControlledLicenseType)
}
