package dbt_cloud

import (
	"encoding/json"
	"fmt"
)

// validateRequiredIntPtr checks if an int pointer field is nil and returns a descriptive error
func validateRequiredIntPtr(ptr *int, fieldName, resourceType string) error {
	if ptr == nil {
		return fmt.Errorf(
			"API response validation failed: required field '%s' is nil in %s response. "+
				"This typically indicates an API issue, insufficient permissions, or a malformed response. "+
				"Please verify your API token has the necessary permissions and try again",
			fieldName,
			resourceType,
		)
	}
	return nil
}

// validateRequiredStringPtr checks if a string pointer field is nil and returns a descriptive error
func validateRequiredStringPtr(ptr *string, fieldName, resourceType string) error {
	if ptr == nil {
		return fmt.Errorf(
			"API response validation failed: required field '%s' is nil in %s response. "+
				"This typically indicates an API issue, insufficient permissions, or a malformed response. "+
				"Please verify your API token has the necessary permissions and try again",
			fieldName,
			resourceType,
		)
	}
	return nil
}

// logAPIResponse logs the full API response for debugging purposes
// This helps with troubleshooting nil pointer issues
func logAPIResponse(resourceType string, response interface{}) string {
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Sprintf("(failed to marshal %s response: %v)", resourceType, err)
	}
	return fmt.Sprintf("%s API Response:\n%s", resourceType, string(data))
}

// ValidationError represents an API response validation error with context
type ValidationError struct {
	ResourceType string
	FieldName    string
	Response     interface{}
}

func (e *ValidationError) Error() string {
	responseJSON := logAPIResponse(e.ResourceType, e.Response)
	return fmt.Sprintf(
		"API response validation failed for %s: required field '%s' is nil.\n\n"+
			"This typically indicates one of the following:\n"+
			"1. Insufficient API token permissions\n"+
			"2. API returned incomplete data due to an error\n"+
			"3. Resource was created but API response is malformed\n\n"+
			"Full API Response:\n%s\n\n"+
			"Please verify:\n"+
			"- Your API token has appropriate permissions\n"+
			"- The resource exists in dbt Cloud\n"+
			"- There are no API-level errors",
		e.ResourceType,
		e.FieldName,
		responseJSON,
	)
}

// ValidateNotificationResponse validates that a notification response has required fields
// It accepts the ID pointer and the full response for error reporting
func ValidateNotificationResponse(id *int, response interface{}) error {
	if response == nil {
		return &ValidationError{
			ResourceType: "Notification",
			FieldName:    "response",
			Response:     nil,
		}
	}

	if err := validateRequiredIntPtr(id, "Id", "Notification"); err != nil {
		return &ValidationError{
			ResourceType: "Notification",
			FieldName:    "Id",
			Response:     response,
		}
	}

	return nil
}

// ValidateModelNotificationsResponse validates that a model notifications response has required fields
// It accepts the ID pointer and the full response for error reporting
func ValidateModelNotificationsResponse(id *int, response interface{}) error {
	if response == nil {
		return &ValidationError{
			ResourceType: "ModelNotifications",
			FieldName:    "response",
			Response:     nil,
		}
	}

	if err := validateRequiredIntPtr(id, "ID", "ModelNotifications"); err != nil {
		return &ValidationError{
			ResourceType: "ModelNotifications",
			FieldName:    "ID",
			Response:     response,
		}
	}

	return nil
}

// ValidateCredentialResponseWithID validates that a credential response has required ID field
// This is a generic validator that works with any credential type that has an ID field
func ValidateCredentialResponseWithID(id *int, credentialType string, response interface{}) error {
	if response == nil {
		return &ValidationError{
			ResourceType: credentialType,
			FieldName:    "response",
			Response:     nil,
		}
	}

	if err := validateRequiredIntPtr(id, "ID", credentialType); err != nil {
		return &ValidationError{
			ResourceType: credentialType,
			FieldName:    "ID",
			Response:     response,
		}
	}

	return nil
}

// ValidateProjectResponse validates that a project response has required fields
// It accepts the ID pointer and the full response for error reporting
func ValidateProjectResponse(id *int, response interface{}) error {
	if response == nil {
		return &ValidationError{
			ResourceType: "Project",
			FieldName:    "response",
			Response:     nil,
		}
	}

	if err := validateRequiredIntPtr(id, "ID", "Project"); err != nil {
		return &ValidationError{
			ResourceType: "Project",
			FieldName:    "ID",
			Response:     response,
		}
	}

	return nil
}

