package testutil

import (
	"fmt"
	"net/http"
	"testing"
)

// ResourceTestConfig holds configuration for testing a dbt Cloud resource
type ResourceTestConfig struct {
	ResourceType string // e.g., "dbtcloud_postgres_credential"
	AccountID    int
	ProjectID    int    // Optional, only needed for project-scoped resources
	ResourceID   int    // The main resource ID
	APIPath      string // e.g., "credentials" for /v3/accounts/{account}/projects/{project}/credentials/
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

	// CREATE handler
	var createPath string
	if config.ProjectID > 0 {
		createPath = fmt.Sprintf("POST /v3/accounts/%d/projects/%d/%s/", config.AccountID, config.ProjectID, config.APIPath)
	} else {
		createPath = fmt.Sprintf("POST /v3/accounts/%d/%s/", config.AccountID, config.APIPath)
	}
	
	handlers[createPath] = func(r *http.Request) (int, interface{}, error) {
		tracker.CreateCount++
		response := copyMap(baseResponse)
		response["status"].(map[string]interface{})["code"] = 201
		return http.StatusCreated, response, nil
	}
	
	// READ handler
	var readPath string
	if config.ProjectID > 0 {
		readPath = fmt.Sprintf("GET /v3/accounts/%d/projects/%d/%s/%d/", config.AccountID, config.ProjectID, config.APIPath, config.ResourceID)
	} else {
		readPath = fmt.Sprintf("GET /v3/accounts/%d/%s/%d/", config.AccountID, config.APIPath, config.ResourceID)
	}
	
	handlers[readPath] = func(r *http.Request) (int, interface{}, error) {
		tracker.ReadCount++
		return http.StatusOK, copyMap(baseResponse), nil
	}
	
	// UPDATE handler
	var updatePath string
	if config.ProjectID > 0 {
		updatePath = fmt.Sprintf("POST /v3/accounts/%d/projects/%d/%s/%d/", config.AccountID, config.ProjectID, config.APIPath, config.ResourceID)
	} else {
		updatePath = fmt.Sprintf("POST /v3/accounts/%d/%s/%d/", config.AccountID, config.APIPath, config.ResourceID)
	}
	
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