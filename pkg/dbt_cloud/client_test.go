package dbt_cloud

import (
	"strings"
	"testing"
)

// TestParseAPIError tests the parseAPIError helper function for various API error responses
func TestParseAPIError(t *testing.T) {
	tests := []struct {
		name          string
		body          []byte
		expectedIs404 bool
		expectedCode  int
		expectedMsg   string
		expectedError bool
	}{
		{
			name: "404 with permission hint",
			body: []byte(`{
				"status": {
					"code": 404,
					"is_success": false,
					"user_message": "The requested resource was not found. Please check that you have the proper permissions.",
					"developer_message": ""
				},
				"data": null
			}`),
			expectedIs404: true,
			expectedCode:  404,
			expectedMsg:   "The requested resource was not found. Please check that you have the proper permissions.",
			expectedError: false,
		},
		{
			name: "404 without permission hint",
			body: []byte(`{
				"status": {
					"code": 404,
					"is_success": false,
					"user_message": "The requested resource was not found.",
					"developer_message": ""
				},
				"data": null
			}`),
			expectedIs404: true,
			expectedCode:  404,
			expectedMsg:   "The requested resource was not found.",
			expectedError: false,
		},
		{
			name: "403 forbidden",
			body: []byte(`{
				"status": {
					"code": 403,
					"is_success": false,
					"user_message": "Forbidden: You do not have permission to access this resource.",
					"developer_message": ""
				},
				"data": null
			}`),
			expectedIs404: false,
			expectedCode:  403,
			expectedMsg:   "Forbidden: You do not have permission to access this resource.",
			expectedError: false,
		},
		{
			name: "401 unauthorized",
			body: []byte(`{
				"status": {
					"code": 401,
					"is_success": false,
					"user_message": "Unauthorized: Invalid or expired token.",
					"developer_message": ""
				},
				"data": null
			}`),
			expectedIs404: false,
			expectedCode:  401,
			expectedMsg:   "Unauthorized: Invalid or expired token.",
			expectedError: false,
		},
		{
			name:          "invalid JSON",
			body:          []byte(`not valid json {}`),
			expectedIs404: false,
			expectedError: true,
		},
		{
			name:          "empty response",
			body:          []byte(``),
			expectedIs404: false,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is404, apiErr, err := parseAPIError(tt.body)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if is404 != tt.expectedIs404 {
				t.Errorf("expected is404=%v, got %v", tt.expectedIs404, is404)
			}

			if apiErr == nil {
				t.Errorf("expected apiErr to not be nil")
				return
			}

			if apiErr.Status.Code != tt.expectedCode {
				t.Errorf("expected status code=%d, got %d", tt.expectedCode, apiErr.Status.Code)
			}

			if apiErr.Status.UserMessage != tt.expectedMsg {
				t.Errorf("expected message=%q, got %q", tt.expectedMsg, apiErr.Status.UserMessage)
			}
		})
	}
}

// TestDetectPermissionIn404 tests detection of permission-related error messages
func TestDetectPermissionIn404(t *testing.T) {
	tests := []struct {
		name                string
		userMessage         string
		shouldBePermissions bool
	}{
		{
			name:                "explicit proper permissions mention",
			userMessage:         "The requested resource was not found. Please check that you have the proper permissions.",
			shouldBePermissions: true,
		},
		{
			name:                "lowercase proper permissions",
			userMessage:         "resource not found. please check that you have the proper permissions.",
			shouldBePermissions: true,
		},
		{
			name:                "generic permission keyword",
			userMessage:         "Resource not found due to insufficient permissions.",
			shouldBePermissions: true,
		},
		{
			name:                "permission in different context",
			userMessage:         "You do not have permission to access this resource.",
			shouldBePermissions: true,
		},
		{
			name:                "generic not found - no permission hint",
			userMessage:         "The requested resource was not found.",
			shouldBePermissions: false,
		},
		{
			name:                "deleted resource",
			userMessage:         "This resource has been deleted.",
			shouldBePermissions: false,
		},
		{
			name:                "invalid ID",
			userMessage:         "Invalid resource ID provided.",
			shouldBePermissions: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the logic used in doRequestWithRetry
			lowerMsg := strings.ToLower(tt.userMessage)
			hasPermissionHint := strings.Contains(lowerMsg, "permission") || strings.Contains(lowerMsg, "proper permissions")

			if hasPermissionHint != tt.shouldBePermissions {
				t.Errorf("expected containsPermissionHint=%v, got %v for message: %q",
					tt.shouldBePermissions, hasPermissionHint, tt.userMessage)
			}
		})
	}
}

