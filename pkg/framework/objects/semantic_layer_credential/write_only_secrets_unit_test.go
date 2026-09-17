package semantic_layer_credential_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/testhelpers"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestSemanticLayerCredentials_WriteOnlySecrets(t *testing.T) {
	testCases := []struct {
		name           string
		resource       string
		adapterVersion string
		readValues     map[string]interface{}
		createSecrets  map[string]string
		updateSecrets  map[string]string
		writeOnlyAttrs []string
		config         func(string, int) string
	}{
		{
			name:           "snowflake keypair",
			resource:       "dbtcloud_snowflake_semantic_layer_credential.test",
			adapterVersion: "snowflake_v0",
			readValues: map[string]interface{}{
				"auth_type": "keypair",
				"role":      "test-role",
				"warehouse": "test-warehouse",
			},
			createSecrets: map[string]string{
				"private_key":            "create-private-key",
				"private_key_passphrase": "create-passphrase",
			},
			updateSecrets: map[string]string{
				"private_key":            "update-private-key",
				"private_key_passphrase": "update-passphrase",
			},
			writeOnlyAttrs: []string{"credential.private_key_wo", "credential.private_key_passphrase_wo"},
			config:         snowflakeWriteOnlyConfig,
		},
		{
			name:           "redshift password",
			resource:       "dbtcloud_redshift_semantic_layer_credential.test",
			adapterVersion: "redshift_v0",
			readValues:     map[string]interface{}{"username": "test-user"},
			createSecrets:  map[string]string{"password": "create-password"},
			updateSecrets:  map[string]string{"password": "update-password"},
			writeOnlyAttrs: []string{"credential.password_wo"},
			config:         redshiftWriteOnlyConfig,
		},
		{
			name:           "postgres password",
			resource:       "dbtcloud_postgres_semantic_layer_credential.test",
			adapterVersion: "postgres_v0",
			readValues:     map[string]interface{}{"username": "test-user"},
			createSecrets:  map[string]string{"password": "create-password"},
			updateSecrets:  map[string]string{"password": "update-password"},
			writeOnlyAttrs: []string{"credential.password_wo"},
			config:         postgresWriteOnlyConfig,
		},
		{
			name:           "databricks token",
			resource:       "dbtcloud_databricks_semantic_layer_credential.test",
			adapterVersion: "databricks_v0",
			readValues:     map[string]interface{}{"catalog": "test-catalog"},
			createSecrets:  map[string]string{"token": "create-token"},
			updateSecrets:  map[string]string{"token": "update-token"},
			writeOnlyAttrs: []string{"credential.token_wo"},
			config:         databricksWriteOnlyConfig,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("TF_ACC", "1")

			const credentialID = 42
			createPath := "/v3/accounts/1/semantic-layer-credentials/"
			credentialPath := fmt.Sprintf("/v3/accounts/1/semantic-layer-credentials/%d", credentialID)
			updatePath := credentialPath + "/"

			response := map[string]interface{}{
				"data": map[string]interface{}{
					"id":              credentialID,
					"account_id":      1,
					"project_id":      123,
					"name":            "test-credential",
					"adapter_version": testCase.adapterVersion,
					"schema_type":     "semantic_layer_credentials",
					"values":          testCase.readValues,
				},
				"status": map[string]interface{}{
					"code":       200,
					"is_success": true,
				},
			}

			updateCalls := 0
			handlers := map[string]testhelpers.MockEndpointHandler{
				"POST " + createPath: func(r *http.Request) (int, interface{}, error) {
					if err := requestContainsSecrets(r, testCase.createSecrets); err != nil {
						return http.StatusBadRequest, nil, err
					}
					return http.StatusOK, response, nil
				},
				"GET " + credentialPath: func(_ *http.Request) (int, interface{}, error) {
					return http.StatusOK, response, nil
				},
				"POST " + updatePath: func(r *http.Request) (int, interface{}, error) {
					if err := requestContainsSecrets(r, testCase.updateSecrets); err != nil {
						return http.StatusBadRequest, nil, err
					}
					updateCalls++
					return http.StatusOK, response, nil
				},
				"DELETE " + updatePath: func(_ *http.Request) (int, interface{}, error) {
					return http.StatusOK, response, nil
				},
			}

			server := testhelpers.SetupMockServer(t, handlers)
			defer server.Close()

			providerConfig := fmt.Sprintf(`
provider "dbtcloud" {
  host_url   = %q
  token      = "test-token"
  account_id = 1
}
`, server.URL)
			createChecks := writeOnlyStateChecks(testCase.resource, testCase.writeOnlyAttrs, "1")
			updateChecks := writeOnlyStateChecks(testCase.resource, testCase.writeOnlyAttrs, "2")

			resource.Test(t, resource.TestCase{
				IsUnitTest:               true,
				ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: providerConfig + testCase.config("create", 1),
						Check:  resource.ComposeTestCheckFunc(createChecks...),
					},
					{
						Config: providerConfig + testCase.config("update", 2),
						Check:  resource.ComposeTestCheckFunc(updateChecks...),
					},
				},
			})

			if updateCalls != 1 {
				t.Fatalf("expected one update request, got %d", updateCalls)
			}
		})
	}
}

