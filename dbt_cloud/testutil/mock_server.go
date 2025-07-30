package testutil

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
// If err is non-nil, the server returns a 400 Bad Request with the error message.
// Otherwise, it returns the specified statusCode and marshals responseBody to JSON.
type MockEndpointHandler func(r *http.Request) (statusCode int, responseBody interface{}, err error)

// MockServer is a wrapper around httptest.Server that captures calls.
type MockServer struct {
	*httptest.Server
	mu            sync.Mutex
	capturedCalls map[string][]*CapturedCall // Maps path to a list of calls
}

// SetupMockServer creates and starts a new MockServer using dynamic handlers.
// The handlers map keys should be in the format "METHOD /path", e.g., "POST /users".
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
		// Restore the body for the handler function to read.
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		call := &CapturedCall{
			Method: r.Method,
			Path:   r.URL.Path,
			Headers: r.Header,
			Body:   body,
		}

		mockServer.mu.Lock()
		mockServer.capturedCalls[r.URL.Path] = append(mockServer.capturedCalls[r.URL.Path], call)
		mockServer.mu.Unlock()

		// Find and execute the appropriate handler.
		key := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		if handlerFunc, ok := handlers[key]; ok {
			statusCode, responseBody, err := handlerFunc(r)

			// If the handler returns an error, send a 400 Bad Request.
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// Otherwise, send the specified success response.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			if responseBody != nil {
				if err := json.NewEncoder(w).Encode(responseBody); err != nil {
					log.Printf("Error encoding JSON response: %v", err)
				}
			}
			return
		}

		// If no handler is matched, return an error.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		
		// Return proper JSON error response instead of plain text
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

// GetLastCall returns the most recent call made to a specific path.
func (s *MockServer) GetLastCall(path string) *CapturedCall {
	s.mu.Lock()
	defer s.mu.Unlock()

	calls, ok := s.capturedCalls[path]
	if !ok || len(calls) == 0 {
		return nil
	}
	return calls[len(calls) - 1]
}