package redshift_credential_test

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

var username = acctest.RandString(10)
var password = acctest.RandString(10)

var createCredentialTestStep = resource.TestStep{
	Config: testAccDbtCloudRedshiftCredentialResourceBasicConfig(projectName, username, password),
	Check: resource.ComposeTestCheckFunc(
		testAccCheckDbtCloudRedshiftCredentialExists(
			"dbtcloud_redshift_credential.test_credential",
		),
		resource.TestCheckResourceAttr(
			"dbtcloud_redshift_credential.test_credential",
			"username",
			username,
		),
		resource.TestCheckResourceAttr(
			"dbtcloud_redshift_credential.test_credential",
			"password",
			password,
		),
	),
}

func TestAccDbtCloudRedshiftCredentialResource(t *testing.T) {
	var importStateTestStep = resource.TestStep{
		ResourceName:            "dbtcloud_redshift_credential.test_credential",
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"password"},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudRedshiftCredentialDestroy,
		Steps: []resource.TestStep{
			createCredentialTestStep,
			importStateTestStep,
		},
	})
}

func testAccDbtCloudRedshiftCredentialResourceBasicConfig(projectName, username string, password string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}
resource "dbtcloud_redshift_credential" "test_credential" {
    is_active = true
    project_id = dbtcloud_project.test_project.id
    num_threads = 3
	default_schema = "test"
	username = "%s"
	password = "%s"
}
`, projectName, username, password)
}

func testAccCheckDbtCloudRedshiftCredentialExists(resource string) resource.TestCheckFunc {
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
			"dbtcloud_redshift_credential",
		)
		if err != nil {
			return err
		}

		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
		_, err = apiClient.GetRedshiftCredential(projectId, credentialId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func TestAccDbtCloudRedshiftCredentialResourceWriteOnly(t *testing.T) {
	t.Skip("Requires Terraform >= 1.11")

	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	username := acctest.RandString(10)
	password := acctest.RandString(10)
	password2 := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudRedshiftCredentialDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with write-only password
			{
				Config: testAccDbtCloudRedshiftCredentialWriteOnlyConfig(
					projectName, username, password, 1,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudRedshiftCredentialExists(
						"dbtcloud_redshift_credential.test_credential_wo",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_redshift_credential.test_credential_wo",
						"username",
						username,
					),
					// password_wo should not be in state
					resource.TestCheckNoResourceAttr(
						"dbtcloud_redshift_credential.test_credential_wo",
						"password_wo",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_redshift_credential.test_credential_wo",
						"password_wo_version",
						"1",
					),
				),
			},
			// Step 2: Update by incrementing version with new password
			{
				Config: testAccDbtCloudRedshiftCredentialWriteOnlyConfig(
					projectName, username, password2, 2,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudRedshiftCredentialExists(
						"dbtcloud_redshift_credential.test_credential_wo",
					),
					resource.TestCheckNoResourceAttr(
						"dbtcloud_redshift_credential.test_credential_wo",
						"password_wo",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_redshift_credential.test_credential_wo",
						"password_wo_version",
						"2",
					),
				),
			},
			// Step 3: Import
			{
				ResourceName:      "dbtcloud_redshift_credential.test_credential_wo",
				ImportState:        true,
				ImportStateVerify:  true,
				ImportStateVerifyIgnore: []string{
					"password",
					"password_wo", "password_wo_version",
				},
			},
		},
	})
}

func testAccDbtCloudRedshiftCredentialWriteOnlyConfig(
	projectName, username, passwordWo string, passwordWoVersion int,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}
resource "dbtcloud_redshift_credential" "test_credential_wo" {
    is_active    = true
    project_id   = dbtcloud_project.test_project.id
    num_threads  = 3
    default_schema = "test"
    username     = "%s"
    password_wo  = "%s"
    password_wo_version = %d
}
`, projectName, username, passwordWo, passwordWoVersion)
}

func testAccCheckDbtCloudRedshiftCredentialDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_redshift_credential" {
			continue
		}
		projectId, credentialId, err := helper.SplitIDToInts(
			rs.Primary.ID,
			"dbtcloud_redshift_credential",
		)
		if err != nil {
			return err
		}

		_, err = apiClient.GetRedshiftCredential(projectId, credentialId)
		if err == nil {
			return fmt.Errorf("Redshift credential still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
