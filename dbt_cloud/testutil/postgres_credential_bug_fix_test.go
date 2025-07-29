package testutil

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
)

func TestPostgresCredential_UpdateBugRegression(t *testing.T) {
	originalTFAcc := os.Getenv("TF_ACC")
	os.Setenv("TF_ACC", "1")
	defer func() {
		if originalTFAcc == "" {
			os.Unsetenv("TF_ACC")
		} else {
			os.Setenv("TF_ACC", originalTFAcc)
		}
	}()

	accountID, projectID, credentialID := 12345, 67890, 111
	tracker := &APICallTracker{}
	
	config := ResourceTestConfig{
		ResourceType: "dbtcloud_postgres_credential",
		AccountID:    accountID,
		ProjectID:    projectID,
		ResourceID:   credentialID,
		APIPath:      "credentials",
	}
	
	handlers := CreateResourceTestHandlers(t, config, tracker)
	updatePostgresCredentialHandlers(handlers, accountID, projectID, credentialID, tracker)

	srv := SetupMockServer(t, handlers)
	defer srv.Close()

	providerConfig := fmt.Sprintf(`
		provider "dbtcloud" {
			host_url   = "%s"
			token      = "dummy-token"
			account_id = %d
		}`, srv.URL, accountID)

	initialConfig := providerConfig + fmt.Sprintf(`
		resource "dbtcloud_postgres_credential" "test" {
			project_id     = %d
			type           = "postgres"
			default_schema = "test_schema" 
			username       = "test_user"
			password       = "test_password"
			num_threads    = 4
			target_name    = "default"
		}`, projectID)

	updatedConfig := providerConfig + fmt.Sprintf(`
		resource "dbtcloud_postgres_credential" "test" {
			project_id     = %d
			type           = "postgres"
			default_schema = "updated_schema"
			username       = "test_user"
			password       = "test_password"
			num_threads    = 4
			target_name    = "default"
		}`, projectID)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: initialConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dbtcloud_postgres_credential.test", "default_schema", "test_schema"),
					resource.TestCheckResourceAttr("dbtcloud_postgres_credential.test", "credential_id", fmt.Sprintf("%d", credentialID)),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dbtcloud_postgres_credential.test", "default_schema", "updated_schema"),
					resource.TestCheckResourceAttr("dbtcloud_postgres_credential.test", "credential_id", fmt.Sprintf("%d", credentialID)),
					verifyBugIsFixed(t, tracker),
				),
			},
		},
	})
}

func verifyBugIsFixed(t *testing.T, tracker *APICallTracker) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		assert.Equal(t, 1, tracker.CreateCount, "expected exactly 1 CREATE call")
		assert.Equal(t, 1, tracker.UpdateCount, "expected exactly 1 UPDATE call")
		assert.GreaterOrEqual(t, tracker.ReadCount, 1, "expected at least 1 READ call")
		return nil
	}
}

func updatePostgresCredentialHandlers(handlers map[string]MockEndpointHandler, accountID, projectID, credentialID int, tracker *APICallTracker) {
	currentSchema := "test_schema"
	
	createResponse := func() dbt_cloud.PostgresCredentialResponse {
		return dbt_cloud.PostgresCredentialResponse{
			Data: dbt_cloud.PostgresCredential{
				ID:             &credentialID,
				Account_Id:     accountID,
				Project_Id:     projectID,
				Type:          "postgres",
				State:         1, // Active state
				Threads:       4,
				Username:      "test_user",
				Default_Schema: currentSchema,
				Target_Name:    "default",
			},
			Status: dbt_cloud.ResponseStatus{
				Code:        200,
				Is_Success:  true,
				User_Message: "",
			},
		}
	}
	
	createPath := fmt.Sprintf("POST /v3/accounts/%d/projects/%d/credentials/", accountID, projectID)
	handlers[createPath] = func(r *http.Request) (int, interface{}, error) {
		tracker.CreateCount++
		response := createResponse()
		response.Status.Code = 201
		return http.StatusCreated, response, nil
	}
	
	readPath := fmt.Sprintf("GET /v3/accounts/%d/projects/%d/credentials/%d/", accountID, projectID, credentialID)
	handlers[readPath] = func(r *http.Request) (int, interface{}, error) {
		tracker.ReadCount++
		return http.StatusOK, createResponse(), nil
	}
	
	updatePath := fmt.Sprintf("POST /v3/accounts/%d/projects/%d/credentials/%d/", accountID, projectID, credentialID)
	handlers[updatePath] = func(r *http.Request) (int, interface{}, error) {
		tracker.UpdateCount++
		currentSchema = "updated_schema"
		return http.StatusOK, createResponse(), nil
	}
} 