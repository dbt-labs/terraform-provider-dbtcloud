package synapse_credential_test

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

func TestAccDbtCloudSynapseCredentialResource(t *testing.T) {

	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	user := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	password := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	clientId := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	tenantId := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	clientSecret := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	var basicConfigTestStep = getBasicConfigTestStep(projectName, user, password)

	var modifyConfigTestStep = getModifyConfigTestStep(projectName, clientId, tenantId, clientSecret)

	var importConfigTestStep = getImportConfigTestStep()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudSynapseCredentialDestroy,
		Steps: []resource.TestStep{
			basicConfigTestStep,
			modifyConfigTestStep,
			importConfigTestStep,
		},
	})

}

func getBasicConfigTestStep(projectName, user, password string) resource.TestStep {
	return resource.TestStep{
		Config: testAccDbtCloudSynapseCredentialResourceUserPassConfig(
			projectName,
			user,
			password,
		),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckDbtCloudSynapseCredentialExists(
				"dbtcloud_synapse_credential.test_credential",
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_synapse_credential.test_credential",
				"user",
				user,
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_synapse_credential.test_credential",
				"schema",
				"my_schema",
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_synapse_credential.test_credential",
				"schema_authorization",
				"sp",
			),
		),
	}
}

func getModifyConfigTestStep(projectName, clientId, tenantId, clientSecret string) resource.TestStep {
	return resource.TestStep{
		Config: testAccDbtCloudSynapseCredentialResourceServicePrincipalConfig(
			projectName,
			"synapse_v0",
			clientId,
			tenantId,
			clientSecret,
		),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckDbtCloudSynapseCredentialExists(
				"dbtcloud_synapse_credential.test_credential",
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_synapse_credential.test_credential",
				"client_id",
				clientId,
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_synapse_credential.test_credential",
				"tenant_id",
				tenantId,
			),
		),
	}
}

func getImportConfigTestStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:            "dbtcloud_synapse_credential.test_credential",
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"password", "client_secret", "schema_authorization", "user", "adapter_type", "adapter_version"},
		ImportStateCheck: func(s []*terraform.InstanceState) error {
			if len(s) != 1 {
				return fmt.Errorf("expected 1 state, got %d", len(s))
			}
			state := s[0]
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

func testAccDbtCloudSynapseCredentialResourceUserPassConfig(
	projectName, user, password string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_global_connection" "synapse" {
  name = "My Synapse connection"
  synapse = {
    project_id = dbtcloud_project.test_project.id
	name = "Synapse"
	database = "testdb"
	host = "example.com"
	port = 1234
  }
}

resource "dbtcloud_synapse_credential" "test_credential" {
    project_id = dbtcloud_project.test_project.id
    schema = "my_schema"
	authentication = "sql"
    user = "%s"
    password = "%s"
    schema_authorization = "sp"
	adapter_type = "synapse"
}
`, projectName, user, password)
}

func testAccDbtCloudSynapseCredentialResourceServicePrincipalConfig(
	projectName, adapterVersion, clientId, tenantId, clientSecret string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_global_connection" "synapse" {
  name = "My Synapse connection"
  synapse = {
    project_id = dbtcloud_project.test_project.id
	name = "Synapse"
	database = "testdb"
	host = "example.com"
	port = 1234
	adapter_version = "%s"
  }
}

resource "dbtcloud_synapse_credential" "test_credential" {
    project_id = dbtcloud_project.test_project.id
    schema = "my_schema_new"
	authentication = "ServicePrincipal"
    client_id = "%s"
    tenant_id = "%s"
    client_secret = "%s"
	adapter_type = "synapse"
	adapter_version = "synapse_v0"

}
`, projectName, adapterVersion, clientId, tenantId, clientSecret)
}

func testAccCheckDbtCloudSynapseCredentialExists(resource string) resource.TestCheckFunc {
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
			"dbtcloud_synapse_credential",
		)
		if err != nil {
			return err
		}

		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
		_, err = apiClient.GetSynapseCredential(projectId, credentialId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudSynapseCredentialDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_synapse_credential" {
			continue
		}
		projectId, credentialId, err := helper.SplitIDToInts(
			rs.Primary.ID,
			"dbtcloud_synapse_credential",
		)
		if err != nil {
			return err
		}

		_, err = apiClient.GetSynapseCredential(projectId, credentialId)
		if err == nil {
			return fmt.Errorf("Synapse credential still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
