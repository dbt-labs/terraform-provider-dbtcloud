package postgres_credential_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
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
		ImportStateVerifyIgnore: []string{"password", "semantic_layer_credential"},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudPostgresCredentialDestroy,
		Steps: []resource.TestStep{
			createCredentialTestStep,
			importStateTestStep,
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

// Mock server utilities for bug fix regression tests

// CapturedCall holds the details of a single HTTP request received by the mock server.
type CapturedCall struct {
	Method  string
	Path    string
	Headers http.Header
	Body    []byte
}

// BodyAsMap unmarshals the JSON body of the call into a map for easy validation.
func (c *CapturedCall) BodyAsMap(t *testing.T) map[string]interface{} {
	t.Helper()
	if len(c.Body) == 0 {
		return nil
	}
	var data map[string]interface{}
	if err := json.Unmarshal(c.Body, &data); err != nil {
		t.Fatalf("Failed to unmarshal call body: %v. Body was: %s", err, string(c.Body))
	}
	return data
}

// MockEndpointHandler is a function that inspects a request and dynamically determines the response.
type MockEndpointHandler func(r *http.Request) (statusCode int, responseBody interface{}, err error)

// MockServer is a wrapper around httptest.Server that captures calls.
type MockServer struct {
	*httptest.Server
	mu            sync.Mutex
	capturedCalls map[string][]*CapturedCall
}

// SetupMockServer creates and starts a new MockServer using dynamic handlers.
func SetupMockServer(t *testing.T, handlers map[string]MockEndpointHandler) *MockServer {
	t.Helper()

	mockServer := &MockServer{
		capturedCalls: make(map[string][]*CapturedCall),
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			http.Error(w, "cannot read body", http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		call := &CapturedCall{
			Method:  r.Method,
			Path:    r.URL.Path,
			Headers: r.Header,
			Body:    body,
		}

		mockServer.mu.Lock()
		mockServer.capturedCalls[r.URL.Path] = append(mockServer.capturedCalls[r.URL.Path], call)
		mockServer.mu.Unlock()

		key := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		if handlerFunc, ok := handlers[key]; ok {
			statusCode, responseBody, err := handlerFunc(r)

			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			if responseBody != nil {
				if err := json.NewEncoder(w).Encode(responseBody); err != nil {
					log.Printf("Error encoding JSON response: %v", err)
				}
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)

		errorResponse := map[string]interface{}{
			"error": fmt.Sprintf("Mock server received unexpected call: %s %s. No handler configured.", r.Method, r.URL.Path),
			"status": map[string]interface{}{
				"code":              404,
				"is_success":        false,
				"user_message":      "Resource not found",
				"developer_message": fmt.Sprintf("No mock handler configured for %s %s", r.Method, r.URL.Path),
			},
		}

		if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
			t.Logf("Error encoding JSON error response: %v", err)
		}
	})

	mockServer.Server = httptest.NewServer(handler)
	return mockServer
}

// ResourceTestConfig holds configuration for testing a dbt Cloud resource
type ResourceTestConfig struct {
	ResourceType string
	AccountID    int
	ProjectID    int
	ResourceID   int
	APIPath      string
}

// APICallTracker tracks API calls made during testing
type APICallTracker struct {
	CreateCount int
	ReadCount   int
	UpdateCount int
	DeleteCount int
}

// CreateResourceTestHandlers creates standard CRUD handlers for testing any dbt Cloud resource
func CreateResourceTestHandlers(t *testing.T, config ResourceTestConfig, tracker *APICallTracker) map[string]MockEndpointHandler {
	handlers := make(map[string]MockEndpointHandler)

	baseResponse := map[string]interface{}{
		"data": map[string]interface{}{
			"id":         config.ResourceID,
			"account_id": config.AccountID,
		},
		"status": map[string]interface{}{
			"code":       200,
			"is_success": true,
		},
	}

	if config.ProjectID > 0 {
		baseResponse["data"].(map[string]interface{})["project_id"] = config.ProjectID
	}

	createPath := fmt.Sprintf("POST /v3/accounts/%d/projects/%d/%s/", config.AccountID, config.ProjectID, config.APIPath)
	handlers[createPath] = func(r *http.Request) (int, interface{}, error) {
		tracker.CreateCount++
		response := copyMap(baseResponse)
		response["status"].(map[string]interface{})["code"] = 201
		return http.StatusCreated, response, nil
	}

	readPath := fmt.Sprintf("GET /v3/accounts/%d/projects/%d/%s/%d/", config.AccountID, config.ProjectID, config.APIPath, config.ResourceID)
	handlers[readPath] = func(r *http.Request) (int, interface{}, error) {
		tracker.ReadCount++
		return http.StatusOK, copyMap(baseResponse), nil
	}

	updatePath := fmt.Sprintf("POST /v3/accounts/%d/projects/%d/%s/%d/", config.AccountID, config.ProjectID, config.APIPath, config.ResourceID)
	handlers[updatePath] = func(r *http.Request) (int, interface{}, error) {
		tracker.UpdateCount++
		return http.StatusOK, copyMap(baseResponse), nil
	}

	return handlers
}

// copyMap creates a deep copy of a map for response reuse
func copyMap(original map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	for k, v := range original {
		if nestedMap, ok := v.(map[string]interface{}); ok {
			copy[k] = copyMap(nestedMap)
		} else {
			copy[k] = v
		}
	}
	return copy
}

// TestPostgresCredential_UpdateBugRegression tests the bug fix for credential updates
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
				Type:           "postgres",
				State:          1,
				Threads:        4,
				Username:       "test_user",
				Default_Schema: currentSchema,
				Target_Name:    "default",
			},
			Status: dbt_cloud.ResponseStatus{
				Code:         200,
				Is_Success:   true,
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