// TestIsResourceNotFoundError tests the legacy helper function
func TestIsResourceNotFoundError(t *testing.T) {
	tests := []struct {
		name          string
		body          []byte
		expectedFound bool
		expectedError bool
	}{
		{
			name: "valid 404 response",
			body: []byte(`{
				"status": {
					"code": 404,
					"is_success": false,
					"user_message": "Resource not found",
					"developer_message": ""
				},
				"data": null
			}`),
			expectedFound: true,
			expectedError: false,
		},
		{
			name: "non-404 response",
			body: []byte(`{
				"status": {
					"code": 200,
					"is_success": true,
					"user_message": "Success",
					"developer_message": ""
				},
				"data": {}
			}`),
			expectedFound: false,
			expectedError: false,
		},
		{
			name:          "invalid JSON",
			body:          []byte(`{invalid json`),
			expectedFound: false,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := isResourceNotFoundError(tt.body)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if found != tt.expectedFound {
				t.Errorf("expected found=%v, got %v", tt.expectedFound, found)
			}
		})
	}
}

// TestNewClientTimeout tests that the timeout configuration is properly set on the client
func TestNewClientTimeout(t *testing.T) {
	// Set environment variable for acceptance test to skip validation
	t.Setenv("TF_ACC", "1")

	tests := []struct {
		name                   string
		timeoutSeconds         int
		expectedTimeoutSeconds int
	}{
		{
			name:                   "default timeout of 30 seconds",
			timeoutSeconds:         30,
			expectedTimeoutSeconds: 30,
		},
		{
			name:                   "custom timeout of 60 seconds",
			timeoutSeconds:         60,
			expectedTimeoutSeconds: 60,
		},
		{
			name:                   "custom timeout of 10 seconds",
			timeoutSeconds:         10,
			expectedTimeoutSeconds: 10,
		},
		{
			name:                   "custom timeout of 120 seconds",
			timeoutSeconds:         120,
			expectedTimeoutSeconds: 120,
		},
		{
			name:                   "short timeout of 5 seconds",
			timeoutSeconds:         5,
			expectedTimeoutSeconds: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test parameters
			accountID := 123456
			token := "test_token_12345"
			hostURL := "https://cloud.getdbt.com/api"
			maxRetries := 3
			retryIntervalSeconds := 10
			retriableStatusCodes := []string{"429", "500"}
			skipValidation := true

			// Create client
			client, err := NewClient(
				&accountID,
				&token,
				&hostURL,
				&maxRetries,
				&retryIntervalSeconds,
				retriableStatusCodes,
				skipValidation,
				&tt.timeoutSeconds,
			)

			if err != nil {
				t.Fatalf("unexpected error creating client: %v", err)
			}

			// Verify timeout is set correctly in the client struct
			if client.TimeoutSeconds != tt.expectedTimeoutSeconds {
				t.Errorf("expected TimeoutSeconds=%d, got %d",
					tt.expectedTimeoutSeconds, client.TimeoutSeconds)
			}

			// Verify HTTP client timeout is set correctly
			expectedTimeout := tt.expectedTimeoutSeconds
			actualTimeoutSeconds := int(client.HTTPClient.Timeout.Seconds())
			if actualTimeoutSeconds != expectedTimeout {
				t.Errorf("expected HTTPClient.Timeout=%ds, got %ds",
					expectedTimeout, actualTimeoutSeconds)
			}
		})
	}
}

// TestNewClientValidation tests basic validation in NewClient
func TestNewClientValidation(t *testing.T) {
	tests := []struct {
		name          string
		token         *string
		hostURL       *string
		expectedError string
	}{
		{
			name:          "empty token",
			token:         stringPtr(""),
			hostURL:       stringPtr("https://cloud.getdbt.com/api"),
			expectedError: "token is set but it is empty",
		},
		{
			name:          "nil token",
			token:         nil,
			hostURL:       stringPtr("https://cloud.getdbt.com/api"),
			expectedError: "token is set but it is empty",
		},
		{
			name:          "invalid host URL",
			token:         stringPtr("test_token"),
			hostURL:       stringPtr("://invalid-url"),
			expectedError: "invalid host URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountID := 123456
			maxRetries := 3
			retryIntervalSeconds := 10
			timeoutSeconds := 30
			retriableStatusCodes := []string{"429"}
			skipValidation := true

			_, err := NewClient(
				&accountID,
				tt.token,
				tt.hostURL,
				&maxRetries,
				&retryIntervalSeconds,
				retriableStatusCodes,
				skipValidation,
				&timeoutSeconds,
			)

			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.expectedError)
				return
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("expected error containing %q, got %q", tt.expectedError, err.Error())
			}
		})
	}
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}
