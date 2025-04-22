package fabric_credential_test

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

func TestConformanceBasicConfig(t *testing.T) {
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	user := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	password := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest_helper.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckDbtCloudFabricCredentialDestroy,
		Steps: []resource.TestStep{
			acctest_helper.MakeExternalProviderTestStep(getBasicConfigTestStep(projectName, user, password), acctest_config.LAST_VERSION_BEFORE_FRAMEWORK_MIGRATION),
			acctest_helper.MakeCurrentProviderNoOpTestStep(getBasicConfigTestStep(projectName, user, password)),
		},
	})
}

func TestConformanceModifyConfig(t *testing.T) {
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	clientId := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	tenantId := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	clientSecret := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	// MODIFY: test that running commands in SDKv2 and then the same commands in Framework generates a NoOp plan
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest_helper.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckDbtCloudFabricCredentialDestroy,
		Steps: []resource.TestStep{
			acctest_helper.MakeExternalProviderTestStep(getModifyConfigTestStep(projectName, clientId, tenantId, clientSecret), acctest_config.LAST_VERSION_BEFORE_FRAMEWORK_MIGRATION),
			acctest_helper.MakeCurrentProviderNoOpTestStep(getModifyConfigTestStep(projectName, clientId, tenantId, clientSecret)),
		},
	})
}

func TestConformanceImportConfig(t *testing.T) {
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	user := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	password := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	// Import: test that running an import in SDKv2 and then the same thing in Framework generates a NoOp plan
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest_helper.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckDbtCloudFabricCredentialDestroy,
		Steps: []resource.TestStep{
			acctest_helper.MakeExternalProviderTestStep(getBasicConfigTestStep(projectName, user, password), acctest_config.LAST_VERSION_BEFORE_FRAMEWORK_MIGRATION),
			acctest_helper.MakeCurrentProviderNoOpTestStep(getBasicConfigTestStep(projectName, user, password)),
		},
	})
}

func TestAccDbtCloudFabricCredentialResource(t *testing.T) {

	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	user := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	password := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	clientId := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	tenantId := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	clientSecret := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	var basicConfigTestStep = getBasicConfigTestStep(projectName, user, password)

	var modifyConfigTestStep = getModifyConfigTestStep(projectName, clientId, tenantId, clientSecret)

	var importConfigTestStep = getImportConfigTestStep()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudFabricCredentialDestroy,
		Steps: []resource.TestStep{
			basicConfigTestStep,
			modifyConfigTestStep,
			importConfigTestStep,
		},
	})

}

func getBasicConfigTestStep(projectName, user, password string) resource.TestStep {
	return resource.TestStep{
		Config: testAccDbtCloudFabricCredentialResourceUserPassConfig(
			projectName,
			user,
			password,
		),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckDbtCloudFabricCredentialExists(
				"dbtcloud_fabric_credential.test_credential",
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_fabric_credential.test_credential",
				"user",
				user,
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_fabric_credential.test_credential",
				"schema",
				"my_schema",
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_fabric_credential.test_credential",
				"schema_authorization",
				"sp",
			),
		),
	}
}

func getModifyConfigTestStep(projectName, clientId, tenantId, clientSecret string) resource.TestStep {
	return resource.TestStep{
		Config: testAccDbtCloudFabricCredentialResourceServicePrincipalConfig(
			projectName,
			clientId,
			tenantId,
			clientSecret,
		),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckDbtCloudFabricCredentialExists(
				"dbtcloud_fabric_credential.test_credential",
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_fabric_credential.test_credential",
				"client_id",
				clientId,
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_fabric_credential.test_credential",
				"tenant_id",
				tenantId,
			),
		),
	}
}

func getImportConfigTestStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:            "dbtcloud_fabric_credential.test_credential",
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"password", "client_secret", "schema_authorization", "user"},
		ImportStateCheck: func(s []*terraform.InstanceState) error {
			if len(s) != 1 {
				return fmt.Errorf("expected 1 state, got %d", len(s))
			}
			state := s[0]
			if state.Attributes["adapter_id"] == "" {
				return fmt.Errorf("missing adapter_id in import state")
			}
			if state.Attributes["client_id"] == "" {
				return fmt.Errorf("missing client_id in import state")
			}
			if state.Attributes["tenant_id"] == "" {
				return fmt.Errorf("missing tenant_id in import state")
			}
			if state.Attributes["schema"] == "" {
				return fmt.Errorf("missing schema in import state")
			}

			return nil
		},
	}
}

func testAccDbtCloudFabricCredentialResourceUserPassConfig(
	projectName, user, password string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_fabric_credential" "test_credential" {
    project_id = dbtcloud_project.test_project.id
    schema = "my_schema"
    user = "%s"
    password = "%s"
    schema_authorization = "sp"
}
`, projectName, user, password)
}

func testAccDbtCloudFabricCredentialResourceServicePrincipalConfig(
	projectName, clientId, tenantId, clientSecret string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_fabric_credential" "test_credential" {
    project_id = dbtcloud_project.test_project.id
    schema = "my_schema_new"
    client_id = "%s"
    tenant_id = "%s"
    client_secret = "%s"
}
`, projectName, clientId, tenantId, clientSecret)
}

func testAccCheckDbtCloudFabricCredentialExists(resource string) resource.TestCheckFunc {
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
			"dbtcloud_fabric_credential",
		)
		if err != nil {
			return err
		}

		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
		_, err = apiClient.GetFabricCredential(projectId, credentialId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudFabricCredentialDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_fabric_credential" {
			continue
		}
		projectId, credentialId, err := helper.SplitIDToInts(
			rs.Primary.ID,
			"dbtcloud_fabric_credential",
		)
		if err != nil {
			return err
		}

		_, err = apiClient.GetFabricCredential(projectId, credentialId)
		if err == nil {
			return fmt.Errorf("Fabric credential still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
