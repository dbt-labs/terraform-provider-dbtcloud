package dbt_cloud

import (
	"strings"
	"testing"
)

func TestValidateResponse_Success(t *testing.T) {
	// Test with valid notification
	validID := 123
	notification := Notification{
		Id:        &validID,
		AccountId: 456,
		UserId:    789,
		State:     STATE_ACTIVE,
	}

	err := ValidateResponse(&notification, "Notification")
	if err != nil {
		t.Errorf("ValidateResponse() should not return error for valid notification, got: %v", err)
	}
}

func TestValidateResponse_NilID(t *testing.T) {
	// Test with nil ID
	notification := Notification{
		Id:        nil,
		AccountId: 456,
		UserId:    789,
		State:     STATE_ACTIVE,
	}

	err := ValidateResponse(&notification, "Notification")
	if err == nil {
		t.Error("ValidateResponse() should return error for nil ID")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("Expected ValidationError type, got %T", err)
	}

	if validationErr != nil {
		if len(validationErr.FieldErrors) == 0 {
			t.Error("Expected at least one field error")
		}
		if !strings.Contains(validationErr.Error(), "Id") {
			t.Errorf("Error message should mention 'Id' field, got: %s", validationErr.Error())
		}
		if !strings.Contains(validationErr.Error(), "Full API response") {
			t.Errorf("Error message should include full API response, got: %s", validationErr.Error())
		}
	}
}

func TestValidateResponse_ZeroID(t *testing.T) {
	// Test with zero ID (should fail with ne=0 validation)
	zeroID := 0
	notification := Notification{
		Id:        &zeroID,
		AccountId: 456,
		UserId:    789,
		State:     STATE_ACTIVE,
	}

	err := ValidateResponse(&notification, "Notification")
	if err == nil {
		t.Error("ValidateResponse() should return error for zero ID")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("Expected ValidationError type, got %T", err)
	}

	if validationErr != nil {
		if len(validationErr.FieldErrors) == 0 {
			t.Error("Expected at least one field error")
		}
	}
}

func TestValidateResponse_Project(t *testing.T) {
	// Test with valid project
	validID := 999
	project := Project{
		ID:          &validID,
		Name:        "Test Project",
		Description: "A test project",
		State:       STATE_ACTIVE,
		AccountID:   123,
	}

	err := ValidateResponse(&project, "Project")
	if err != nil {
		t.Errorf("ValidateResponse() should not return error for valid project, got: %v", err)
	}
}

func TestValidateResponse_ProjectNilID(t *testing.T) {
	// Test project with nil ID
	project := Project{
		ID:          nil,
		Name:        "Test Project",
		Description: "A test project",
		State:       STATE_ACTIVE,
		AccountID:   123,
	}

	err := ValidateResponse(&project, "Project")
	if err == nil {
		t.Error("ValidateResponse() should return error for nil project ID")
	}
}

func TestValidateResponse_ModelNotifications(t *testing.T) {
	// Test with valid model notifications
	validID := 777
	modelNotifications := ModelNotifications{
		ID:            &validID,
		EnvironmentID: 123,
		Enabled:       true,
	}

	err := ValidateResponse(&modelNotifications, "ModelNotifications")
	if err != nil {
		t.Errorf("ValidateResponse() should not return error for valid model notifications, got: %v", err)
	}
}

func TestValidateResponse_ModelNotificationsNilID(t *testing.T) {
	// Test model notifications with nil ID
	modelNotifications := ModelNotifications{
		ID:            nil,
		EnvironmentID: 123,
		Enabled:       true,
	}

	err := ValidateResponse(&modelNotifications, "ModelNotifications")
	if err == nil {
		t.Error("ValidateResponse() should return error for nil model notifications ID")
	}
}

func TestValidationError_ErrorMessage(t *testing.T) {
	// Test that ValidationError produces a helpful error message
	zeroID := 0
	notification := Notification{
		Id:        &zeroID,
		AccountId: 456,
		UserId:    789,
		State:     STATE_ACTIVE,
	}

	err := ValidateResponse(&notification, "Notification")
	if err == nil {
		t.Fatal("Expected validation error")
	}

	errMsg := err.Error()

	// Check that the error message contains key information
	expectedStrings := []string{
		"API response validation failed",
		"Notification",
		"Fields with errors",
		"Full API response",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(errMsg, expected) {
			t.Errorf("Error message should contain '%s', got: %s", expected, errMsg)
		}
	}
}

func TestGetErrorMessage(t *testing.T) {
	// This is an internal function, but we can test it indirectly
	// by checking the error messages produced by ValidateResponse

	tests := []struct {
		name     string
		value    interface{}
		wantText string
	}{
		{
			name:     "nil ID (required)",
			value:    &Notification{Id: nil, State: STATE_ACTIVE},
			wantText: "required",
		},
		{
			name:     "zero ID (ne)",
			value:    &Notification{Id: intPtr(0), State: STATE_ACTIVE},
			wantText: "not equal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResponse(tt.value, "Test")
			if err == nil {
				t.Fatal("Expected validation error")
			}

			errMsg := err.Error()
			if !strings.Contains(strings.ToLower(errMsg), strings.ToLower(tt.wantText)) {
				t.Errorf("Error message should contain '%s', got: %s", tt.wantText, errMsg)
			}
		})
	}
}

// Helper function
func intPtr(i int) *int {
	return &i
}
