package scim_config_token_test

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

// TestAccDbtCloudSCIMConfigTokenResource tests the full lifecycle of a SCIM
// config token: create, verify token_string is populated, name change forces
// replace, import (with token_string absent as expected), and destroy.
func TestAccDbtCloudSCIMConfigTokenResource(t *testing.T) {
	tokenName := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	tokenNameUpdated := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudSCIMConfigTokenDestroy,
		Steps: []resource.TestStep{
			// Create — verify token_string is set (write-once, sensitive)
			{
				Config: testAccDbtCloudSCIMConfigTokenResourceConfig(tokenName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_scim_config_token.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_scim_config_token.test",
						"name",
						tokenName,
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_scim_config_token.test",
						"token_string",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_scim_config_token.test",
						"created_at",
					),
				),
			},
			// Update name — forces replace (RequiresReplace), new token_string is issued
			{
				Config: testAccDbtCloudSCIMConfigTokenResourceConfig(tokenNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_scim_config_token.test",
						"name",
						tokenNameUpdated,
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_scim_config_token.test",
						"token_string",
					),
				),
			},
			// Import — token_string is not returned by the API after creation,
			// so it will be absent in the imported state. This is expected and documented.
			{
				ResourceName:            "dbtcloud_scim_config_token.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token_string"},
			},
		},
	})
}

// testAccCheckDbtCloudSCIMConfigTokenDestroy verifies the token no longer
// exists in the API after the resource is destroyed.
func testAccCheckDbtCloudSCIMConfigTokenDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("issue getting the client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_scim_config_token" {
			continue
		}

		tokenID, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
		if err != nil {
			return fmt.Errorf("could not parse SCIM config token ID %q: %w", rs.Primary.ID, err)
		}

		_, err = apiClient.GetSCIMConfigToken(tokenID)
		if err == nil {
			return fmt.Errorf("SCIM config token %d still exists after destroy", tokenID)
		}

		notFoundErr := regexp.MustCompile("resource-not-found")
		if !notFoundErr.MatchString(err.Error()) {
			return fmt.Errorf("unexpected error checking SCIM config token destroy: %w", err)
		}
	}

	return nil
}

func testAccDbtCloudSCIMConfigTokenResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "dbtcloud_scim_config_token" "test" {
  name = %q
}
`, name)
}
