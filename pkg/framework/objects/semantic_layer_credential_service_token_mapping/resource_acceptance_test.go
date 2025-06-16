package semantic_layer_credential_service_token_mapping_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudSLCredentialServiceTokenMappingResource(t *testing.T) {

	serviceTokenName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	projectName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudServiceTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudSLCredServiceTokenMappingResourceConfig(
					serviceTokenName,
					projectName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_semantic_layer_credential_service_token_mapping.test_mapping",
						"id",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_semantic_layer_credential_service_token_mapping.test_mapping",
						"semantic_layer_credential_id",
					),
					resource.TestCheckResourceAttrPair(
						"dbtcloud_redshift_semantic_layer_credential.test",
						"id",
						"dbtcloud_semantic_layer_credential_service_token_mapping.test_mapping",
						"semantic_layer_credential_id",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_semantic_layer_credential_service_token_mapping.test_mapping",
						"service_token_id",
					),
					resource.TestCheckResourceAttrPair(
						"dbtcloud_service_token.test_service_token",
						"id",
						"dbtcloud_semantic_layer_credential_service_token_mapping.test_mapping",
						"service_token_id",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_semantic_layer_credential_service_token_mapping.test_mapping",
						"project_id",
					),
					resource.TestCheckResourceAttrPair(
						"dbtcloud_project.test_project",
						"id",
						"dbtcloud_semantic_layer_credential_service_token_mapping.test_mapping",
						"project_id",
					),
				),
			},
		},
	})
}

func testAccDbtCloudSLCredServiceTokenMappingResourceConfig(serviceTokenName, projectName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}
resource "dbtcloud_service_token" "test_service_token" {
    name = "%s"
    service_token_permissions {
        permission_set = "git_admin"
        all_projects = false
        project_id = dbtcloud_project.test_project.id
    }
    service_token_permissions {
        permission_set = "job_admin"
        all_projects = true
    }
    service_token_permissions {
        permission_set = "developer"
        all_projects = true
    }
}
resource "dbtcloud_redshift_semantic_layer_credential" "test" {
  configuration = {
    project_id = dbtcloud_project.test_project.id
	name = "CredentialName"
	adapter_version = "redshift_v0"
  }
  credential = {
  	project_id = dbtcloud_project.test_project.id
	username = "user"
	dataset = "test"
	is_active = true
	password = "password"
	num_threads = 4
	default_schema = "dataset"
  }
}

resource "dbtcloud_semantic_layer_credential_service_token_mapping" "test_mapping" {
  semantic_layer_credential_id = dbtcloud_redshift_semantic_layer_credential.test.id
  service_token_id = dbtcloud_service_token.test_service_token.id
  project_id = dbtcloud_project.test_project.id
}
`, projectName, serviceTokenName)
}

func testAccCheckDbtCloudServiceTokenDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_semantic_layer_credential_service_token_mapping" {
			continue
		}

		idInt, _ := strconv.Atoi(rs.Primary.Attributes["id"])
		cred_id, _ := strconv.Atoi(rs.Primary.Attributes["semantic_layer_credential_id"])
		token_id, _ := strconv.Atoi(rs.Primary.Attributes["service_token_id"])
		project_id, _ := strconv.Atoi(rs.Primary.Attributes["project_id"])

		sm := dbt_cloud.SemanticLayerCredentialServiceTokenMapping{
			ID:                        &idInt,
			SemanticLayerCredentialID: cred_id,
			ServiceTokenID:            token_id,
			ProjectID:                 project_id,
		}

		_, err = apiClient.GetSemanticLayerCredentialServiceTokenMapping(sm)
		if err == nil {
			return fmt.Errorf("Mapping still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
