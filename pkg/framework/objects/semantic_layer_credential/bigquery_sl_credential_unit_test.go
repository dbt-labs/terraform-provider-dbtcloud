package semantic_layer_credential_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/dbt_cloud/testutil"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/semantic_layer_credential"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// helpers

func intPtr(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
}

func newTestClient(t *testing.T, serverURL string) *dbt_cloud.Client {
	t.Helper()
	accountID := int64(1)
	maxRetries := 0
	retryIntervalSeconds := 1
	timeoutSeconds := 30
	hostURL := serverURL + "/api"
	token := "test-token"

	client, err := dbt_cloud.NewClient(
		&accountID,
		strPtr(token),
		strPtr(hostURL),
		intPtr(maxRetries),
		intPtr(retryIntervalSeconds),
		[]string{},
		true,
		intPtr(timeoutSeconds),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	return client
}

func mockCredentialResponse(id int, values map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"id":              id,
			"account_id":      1,
			"project_id":      100,
			"name":            "test-credential",
			"adapter_version": "dbt-bigquery>=1.0.0,<2.0.0",
			"schema_type":     "semantic_layer_credentials",
			"values":          values,
		},
		"status": map[string]interface{}{
			"code":       200,
			"is_success": true,
		},
	}
}

// TestBigQuerySLCredentialResource_Schema verifies the schema contains
// execution_project as an optional attribute.
func TestBigQuerySLCredentialResource_Schema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaRequest := resource.SchemaRequest{}
	schemaResponse := &resource.SchemaResponse{}

	r := semantic_layer_credential.BigQuerySemanticLayerCredentialResource()
	r.Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema method returned errors: %v", schemaResponse.Diagnostics.Errors())
	}

	schema := schemaResponse.Schema
	attr, ok := schema.Attributes["execution_project"]
	if !ok {
		t.Fatal("Schema missing execution_project attribute")
	}
	if attr.IsRequired() {
		t.Error("execution_project should be optional, not required")
	}
	if !attr.IsOptional() {
		t.Error("execution_project should be optional")
	}
}

// TestBigQuerySLCredentialResource_Metadata verifies the resource type name.
func TestBigQuerySLCredentialResource_Metadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	metadataRequest := resource.MetadataRequest{ProviderTypeName: "dbtcloud"}
	metadataResponse := &resource.MetadataResponse{}

	r := semantic_layer_credential.BigQuerySemanticLayerCredentialResource()
	r.Metadata(ctx, metadataRequest, metadataResponse)

	expectedTypeName := "dbtcloud_bigquery_semantic_layer_credential"
	if metadataResponse.TypeName != expectedTypeName {
		t.Errorf("Expected TypeName %s, got %s", expectedTypeName, metadataResponse.TypeName)
	}
}

// TestCreateBigQuerySLCredentialWithExecutionProject verifies that execution_project
// is included in the CREATE request body when set.
func TestCreateBigQuerySLCredentialWithExecutionProject(t *testing.T) {
	credentialID := 42
	createPath := "/api/v3/accounts/1/semantic-layer-credentials/"

	var capturedBody map[string]interface{}

	handlers := map[string]testutil.MockEndpointHandler{
		"POST " + createPath: func(r *http.Request) (int, interface{}, error) {
			if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
				return 400, nil, err
			}
			return 200, mockCredentialResponse(credentialID, map[string]interface{}{
				"execution_project": "my-execution-project",
			}), nil
		},
	}

	server := testutil.SetupMockServer(t, handlers)
	defer server.Close()

	client := newTestClient(t, server.URL)

	values := map[string]interface{}{
		"private_key_id":              "key-id",
		"private_key":                 "private-key",
		"client_email":                "svc@project.iam.gserviceaccount.com",
		"client_id":                   "client-id",
		"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
		"token_uri":                   "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url":        "https://www.googleapis.com/robot/v1/metadata/x509/svc",
		"execution_project":           "my-execution-project",
	}

	_, err := client.CreateSemanticLayerCredential(100, values, "test-credential", "dbt-bigquery>=1.0.0,<2.0.0")
	if err != nil {
		t.Fatalf("CreateSemanticLayerCredential failed: %v", err)
	}

	capturedValues, ok := capturedBody["values"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected 'values' map in request body, got: %v", capturedBody["values"])
	}
	if execProj, ok := capturedValues["execution_project"]; !ok {
		t.Error("execution_project was not sent in the create request body")
	} else if execProj != "my-execution-project" {
		t.Errorf("execution_project: expected %q, got %q", "my-execution-project", execProj)
	}
}

