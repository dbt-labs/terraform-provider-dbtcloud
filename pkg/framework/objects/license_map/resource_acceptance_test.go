package license_map_test

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudLicenseMapResource(t *testing.T) {

	groupName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	groupName2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest_helper.TestAccPreCheck(t)
			sweepLicenseMaps(t, "developer")
		},
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudLicenseMapDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudLicenseMapResourceBasicConfig("developer", groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudLicenseMapExists("dbtcloud_license_map.test_license_map"),
					resource.TestCheckResourceAttr(
						"dbtcloud_license_map.test_license_map",
						"license_type",
						"developer",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_license_map.test_license_map",
						"sso_license_mapping_groups.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_license_map.test_license_map",
						"sso_license_mapping_groups.0",
						groupName,
					),
				),
			},
			// MODIFY
			{
				Config: testAccDbtCloudLicenseMapResourceBasicConfig(
					"developer",
					fmt.Sprintf(`%s","%s`, groupName, groupName2),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudLicenseMapExists("dbtcloud_license_map.test_license_map"),
					resource.TestCheckResourceAttr(
						"dbtcloud_license_map.test_license_map",
						"license_type",
						"developer",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_license_map.test_license_map",
						"sso_license_mapping_groups.#",
						"2",
					),
					resource.TestCheckTypeSetElemAttr(
						"dbtcloud_license_map.test_license_map",
						"sso_license_mapping_groups.*",
						groupName,
					),
					resource.TestCheckTypeSetElemAttr(
						"dbtcloud_license_map.test_license_map",
						"sso_license_mapping_groups.*",
						groupName2,
					),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_license_map.test_license_map",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccDbtCloudLicenseMapResourceBasicConfig(licenseType string, groupName string) string {
	return fmt.Sprintf(`

resource "dbtcloud_license_map" "test_license_map" {
    license_type       = "%s"
    sso_license_mapping_groups = ["%s"]
}
`, licenseType, groupName)
}

func testAccCheckDbtCloudLicenseMapExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
		licenseMapID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Can't get groupID")
		}
		_, err = apiClient.GetLicenseMap(licenseMapID)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudLicenseMapDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_license_map" {
			continue
		}
		licenseMapID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Can't get licenseMapID")
		}
		_, err = apiClient.GetLicenseMap(licenseMapID)
		if err == nil {
			return fmt.Errorf("License Map still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}

// licenseMapListResponse mirrors the list endpoint. The client's GetAllLicenseMaps goes
// through the paginated helper, which calls log.Fatal on a request error and would take
// the whole test binary down; this sweep reads the single page directly instead.
type licenseMapListResponse struct {
	Data []dbt_cloud.LicenseMap `json:"data"`
}

// sweepLicenseMaps deletes license maps left behind by an earlier run. The API allows one
// per license type per account, so a leftover makes every create for that type fail with
// "License map must be unique" until it is removed.
func sweepLicenseMaps(t *testing.T, licenseTypes ...string) {
	t.Helper()

	client, err := acctest_helper.SharedClient()
	if err != nil {
		t.Fatalf("Issue getting the client: %s", err)
	}

	body, err := client.GetEndpoint(
		fmt.Sprintf("%s/v3/accounts/%d/license-maps/", client.HostURL, client.AccountID),
	)
	if err != nil {
		// Best effort: the test below reports the real failure if a leftover is blocking it.
		t.Logf("could not list the existing license maps: %s", err)
		return
	}

	response := licenseMapListResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Logf("could not read the existing license maps: %s", err)
		return
	}

	for _, licenseMap := range response.Data {
		if licenseMap.ID == nil || !slices.Contains(licenseTypes, licenseMap.LicenseType) {
			continue
		}
		if err := client.DestroyLicenseMap(*licenseMap.ID); err != nil {
			t.Logf("could not delete the leftover license map %d: %s", *licenseMap.ID, err)
			continue
		}
		t.Logf("deleted license map %d left behind by an earlier run", *licenseMap.ID)
	}
}