// ValidateJobResponse validates that a job response has required fields
// It accepts the ID pointer and the full response for error reporting
func ValidateJobResponse(id *int, response interface{}) error {
	if response == nil {
		return &ValidationError{
			ResourceType: "Job",
			FieldName:    "response",
			Response:     nil,
		}
	}

	if err := validateRequiredIntPtr(id, "ID", "Job"); err != nil {
		return &ValidationError{
			ResourceType: "Job",
			FieldName:    "ID",
			Response:     response,
		}
	}

	return nil
}

// ValidateRepositoryResponse validates that a repository response has required fields
// It accepts the ID pointer and the full response for error reporting
func ValidateRepositoryResponse(id *int, response interface{}) error {
	if response == nil {
		return &ValidationError{
			ResourceType: "Repository",
			FieldName:    "response",
			Response:     nil,
		}
	}

	if err := validateRequiredIntPtr(id, "ID", "Repository"); err != nil {
		return &ValidationError{
			ResourceType: "Repository",
			FieldName:    "ID",
			Response:     response,
		}
	}

	return nil
}

// ValidateExtendedAttributesResponse validates that an extended attributes response has required fields
// It accepts the ID pointer and the full response for error reporting
func ValidateExtendedAttributesResponse(id *int, response interface{}) error {
	if response == nil {
		return &ValidationError{
			ResourceType: "ExtendedAttributes",
			FieldName:    "response",
			Response:     nil,
		}
	}

	if err := validateRequiredIntPtr(id, "ID", "ExtendedAttributes"); err != nil {
		return &ValidationError{
			ResourceType: "ExtendedAttributes",
			FieldName:    "ID",
			Response:     response,
		}
	}

	return nil
}

// ValidateServiceTokenResponse validates that a service token response has required fields
// It accepts the ID pointer, token string pointer and the full response for error reporting
func ValidateServiceTokenResponse(id *int, tokenString *string, response interface{}) error {
	if response == nil {
		return &ValidationError{
			ResourceType: "ServiceToken",
			FieldName:    "response",
			Response:     nil,
		}
	}

	if err := validateRequiredIntPtr(id, "ID", "ServiceToken"); err != nil {
		return &ValidationError{
			ResourceType: "ServiceToken",
			FieldName:    "ID",
			Response:     response,
		}
	}

	// TokenString is also required for service tokens during creation
	if tokenString != nil && *tokenString == "" {
		return &ValidationError{
			ResourceType: "ServiceToken",
			FieldName:    "TokenString (empty)",
			Response:     response,
		}
	}

	return nil
}

// ValidateGroupResponse validates that a group response has required fields
// It accepts the ID pointer and the full response for error reporting
func ValidateGroupResponse(id *int, response interface{}) error {
	if response == nil {
		return &ValidationError{
			ResourceType: "Group",
			FieldName:    "response",
			Response:     nil,
		}
	}

	if err := validateRequiredIntPtr(id, "ID", "Group"); err != nil {
		return &ValidationError{
			ResourceType: "Group",
			FieldName:    "ID",
			Response:     response,
		}
	}

	return nil
}

// ValidateEnvironmentVariableJobOverrideResponse validates environment variable job override response
// It accepts the ID pointer and the full response for error reporting
func ValidateEnvironmentVariableJobOverrideResponse(id *int, response interface{}) error {
	if response == nil {
		return &ValidationError{
			ResourceType: "EnvironmentVariableJobOverride",
			FieldName:    "response",
			Response:     nil,
		}
	}

	if err := validateRequiredIntPtr(id, "ID", "EnvironmentVariableJobOverride"); err != nil {
		return &ValidationError{
			ResourceType: "EnvironmentVariableJobOverride",
			FieldName:    "ID",
			Response:     response,
		}
	}

	return nil
}

// ValidateSemanticLayerCredentialResponse validates semantic layer credential response
func ValidateSemanticLayerCredentialResponse(cred interface{}, credType string) error {
	// This is a generic validator for semantic layer credentials
	// The specific type checking is done at the caller level
	if cred == nil {
		return &ValidationError{
			ResourceType: credType,
			FieldName:    "response",
			Response:     nil,
		}
	}

	return nil
}
