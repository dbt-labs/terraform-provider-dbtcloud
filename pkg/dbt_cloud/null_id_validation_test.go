package dbt_cloud

import (
	"strings"
	"testing"
)

// Helper functions for creating pointers
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

// Test type definitions to avoid import cycles
type testResponse struct {
	ID          *int
	Name        string
	TokenString *string
}

// TestValidateRequiredIntPtr tests the internal validation function for int pointers
func TestValidateRequiredIntPtr(t *testing.T) {
	tests := []struct {
		name         string
		ptr          *int
		fieldName    string
		resourceType string
		wantErr      bool
	}{
		{
			name:         "valid int pointer",
			ptr:          intPtr(123),
			fieldName:    "ID",
			resourceType: "TestResource",
			wantErr:      false,
		},
		{
			name:         "nil int pointer",
			ptr:          nil,
			fieldName:    "ID",
			resourceType: "TestResource",
			wantErr:      true,
		},
		{
			name:         "zero value int pointer",
			ptr:          intPtr(0),
			fieldName:    "ID",
			resourceType: "TestResource",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequiredIntPtr(tt.ptr, tt.fieldName, tt.resourceType)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequiredIntPtr() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), tt.fieldName) {
				t.Errorf("error message should contain field name '%s', got: %v", tt.fieldName, err)
			}
		})
	}
}

// TestValidateRequiredStringPtr tests the internal validation function for string pointers
func TestValidateRequiredStringPtr(t *testing.T) {
	tests := []struct {
		name         string
		ptr          *string
		fieldName    string
		resourceType string
		wantErr      bool
	}{
		{
			name:         "valid string pointer",
			ptr:          stringPtr("test"),
			fieldName:    "Name",
			resourceType: "TestResource",
			wantErr:      false,
		},
		{
			name:         "nil string pointer",
			ptr:          nil,
			fieldName:    "Name",
			resourceType: "TestResource",
			wantErr:      true,
		},
		{
			name:         "empty string pointer",
			ptr:          stringPtr(""),
			fieldName:    "Name",
			resourceType: "TestResource",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequiredStringPtr(tt.ptr, tt.fieldName, tt.resourceType)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequiredStringPtr() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidationError tests the ValidationError custom error type
func TestValidationError(t *testing.T) {
	tests := []struct {
		name          string
		validationErr *ValidationError
		wantContains  []string
	}{
		{
			name: "error with response",
			validationErr: &ValidationError{
				ResourceType: "Notification",
				FieldName:    "Id",
				Response: &testResponse{
					ID:   nil,
					Name: "test",
				},
			},
			wantContains: []string{"Notification", "Id", "nil", "permissions"},
		},
		{
			name: "error with nil response",
			validationErr: &ValidationError{
				ResourceType: "Project",
				FieldName:    "response",
				Response:     nil,
			},
			wantContains: []string{"Project", "response", "nil"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.validationErr.Error()
			for _, want := range tt.wantContains {
				if !strings.Contains(errMsg, want) {
					t.Errorf("ValidationError.Error() should contain '%s', got: %s", want, errMsg)
				}
			}
		})
	}
}

// TestValidateNotificationResponse tests notification response validation
func TestValidateNotificationResponse(t *testing.T) {
	tests := []struct {
		name        string
		id          *int
		response    interface{}
		wantErr     bool
		errContains string
	}{
		{
			name: "valid notification",
			id:   intPtr(123),
			response: &testResponse{
				ID:   intPtr(123),
				Name: "test",
			},
			wantErr: false,
		},
		{
			name: "nil ID",
			id:   nil,
			response: &testResponse{
				ID:   nil,
				Name: "test",
			},
			wantErr:     true,
			errContains: "Id",
		},
		{
			name:        "nil response",
			id:          nil,
			response:    nil,
			wantErr:     true,
			errContains: "response",
		},
		{
			name: "ID is zero",
			id:   intPtr(0),
			response: &testResponse{
				ID:   intPtr(0),
				Name: "test",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotificationResponse(tt.id, tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNotificationResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("error should contain '%s', got: %v", tt.errContains, err)
			}
		})
	}
}

// TestValidateModelNotificationsResponse tests model notifications response validation
func TestValidateModelNotificationsResponse(t *testing.T) {
	tests := []struct {
		name        string
		id          *int
		response    interface{}
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid model notifications",
			id:       intPtr(123),
			response: &testResponse{ID: intPtr(123)},
			wantErr:  false,
		},
		{
			name:        "nil ID",
			id:          nil,
			response:    &testResponse{ID: nil},
			wantErr:     true,
			errContains: "ID",
		},
		{
			name:        "nil response",
			id:          nil,
			response:    nil,
			wantErr:     true,
			errContains: "response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateModelNotificationsResponse(tt.id, tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateModelNotificationsResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("error should contain '%s', got: %v", tt.errContains, err)
			}
		})
	}
}

