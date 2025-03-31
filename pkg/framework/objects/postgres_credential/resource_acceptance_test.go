package postgres_credential_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_config"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var projectName = strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
var default_schema = strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
var username = strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
var password = strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

var createCredentialTestStep = resource.TestStep{
	Config: testAccDbtCloudPostgresCredentialResourceBasicConfig(
		projectName,
		default_schema,
		username,
		password,
	),
	Check: resource.ComposeTestCheckFunc(
		testAccCheckDbtCloudPostgresCredentialExists(
			"dbtcloud_postgres_credential.test_credential",
		),
		resource.TestCheckResourceAttr(
			"dbtcloud_postgres_credential.test_credential",
			"default_schema",
			default_schema,
		),
		resource.TestCheckResourceAttr(
			"dbtcloud_postgres_credential.test_credential",
			"username",
			username,
		),
		resource.TestCheckResourceAttr(
			"dbtcloud_postgres_credential.test_credential",
			"target_name",
			"default",
		),
		resource.TestCheckResourceAttr(
			"dbtcloud_postgres_credential.test_credential",
			"type",
			"postgres",
		),
	),
}

func TestAccDbtCloudPostgresCredentialResource(t *testing.T) {
	var importStateTestStep = resource.TestStep{
		ResourceName:            "dbtcloud_postgres_credential.test_credential",
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"password"},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudPostgresCredentialDestroy,
		Steps: []resource.TestStep{
			createCredentialTestStep,
			// RENAME
			// MODIFY
			importStateTestStep,
		},
	})

}

func TestConfDbtCloudPostgresCredentialResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest_helper.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckDbtCloudPostgresCredentialDestroy,
		Steps: []resource.TestStep{
			acctest_helper.MakeExternalProviderTestStep(createCredentialTestStep, acctest_config.LAST_VERSION_BEFORE_FRAMEWORK_MIGRATION),
			acctest_helper.MakeCurrentProviderNoOpTestStep(createCredentialTestStep),
		},
	})
}

func testAccDbtCloudPostgresCredentialResourceBasicConfig(
	projectName, default_schema, username, password string,
) string {
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
		projectId, credentialId, err := helper.SplitIDToInts(
			rs.Primary.ID,
			"dbtcloud_postgres_credential",
		)
		if err != nil {
			return err
		}

		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
		_, err = apiClient.GetPostgresCredential(projectId, credentialId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudPostgresCredentialDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_postgres_credential" {
			continue
		}
		projectId, credentialId, err := helper.SplitIDToInts(
			rs.Primary.ID,
			"dbtcloud_postgres_credential",
		)
		if err != nil {
			return err
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
