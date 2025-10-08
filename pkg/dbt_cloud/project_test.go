package dbt_cloud_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
)

func TestIsValidSubdirectory(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		// Valid cases
		{
			name:        "valid subdirectory - simple path",
			input:       "models",
			expectError: false,
		},
		{
			name:        "valid subdirectory - nested path",
			input:       "dbt/models",
			expectError: false,
		},
		{
			name:        "valid subdirectory - deep nested path",
			input:       "src/main/dbt/models",
			expectError: false,
		},
		{
			name:        "valid subdirectory - with dots in filename",
			input:       "models/v1.0",
			expectError: false,
		},
		{
			name:        "valid subdirectory - with underscores and hyphens",
			input:       "my_models/test-data",
			expectError: false,
		},
		{
			name:        "valid subdirectory - empty string",
			input:       "",
			expectError: false,
		},
		{
			name:        "valid subdirectory - numbers and letters",
			input:       "models123/data456",
			expectError: false,
		},

		// Invalid cases - starts with slash
		{
			name:        "invalid subdirectory - starts with slash",
			input:       "/models",
			expectError: true,
			errorMsg:    `project subdirectory path should not start with a slash: "/models"`,
		},
		{
			name:        "invalid subdirectory - starts with slash nested",
			input:       "/dbt/models",
			expectError: true,
			errorMsg:    `project subdirectory path should not start with a slash: "/dbt/models"`,
		},
		{
			name:        "invalid subdirectory - absolute path",
			input:       "/usr/local/dbt/models",
			expectError: true,
			errorMsg:    `project subdirectory path should not start with a slash: "/usr/local/dbt/models"`,
		},

		// Invalid cases - ends with slash
		{
			name:        "invalid subdirectory - ends with slash",
			input:       "models/",
			expectError: true,
			errorMsg:    `project subdirectory path should not end with a slash: "models/"`,
		},
		{
			name:        "invalid subdirectory - nested path ends with slash",
			input:       "dbt/models/",
			expectError: true,
			errorMsg:    `project subdirectory path should not end with a slash: "dbt/models/"`,
		},

		// Invalid cases - relative paths
		{
			name:        "invalid subdirectory - contains dot slash",
			input:       "models/./data",
			expectError: true,
			errorMsg:    `project subdirectory path should not contain relative paths: "models/./data"`,
		},
		{
			name:        "invalid subdirectory - contains double dot slash",
			input:       "models/../data",
			expectError: true,
			errorMsg:    `project subdirectory path should not contain relative paths: "models/../data"`,
		},
		{
			name:        "invalid subdirectory - contains tilde slash",
			input:       "~/models",
			expectError: true,
			errorMsg:    `project subdirectory path should not contain relative paths: "~/models"`,
		},
		{
			name:        "invalid subdirectory - starts with dot slash",
			input:       "./models",
			expectError: true,
			errorMsg:    `project subdirectory path should not contain relative paths: "./models"`,
		},
		{
			name:        "invalid subdirectory - contains double dot slash",
			input:       "../models",
			expectError: true,
			errorMsg:    `project subdirectory path should not contain relative paths: "../models"`,
		},

		// Invalid cases - invalid characters
		{
			name:        "invalid subdirectory - contains hash",
			input:       "models#data",
			expectError: true,
			errorMsg:    fmt.Sprintf(`project subdirectory path should not contain file characters ("%s"): "models#data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:        "invalid subdirectory - contains percent",
			input:       "models%data",
			expectError: true,
			errorMsg:    fmt.Sprintf(`project subdirectory path should not contain file characters ("%s"): "models%%data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:        "invalid subdirectory - contains ampersand",
			input:       "models&data",
			expectError: true,
			errorMsg:    fmt.Sprintf(`project subdirectory path should not contain file characters ("%s"): "models&data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:        "invalid subdirectory - contains curly braces",
			input:       "models{data}",
			expectError: true,
			errorMsg:    fmt.Sprintf(`project subdirectory path should not contain file characters ("%s"): "models{data}"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:        "invalid subdirectory - contains angle brackets",
			input:       "models<data>",
			expectError: true,
			errorMsg:    fmt.Sprintf(`project subdirectory path should not contain file characters ("%s"): "models<data>"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:        "invalid subdirectory - contains asterisk",
			input:       "models*data",
			expectError: true,
			errorMsg:    fmt.Sprintf(`project subdirectory path should not contain file characters ("%s"): "models*data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:        "invalid subdirectory - contains question mark",
			input:       "models?data",
			expectError: true,
			errorMsg:    fmt.Sprintf(`project subdirectory path should not contain file characters ("%s"): "models?data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:        "invalid subdirectory - contains dollar sign",
			input:       "models$data",
			expectError: true,
			errorMsg:    fmt.Sprintf(`project subdirectory path should not contain file characters ("%s"): "models$data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:        "invalid subdirectory - contains exclamation",
			input:       "models!data",
			expectError: true,
			errorMsg:    fmt.Sprintf(`project subdirectory path should not contain file characters ("%s"): "models!data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:        "invalid subdirectory - contains single quote",
			input:       "models'data",
			expectError: true,
			errorMsg:    fmt.Sprintf(`project subdirectory path should not contain file characters ("%s"): "models'data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:        "invalid subdirectory - contains double quote",
			input:       "models\"data",
			expectError: true,
			errorMsg:    fmt.Sprintf(`project subdirectory path should not contain file characters ("%s"): "models"data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:        "invalid subdirectory - contains colon",
			input:       "models:data",
			expectError: true,
			errorMsg:    fmt.Sprintf(`project subdirectory path should not contain file characters ("%s"): "models:data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:        "invalid subdirectory - contains at symbol",
			input:       "models@data",
			expectError: true,
			errorMsg:    fmt.Sprintf(`project subdirectory path should not contain file characters ("%s"): "models@data"`, dbt_cloud.InvalidFileCharacters),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dbt_cloud.IsValidSubdirectory(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input '%s', but got nil", tt.input)
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', but got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for input '%s', but got: %v", tt.input, err)
				}
			}
		})
	}
}
