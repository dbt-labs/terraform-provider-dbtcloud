package testhelpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// CapturedCall holds the details of a single HTTP request received by the mock server
type CapturedCall struct {
	Method  string
	Path    string
	Headers http.Header
	Body    []byte
}

// BodyAsMap unmarshals the JSON body of the call into a map for easy validation
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

// MockEndpointHandler is a function that inspects a request and dynamically determines the response
type MockEndpointHandler func(r *http.Request) (statusCode int, responseBody interface{}, err error)

// MockServer is a wrapper around httptest.Server that captures calls
type MockServer struct {
	*httptest.Server
	mu            sync.Mutex
	capturedCalls map[string][]*CapturedCall
}

// SetupMockServer creates and starts a new MockServer using dynamic handlers
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

// GetCapturedCalls returns all captured calls for a given path
func (m *MockServer) GetCapturedCalls(path string) []*CapturedCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.capturedCalls[path]
}

// ResourceTestConfig holds configuration for testing a dbt Cloud resource
type ResourceTestConfig struct {
	ResourceType string
	AccountID    int64
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
