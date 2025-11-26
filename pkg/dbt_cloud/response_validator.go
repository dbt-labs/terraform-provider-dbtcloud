package dbt_cloud

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ResponseValidator is a singleton instance of the validator
var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

// ValidationError wraps validator errors with the API response for debugging
type ValidationError struct {
	Message     string
	FieldErrors []FieldError
	Response    interface{}
}

type FieldError struct {
	Field   string
	Tag     string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	responseJSON, _ := json.MarshalIndent(e.Response, "", "  ")

	msg := fmt.Sprintf("API response validation failed:\n%s\n\nFields with errors:\n", e.Message)
	for _, fe := range e.FieldErrors {
		msg += fmt.Sprintf("  - %s: %s (value: %v)\n", fe.Field, fe.Message, fe.Value)
	}
	msg += fmt.Sprintf("\nFull API response:\n%s", string(responseJSON))

	return msg
}

// ValidateResponse validates a struct using validator tags
func ValidateResponse(response interface{}, resourceType string) error {
	err := validate.Struct(response)
	if err == nil {
		return nil
	}

	// Convert validation errors to our custom format
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		// Not a validation error, return as is
		return err
	}

	fieldErrors := make([]FieldError, 0, len(validationErrs))
	for _, e := range validationErrs {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   e.Field(),
			Tag:     e.Tag(),
			Value:   e.Value(),
			Message: getErrorMessage(e),
		})
	}

	return &ValidationError{
		Message:     fmt.Sprintf("Validation failed for %s", resourceType),
		FieldErrors: fieldErrors,
		Response:    response,
	}
}

// getErrorMessage returns a human-readable error message for a validation error
func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "field is required but was nil or empty"
	case "ne":
		return fmt.Sprintf("must not equal %s", e.Param())
	default:
		return fmt.Sprintf("failed validation tag '%s'", e.Tag())
	}
}