// TestCreateBigQuerySLCredentialWithoutExecutionProject verifies that execution_project
// is NOT included in the CREATE request body when absent.
func TestCreateBigQuerySLCredentialWithoutExecutionProject(t *testing.T) {
	credentialID := 43
	createPath := "/api/v3/accounts/1/semantic-layer-credentials/"

	var capturedBody map[string]interface{}

	handlers := map[string]testutil.MockEndpointHandler{
		"POST " + createPath: func(r *http.Request) (int, interface{}, error) {
			if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
				return 400, nil, err
			}
			return 200, mockCredentialResponse(credentialID, map[string]interface{}{}), nil
		},
	}

	server := testutil.SetupMockServer(t, handlers)
	defer server.Close()

	client := newTestClient(t, server.URL)

	values := map[string]interface{}{
		"private_key_id":              "key-id",
		"private_key":                 "private-key",
		"client_email":                "svc@project.iam.gserviceaccount.com",
		"client_id":                   "client-id",
		"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
		"token_uri":                   "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url":        "https://www.googleapis.com/robot/v1/metadata/x509/svc",
		// no execution_project
	}

	_, err := client.CreateSemanticLayerCredential(100, values, "test-credential", "dbt-bigquery>=1.0.0,<2.0.0")
	if err != nil {
		t.Fatalf("CreateSemanticLayerCredential failed: %v", err)
	}

	if capturedValues, ok := capturedBody["values"].(map[string]interface{}); ok {
		if _, exists := capturedValues["execution_project"]; exists {
			t.Error("execution_project should not be present in the create request body when not set")
		}
	}
}

// TestReadBigQuerySLCredentialWithExecutionProject verifies that execution_project
// is populated from the API response when present.
func TestReadBigQuerySLCredentialWithExecutionProject(t *testing.T) {
	credentialID := 44
	getPath := fmt.Sprintf("/api/v3/accounts/1/semantic-layer-credentials/%d", credentialID)

	handlers := map[string]testutil.MockEndpointHandler{
		"GET " + getPath: func(r *http.Request) (int, interface{}, error) {
			return 200, mockCredentialResponse(credentialID, map[string]interface{}{
				"execution_project": "my-execution-project",
			}), nil
		},
	}

	server := testutil.SetupMockServer(t, handlers)
	defer server.Close()

	credential, err := newTestClient(t, server.URL).GetSemanticLayerCredential(int64(credentialID))
	if err != nil {
		t.Fatalf("GetSemanticLayerCredential failed: %v", err)
	}

	if execProj, ok := credential.Values["execution_project"]; !ok {
		t.Error("execution_project was not returned in the credential values")
	} else if execProj != "my-execution-project" {
		t.Errorf("execution_project: expected %q, got %q", "my-execution-project", execProj)
	}
}

// TestReadBigQuerySLCredentialWithoutExecutionProject verifies that execution_project
// is absent from the values map when not returned by the API.
func TestReadBigQuerySLCredentialWithoutExecutionProject(t *testing.T) {
	credentialID := 45
	getPath := fmt.Sprintf("/api/v3/accounts/1/semantic-layer-credentials/%d", credentialID)

	handlers := map[string]testutil.MockEndpointHandler{
		"GET " + getPath: func(r *http.Request) (int, interface{}, error) {
			return 200, mockCredentialResponse(credentialID, map[string]interface{}{
				"client_email": "svc@project.iam.gserviceaccount.com",
			}), nil
		},
	}

	server := testutil.SetupMockServer(t, handlers)
	defer server.Close()

	credential, err := newTestClient(t, server.URL).GetSemanticLayerCredential(int64(credentialID))
	if err != nil {
		t.Fatalf("GetSemanticLayerCredential failed: %v", err)
	}

	if _, exists := credential.Values["execution_project"]; exists {
		t.Error("execution_project should not be present in values when not returned by the API")
	}
}

