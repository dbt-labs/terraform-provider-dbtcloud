package resources_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDbtCloudPostgresCredentialResource(t *testing.T) {

	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	default_schema := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	username := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	password := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudPostgresCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudPostgresCredentialResourceBasicConfig(projectName, default_schema, username, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudPostgresCredentialExists("dbtcloud_postgres_credential.test_credential"),
					resource.TestCheckResourceAttr("dbtcloud_postgres_credential.test_credential", "default_schema", default_schema),
					resource.TestCheckResourceAttr("dbtcloud_postgres_credential.test_credential", "username", username),
					resource.TestCheckResourceAttr("dbtcloud_postgres_credential.test_credential", "target_name", "default"),
					resource.TestCheckResourceAttr("dbtcloud_postgres_credential.test_credential", "type", "postgres"),
				),
			},
			// RENAME
			// MODIFY
			// IMPORT
			{
				ResourceName:            "dbtcloud_postgres_credential.test_credential",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func testAccDbtCloudPostgresCredentialResourceBasicConfig(projectName, default_schema, username, password string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}
resource "dbtcloud_postgres_credential" "test_credential" {
    is_active = true
    project_id = dbtcloud_project.test_project.id
	type = "postgres"
    default_schema = "%s"
    username = "%s"
    password = "%s"
    num_threads = 3
}
`, projectName, default_schema, username, password)
}

func testAccCheckDbtCloudPostgresCredentialExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		projectId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0])
		if err != nil {
			return fmt.Errorf("Can't get projectId")
		}
		credentialId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1])
		if err != nil {
			return fmt.Errorf("Can't get credentialId")
		}

		apiClient := testAccProvider.Meta().(*dbt_cloud.Client)
		_, err = apiClient.GetPostgresCredential(projectId, credentialId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudPostgresCredentialDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_postgres_credential" {
			continue
		}
		projectId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0])
		if err != nil {
			return fmt.Errorf("Can't get projectId")
		}
		credentialId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1])
		if err != nil {
			return fmt.Errorf("Can't get credentialId")
		}

		_, err = apiClient.GetPostgresCredential(projectId, credentialId)
		if err == nil {
			return fmt.Errorf("Postgres credential still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
