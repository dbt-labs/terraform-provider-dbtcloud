package global_connection_test

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/testhelpers"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// Write-only values are null in the plan, so they have to be read from the config. This
// guards against sending an empty private key to the API (see #732): the key must be in
// the POST payload on create and in the PATCH payload when the version is bumped, and
// must never be written to the state.
//
// write-only attributes require Terraform >= 1.11
func TestGlobalConnectionBigQuery_PrivateKeyWriteOnlySentToAPI(t *testing.T) {
	originalTFAcc := os.Getenv("TF_ACC")
	os.Setenv("TF_ACC", "1")
	defer func() {
		if originalTFAcc == "" {
			os.Unsetenv("TF_ACC")
		} else {
			os.Setenv("TF_ACC", originalTFAcc)
		}
	}()

	const (
		accountID    = 1
		connectionID = 42
	)
	collectionPath := fmt.Sprintf("/v3/accounts/%d/connections/", accountID)
	itemPath := fmt.Sprintf("/v3/accounts/%d/connections/%d/", accountID, connectionID)

	// like the real API, the responses never contain the private key
	connectionResponse := func() map[string]interface{} {
		// the API returns the nullable fields as explicit nulls
		config := map[string]interface{}{
			"project_id":                  "my-gcp-project-id",
			"timeout_seconds":             300,
			"private_key_id":              "my-private-key-id",
			"client_email":                "my_client_email",
			"client_id":                   "my_client_id",
			"auth_uri":                    "my_auth_uri",
			"token_uri":                   "my_token_uri",
			"auth_provider_x509_cert_url": "my_auth_provider_x509_cert_url",
			"client_x509_cert_url":        "my_client_x509_cert_url",
			"retries":                     1,
			"scopes": []string{
				"https://www.googleapis.com/auth/bigquery",
				"https://www.googleapis.com/auth/cloud-platform",
				"https://www.googleapis.com/auth/drive",
			},
		}
		for _, nullable := range []string{
			"job_execution_timeout_seconds", "priority", "location", "maximum_bytes_billed",
			"execution_project", "impersonate_service_account", "job_retry_deadline_seconds",
			"job_creation_timeout_seconds", "application_id", "application_secret", "gcs_bucket",
			"dataproc_region", "dataproc_cluster_name", "api_endpoint", "deployment_env_auth_type",
		} {
			config[nullable] = nil
		}

		return map[string]interface{}{
			"status": map[string]interface{}{"code": 200, "is_success": true},
			"data": map[string]interface{}{
				"id":                       connectionID,
				"account_id":               accountID,
				"name":                     "test-bigquery",
				"adapter_version":          "bigquery_v0",
				"is_ssh_tunnel_enabled":    false,
				"config":                   config,
				"private_link_endpoint_id": nil,
				"oauth_configuration_id":   nil,
			},
		}
	}

	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"POST " + collectionPath: func(r *http.Request) (int, interface{}, error) {
			return http.StatusCreated, connectionResponse(), nil
		},
		"GET " + itemPath: func(r *http.Request) (int, interface{}, error) {
			return http.StatusOK, connectionResponse(), nil
		},
		"PATCH " + itemPath: func(r *http.Request) (int, interface{}, error) {
			return http.StatusOK, connectionResponse(), nil
		},
		"DELETE " + itemPath: func(r *http.Request) (int, interface{}, error) {
			return http.StatusOK, map[string]interface{}{
				"status": map[string]interface{}{"code": 200, "is_success": true},
			}, nil
		},
	})
	defer srv.Close()

	privateKeyOf := func(method, path string, index int) string {
		t.Helper()
		var calls []map[string]interface{}
		for _, call := range srv.GetCapturedCalls(path) {
			if call.Method == method {
				calls = append(calls, call.BodyAsMap(t))
			}
		}
		if len(calls) <= index {
			t.Fatalf("expected at least %d %s call(s) to %s, got %d", index+1, method, path, len(calls))
		}
		config, ok := calls[index]["config"].(map[string]interface{})
		if !ok {
			t.Fatalf("no config in the payload: %v", calls[index])
		}
		privateKey, _ := config["private_key"].(string)
		return privateKey
	}

	config := func(privateKeyWo string, version int) string {
		return bigqueryPlanOnlyConfig(
			srv.URL,
			fmt.Sprintf("private_key_wo = %q\n    private_key_wo_version = %d", privateKeyWo, version),
		)
	}

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		Steps: []resource.TestStep{
			{
				Config: config("first-key", 1),
				Check: resource.ComposeTestCheckFunc(
					func(*terraform.State) error {
						if got := privateKeyOf("POST", collectionPath, 0); got != "first-key" {
							return fmt.Errorf("expected private_key first-key on create, got %q", got)
						}
						return nil
					},
					resource.TestCheckNoResourceAttr("dbtcloud_global_connection.test", "bigquery.private_key_wo"),
					resource.TestCheckNoResourceAttr("dbtcloud_global_connection.test", "bigquery.private_key"),
					resource.TestCheckResourceAttr("dbtcloud_global_connection.test", "bigquery.private_key_wo_version", "1"),
				),
			},
			{
				// same version, new value: nothing to update, the version is what triggers it
				Config:   config("ignored-key", 1),
				PlanOnly: true,
			},
			{
				Config: config("second-key", 2),
				Check: resource.ComposeTestCheckFunc(
					func(*terraform.State) error {
						if got := privateKeyOf("PATCH", itemPath, 0); got != "second-key" {
							return fmt.Errorf("expected private_key second-key on update, got %q", got)
						}
						return nil
					},
					resource.TestCheckNoResourceAttr("dbtcloud_global_connection.test", "bigquery.private_key_wo"),
					resource.TestCheckResourceAttr("dbtcloud_global_connection.test", "bigquery.private_key_wo_version", "2"),
				),
			},
		},
	})
}
