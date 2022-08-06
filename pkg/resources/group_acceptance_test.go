package resources_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDbtCloudGroupResource(t *testing.T) {

	groupName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudGroupResourceBasicConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudGroupExists("dbt_cloud_group.test_group"),
					resource.TestCheckResourceAttr("dbt_cloud_group.test_group", "name", groupName),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbt_cloud_group.test_group",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccDbtCloudGroupResourceBasicConfig(groupName string) string {
	return fmt.Sprintf(`
resource "dbt_cloud_group" "test_group" {
    name = "%s"
}
`, groupName)
}

func testAccCheckDbtCloudGroupExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		apiClient := testAccProvider.Meta().(*dbt_cloud.Client)
		groupID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Can't get groupID")
		}
		_, err = apiClient.GetGroup(groupID)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudGroupDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbt_cloud_group" {
			continue
		}
		groupID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Can't get groupID")
		}
		_, err = apiClient.GetGroup(groupID)
		if err == nil {
			return fmt.Errorf("Group still exists")
		}
		notFoundErr := "not found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