func requestContainsSecrets(r *http.Request, expected map[string]string) error {
	var body struct {
		Values map[string]interface{} `json:"values"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return fmt.Errorf("decode request: %w", err)
	}
	for key, expectedValue := range expected {
		if body.Values[key] != expectedValue {
			return fmt.Errorf("request did not contain the configured %s", key)
		}
	}
	return nil
}

func writeOnlyStateChecks(resourceName string, attributes []string, version string) []resource.TestCheckFunc {
	checks := make([]resource.TestCheckFunc, 0, len(attributes)*2)
	for _, attribute := range attributes {
		checks = append(checks,
			resource.TestCheckNoResourceAttr(resourceName, attribute),
			resource.TestCheckResourceAttr(resourceName, attribute+"_version", version),
		)
	}
	return checks
}

func snowflakeWriteOnlyConfig(prefix string, version int) string {
	return fmt.Sprintf(`
resource "dbtcloud_snowflake_semantic_layer_credential" "test" {
  configuration = {
    project_id      = 123
    name            = "test-credential"
    adapter_version = "snowflake_v0"
  }
  credential = {
    project_id                        = 123
    is_active                        = true
    auth_type                        = "keypair"
    role                             = "test-role"
    warehouse                        = "test-warehouse"
    private_key_wo                   = "%s-private-key"
    private_key_wo_version           = %d
    private_key_passphrase_wo         = "%s-passphrase"
    private_key_passphrase_wo_version = %d
    num_threads                       = 3
    semantic_layer_credential         = true
  }
}
`, prefix, version, prefix, version)
}

func redshiftWriteOnlyConfig(prefix string, version int) string {
	return fmt.Sprintf(`
resource "dbtcloud_redshift_semantic_layer_credential" "test" {
  configuration = {
    project_id      = 123
    name            = "test-credential"
    adapter_version = "redshift_v0"
  }
  credential = {
    project_id          = 123
    username            = "test-user"
    password_wo         = "%s-password"
    password_wo_version = %d
    default_schema      = "test_schema"
    num_threads         = 3
  }
}
`, prefix, version)
}

func postgresWriteOnlyConfig(prefix string, version int) string {
	return fmt.Sprintf(`
resource "dbtcloud_postgres_semantic_layer_credential" "test" {
  configuration = {
    project_id      = 123
    name            = "test-credential"
    adapter_version = "postgres_v0"
  }
  credential = {
    project_id                = 123
    username                  = "test-user"
    password_wo               = "%s-password"
    password_wo_version       = %d
    semantic_layer_credential = true
  }
}
`, prefix, version)
}

func databricksWriteOnlyConfig(prefix string, version int) string {
	return fmt.Sprintf(`
resource "dbtcloud_databricks_semantic_layer_credential" "test" {
  configuration = {
    project_id      = 123
    name            = "test-credential"
    adapter_version = "databricks_v0"
  }
  credential = {
    project_id                = 123
    catalog                   = "test-catalog"
    token_wo                  = "%s-token"
    token_wo_version          = %d
    semantic_layer_credential = true
  }
}
`, prefix, version)
}
