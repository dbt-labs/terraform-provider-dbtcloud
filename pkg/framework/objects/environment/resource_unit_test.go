package environment_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/dbt_cloud/testutil"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
)

// Helper to create int pointer
func intPtr(i int) *int {
	return &i
}

// Helper to create string pointer
func strPtr(s string) *string {
	return &s
}

// TestUpdateEnvironmentRemoveExtendedAttributesID validates that when removing
// extended_attributes_id from an environment, the API receives null instead of 0.
// This is a regression test for the foreign key constraint violation:
// "Key (extended_attributes_id)=(0) is not present in table extended_attributes"
func TestUpdateEnvironmentRemoveExtendedAttributesID(t *testing.T) {
	projectID := 54321
	environmentID := 98765

	basePath := fmt.Sprintf("/api/v3/accounts/1/projects/%d/environments/%d/", projectID, environmentID)

	// Track what the update request body contains
	var lastUpdateBody map[string]interface{}

	// Mock environment response (without extended_attributes_id)
	mockEnvResponse := map[string]interface{}{
		"data": map[string]interface{}{
			"id":                         environmentID,
			"account_id":                 1,
			"project_id":                 projectID,
			"name":                       "Florinda",
			"dbt_version":                "versionless",
			"type":                       "deployment",
			"use_custom_branch":          false,
			"state":                      1,
			"extended_attributes_id":     nil,
			"enable_model_query_history": false,
		},
		"status": map[string]interface{}{
			"code":       200,
			"is_success": true,
		},
	}

	handlers := map[string]testutil.MockEndpointHandler{
		"GET " + basePath: func(r *http.Request) (int, interface{}, error) {
			return 200, mockEnvResponse, nil
		},
		"POST " + basePath: func(r *http.Request) (int, interface{}, error) {
			// Capture the request body for validation
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				return 400, nil, err
			}
			lastUpdateBody = body
			return 200, mockEnvResponse, nil
		},
	}

	server := testutil.SetupMockServer(t, handlers)
	defer server.Close()

	// Create the client pointing to our mock server
	hostURL := server.URL + "/api"
	accountID := 1
	token := "test-token"
	maxRetries := 0
	retryIntervalSeconds := 1
	timeoutSeconds := 30

	client, err := dbt_cloud.NewClient(
		intPtr(accountID),
		strPtr(token),
		strPtr(hostURL),
		intPtr(maxRetries),
		intPtr(retryIntervalSeconds),
		[]string{},
		true, // skipCredentialsValidation
		intPtr(timeoutSeconds),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create an environment object with ExtendedAttributesID set to nil (simulating removal)
	envWithNullExtAttrs := dbt_cloud.Environment{
		ID:                   &environmentID,
		Account_Id:           1,
		Project_Id:           projectID,
		Name:                 "Florinda",
		Dbt_Version:          "versionless",
		Type:                 "deployment",
		State:                1,
		ExtendedAttributesID: nil, // This is the key - it should be null, not 0
	}

	// Call UpdateEnvironment
	_, err = client.UpdateEnvironment(projectID, environmentID, envWithNullExtAttrs)
	if err != nil {
		t.Fatalf("UpdateEnvironment failed: %v", err)
	}

	// Validate that the request body does NOT contain extended_attributes_id = 0
	if lastUpdateBody != nil {
		if val, exists := lastUpdateBody["extended_attributes_id"]; exists {
			// If it exists and is 0, that's the bug we're testing for
			if numVal, ok := val.(float64); ok && numVal == 0 {
				t.Errorf("BUG: extended_attributes_id was sent as 0 instead of being omitted or null. Body: %v", lastUpdateBody)
			}
		}
		// If extended_attributes_id is not present or is null, that's correct behavior
	}
}

// TestUpdateEnvironmentWithExtendedAttributesID validates that when setting
// extended_attributes_id, the API receives the correct value.
func TestUpdateEnvironmentWithExtendedAttributesID(t *testing.T) {
	projectID := 54321
	environmentID := 98765
	extendedAttributesID := 12345

	basePath := fmt.Sprintf("/api/v3/accounts/1/projects/%d/environments/%d/", projectID, environmentID)

	// Track what the update request body contains
	var lastUpdateBody map[string]interface{}

	// Mock environment response with extended_attributes_id set
	mockEnvResponse := map[string]interface{}{
		"data": map[string]interface{}{
			"id":                         environmentID,
			"account_id":                 1,
			"project_id":                 projectID,
			"name":                       "Gilberto",
			"dbt_version":                "versionless",
			"type":                       "deployment",
			"use_custom_branch":          false,
			"state":                      1,
			"extended_attributes_id":     extendedAttributesID,
			"enable_model_query_history": false,
		},
		"status": map[string]interface{}{
			"code":       200,
			"is_success": true,
		},
	}

	handlers := map[string]testutil.MockEndpointHandler{
		"GET " + basePath: func(r *http.Request) (int, interface{}, error) {
			return 200, mockEnvResponse, nil
		},
		"POST " + basePath: func(r *http.Request) (int, interface{}, error) {
			// Capture the request body for validation
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				return 400, nil, err
			}
			lastUpdateBody = body
			return 200, mockEnvResponse, nil
		},
	}

	server := testutil.SetupMockServer(t, handlers)
	defer server.Close()

	// Create the client pointing to our mock server
	hostURL := server.URL + "/api"
	accountID := 1
	token := "test-token"
	maxRetries := 0
	retryIntervalSeconds := 1
	timeoutSeconds := 30

	client, err := dbt_cloud.NewClient(
		intPtr(accountID),
		strPtr(token),
		strPtr(hostURL),
		intPtr(maxRetries),
		intPtr(retryIntervalSeconds),
		[]string{},
		true, // skipCredentialsValidation
		intPtr(timeoutSeconds),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create an environment object with ExtendedAttributesID set
	envWithExtAttrs := dbt_cloud.Environment{
		ID:                   &environmentID,
		Account_Id:           1,
		Project_Id:           projectID,
		Name:                 "Gilberto",
		Dbt_Version:          "versionless",
		Type:                 "deployment",
		State:                1,
		ExtendedAttributesID: &extendedAttributesID,
	}

	// Call UpdateEnvironment
	_, err = client.UpdateEnvironment(projectID, environmentID, envWithExtAttrs)
	if err != nil {
		t.Fatalf("UpdateEnvironment failed: %v", err)
	}

	// Validate that the request body contains the correct extended_attributes_id
	if lastUpdateBody != nil {
		val, exists := lastUpdateBody["extended_attributes_id"]
		if !exists {
			t.Errorf("extended_attributes_id was not sent in the update body")
		} else if numVal, ok := val.(float64); !ok || int(numVal) != extendedAttributesID {
			t.Errorf("extended_attributes_id was sent as %v, expected %d", val, extendedAttributesID)
		}
	}
}