// TestUpdateBigQuerySLCredentialWithExecutionProject verifies that execution_project
// is included in the UPDATE request body when set.
func TestUpdateBigQuerySLCredentialWithExecutionProject(t *testing.T) {
	credentialID := 46
	updatePath := fmt.Sprintf("/api/v3/accounts/1/semantic-layer-credentials/%d/", credentialID)

	var capturedBody map[string]interface{}

	handlers := map[string]testutil.MockEndpointHandler{
		"POST " + updatePath: func(r *http.Request) (int, interface{}, error) {
			if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
				return 400, nil, err
			}
			return 200, mockCredentialResponse(credentialID, map[string]interface{}{
				"execution_project": "updated-execution-project",
			}), nil
		},
	}

	server := testutil.SetupMockServer(t, handlers)
	defer server.Close()

	credential := dbt_cloud.SemanticLayerCredentials{
		ID:             intPtr(credentialID),
		Name:           "test-credential",
		ProjectID:      100,
		AccountID:      1,
		AdapterVersion: "dbt-bigquery>=1.0.0,<2.0.0",
		SchemaType:     "semantic_layer_credentials",
		Values: map[string]interface{}{
			"private_key_id":              "key-id",
			"private_key":                 "private-key",
			"client_email":                "svc@project.iam.gserviceaccount.com",
			"client_id":                   "client-id",
			"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
			"token_uri":                   "https://oauth2.googleapis.com/token",
			"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
			"client_x509_cert_url":        "https://www.googleapis.com/robot/v1/metadata/x509/svc",
			"execution_project":           "updated-execution-project",
		},
	}

	_, err := newTestClient(t, server.URL).UpdateSemanticLayerCredential(int64(credentialID), credential)
	if err != nil {
		t.Fatalf("UpdateSemanticLayerCredential failed: %v", err)
	}

	capturedValues, ok := capturedBody["values"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected 'values' map in request body, got: %v", capturedBody["values"])
	}
	if execProj, ok := capturedValues["execution_project"]; !ok {
		t.Error("execution_project was not sent in the update request body")
	} else if execProj != "updated-execution-project" {
		t.Errorf("execution_project: expected %q, got %q", "updated-execution-project", execProj)
	}
}

// TestUpdateBigQuerySLCredentialRemoveExecutionProject verifies that execution_project
// is NOT sent in the UPDATE request body when removed.
func TestUpdateBigQuerySLCredentialRemoveExecutionProject(t *testing.T) {
	credentialID := 47
	updatePath := fmt.Sprintf("/api/v3/accounts/1/semantic-layer-credentials/%d/", credentialID)

	var capturedBody map[string]interface{}

	handlers := map[string]testutil.MockEndpointHandler{
		"POST " + updatePath: func(r *http.Request) (int, interface{}, error) {
			if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
				return 400, nil, err
			}
			return 200, mockCredentialResponse(credentialID, map[string]interface{}{}), nil
		},
	}

	server := testutil.SetupMockServer(t, handlers)
	defer server.Close()

	credential := dbt_cloud.SemanticLayerCredentials{
		ID:             intPtr(credentialID),
		Name:           "test-credential",
		ProjectID:      100,
		AccountID:      1,
		AdapterVersion: "dbt-bigquery>=1.0.0,<2.0.0",
		SchemaType:     "semantic_layer_credentials",
		Values: map[string]interface{}{
			"private_key_id":              "key-id",
			"private_key":                 "private-key",
			"client_email":                "svc@project.iam.gserviceaccount.com",
			"client_id":                   "client-id",
			"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
			"token_uri":                   "https://oauth2.googleapis.com/token",
			"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
			"client_x509_cert_url":        "https://www.googleapis.com/robot/v1/metadata/x509/svc",
			// no execution_project
		},
	}

	_, err := newTestClient(t, server.URL).UpdateSemanticLayerCredential(int64(credentialID), credential)
	if err != nil {
		t.Fatalf("UpdateSemanticLayerCredential failed: %v", err)
	}

	if capturedValues, ok := capturedBody["values"].(map[string]interface{}); ok {
		if _, exists := capturedValues["execution_project"]; exists {
			t.Error("execution_project should not be sent in the update body when removed")
		}
	}
}