// TestValidateProjectResponse tests project response validation
func TestValidateProjectResponse(t *testing.T) {
	tests := []struct {
		name        string
		id          *int
		response    interface{}
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid project",
			id:       intPtr(123),
			response: &testResponse{ID: intPtr(123), Name: "Test Project"},
			wantErr:  false,
		},
		{
			name:        "nil ID",
			id:          nil,
			response:    &testResponse{ID: nil, Name: "Test Project"},
			wantErr:     true,
			errContains: "ID",
		},
		{
			name:        "nil project",
			id:          nil,
			response:    nil,
			wantErr:     true,
			errContains: "response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProjectResponse(tt.id, tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProjectResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("error should contain '%s', got: %v", tt.errContains, err)
			}
		})
	}
}

// TestValidateServiceTokenResponse tests service token response validation
func TestValidateServiceTokenResponse(t *testing.T) {
	tests := []struct {
		name        string
		id          *int
		tokenString *string
		response    interface{}
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid service token",
			id:          intPtr(123),
			tokenString: stringPtr("secret-token"),
			response:    &testResponse{ID: intPtr(123), TokenString: stringPtr("secret-token")},
			wantErr:     false,
		},
		{
			name:        "nil ID",
			id:          nil,
			tokenString: stringPtr("secret-token"),
			response:    &testResponse{ID: nil, TokenString: stringPtr("secret-token")},
			wantErr:     true,
			errContains: "ID",
		},
		{
			name:        "empty token string",
			id:          intPtr(123),
			tokenString: stringPtr(""),
			response:    &testResponse{ID: intPtr(123), TokenString: stringPtr("")},
			wantErr:     true,
			errContains: "TokenString",
		},
		{
			name:        "nil response",
			id:          nil,
			tokenString: nil,
			response:    nil,
			wantErr:     true,
			errContains: "response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServiceTokenResponse(tt.id, tt.tokenString, tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateServiceTokenResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("error should contain '%s', got: %v", tt.errContains, err)
			}
		})
	}
}

// TestValidationErrorMessageQuality ensures error messages are helpful
func TestValidationErrorMessageQuality(t *testing.T) {
	err := &ValidationError{
		ResourceType: "Notification",
		FieldName:    "Id",
		Response: &testResponse{
			ID:   nil,
			Name: "test",
		},
	}

	errMsg := err.Error()

	// Check for key elements that make errors helpful
	requiredPhrases := []string{
		"validation failed",
		"Notification",
		"Id",
		"nil",
		"permissions",
		"API",
		"verify",
	}

	for _, phrase := range requiredPhrases {
		if !strings.Contains(strings.ToLower(errMsg), strings.ToLower(phrase)) {
			t.Errorf("Error message should contain '%s' for better user experience. Got: %s", phrase, errMsg)
		}
	}

	// Ensure error message is not too short (should be descriptive)
	if len(errMsg) < 100 {
		t.Errorf("Error message seems too short (%d chars). Should provide more context.", len(errMsg))
	}
}

// TestLogAPIResponse tests the API response logging function
func TestLogAPIResponse(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		response     interface{}
		wantContains []string
	}{
		{
			name:         "simple struct",
			resourceType: "Notification",
			response: &testResponse{
				ID:   intPtr(123),
				Name: "test",
			},
			wantContains: []string{"Notification", "Response", "123"},
		},
		{
			name:         "nil response",
			resourceType: "Project",
			response:     nil,
			wantContains: []string{"Project", "Response", "null"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logAPIResponse(tt.resourceType, tt.response)
			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("logAPIResponse() should contain '%s', got: %s", want, result)
				}
			}
		})
	}
}

// TestValidateCredentialResponseWithID tests generic credential validation
func TestValidateCredentialResponseWithID(t *testing.T) {
	tests := []struct {
		name           string
		id             *int
		credentialType string
		response       interface{}
		wantErr        bool
		errContains    string
	}{
		{
			name:           "valid credential",
			id:             intPtr(123),
			credentialType: "BigQueryCredential",
			response: &testResponse{
				ID:   intPtr(123),
				Name: "Test Cred",
			},
			wantErr: false,
		},
		{
			name:           "nil ID",
			id:             nil,
			credentialType: "BigQueryCredential",
			response:       &testResponse{ID: nil},
			wantErr:        true,
			errContains:    "ID",
		},
		{
			name:           "nil response",
			id:             nil,
			credentialType: "PostgresCredential",
			response:       nil,
			wantErr:        true,
			errContains:    "response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCredentialResponseWithID(tt.id, tt.credentialType, tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCredentialResponseWithID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("error should contain '%s', got: %v", tt.errContains, err)
			}
			if err != nil && tt.credentialType != "" && !strings.Contains(err.Error(), tt.credentialType) {
				t.Errorf("error should contain credential type '%s', got: %v", tt.credentialType, err)
			}
		})
	}
}
