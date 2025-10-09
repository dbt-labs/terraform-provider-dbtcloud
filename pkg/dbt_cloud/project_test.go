package dbt_cloud_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/stretchr/testify/assert"
)

func TestIsValidSubdirectory(t *testing.T) {
	tests := []struct {
		name  string
		input string
		err   error
	}{
		// Valid cases
		{
			name:  "valid subdirectory - simple path",
			input: "models",
			err:   nil,
		},
		{
			name:  "valid subdirectory - nested path",
			input: "dbt/models",
			err:   nil,
		},
		{
			name:  "valid subdirectory - deep nested path",
			input: "src/main/dbt/models",
			err:   nil,
		},
		{
			name:  "valid subdirectory - with dots in filename",
			input: "models/v1.0",
			err:   nil,
		},
		{
			name:  "valid subdirectory - with underscores and hyphens",
			input: "my_models/test-data",
			err:   nil,
		},
		{
			name:  "valid subdirectory - empty string",
			input: "",
			err:   nil,
		},
		{
			name:  "valid subdirectory - numbers and letters",
			input: "models123/data456",
			err:   nil,
		},

		// Invalid cases - starts with slash
		{
			name:  "invalid subdirectory - starts with slash",
			input: "/models",
			err:   fmt.Errorf(`project subdirectory path should not start with a slash: "/models"`),
		},
		{
			name:  "invalid subdirectory - starts with slash nested",
			input: "/dbt/models",
			err:   fmt.Errorf(`project subdirectory path should not start with a slash: "/dbt/models"`),
		},
		{
			name:  "invalid subdirectory - absolute path",
			input: "/usr/local/dbt/models",
			err:   fmt.Errorf(`project subdirectory path should not start with a slash: "/usr/local/dbt/models"`),
		},

		// Invalid cases - ends with slash
		{
			name:  "invalid subdirectory - ends with slash",
			input: "models/",
			err:   fmt.Errorf(`project subdirectory path should not end with a slash: "models/"`),
		},
		{
			name:  "invalid subdirectory - nested path ends with slash",
			input: "dbt/models/",
			err:   fmt.Errorf(`project subdirectory path should not end with a slash: "dbt/models/"`),
		},

		// Invalid cases - relative paths
		{
			name:  "invalid subdirectory - contains dot slash",
			input: "models/./data",
			err:   fmt.Errorf(`project subdirectory path should not contain relative paths: "models/./data"`),
		},
		{
			name:  "invalid subdirectory - contains double dot slash",
			input: "models/../data",
			err:   fmt.Errorf(`project subdirectory path should not contain relative paths: "models/../data"`),
		},
		{
			name:  "invalid subdirectory - contains tilde slash",
			input: "~/models",
			err:   fmt.Errorf(`project subdirectory path should not contain relative paths: "~/models"`),
		},
		{
			name:  "invalid subdirectory - starts with dot slash",
			input: "./models",
			err:   fmt.Errorf(`project subdirectory path should not contain relative paths: "./models"`),
		},
		{
			name:  "invalid subdirectory - contains double dot slash",
			input: "../models",
			err:   fmt.Errorf(`project subdirectory path should not contain relative paths: "../models"`),
		},

		// Invalid cases - invalid characters
		{
			name:  "invalid subdirectory - contains hash",
			input: "models#data",
			err:   fmt.Errorf(`project subdirectory path should not contain file characters ("%s"): "models#data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:  "invalid subdirectory - contains percent",
			input: "models%data",
			err:   fmt.Errorf(`project subdirectory path should not contain file characters ("%s"): "models%%data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:  "invalid subdirectory - contains ampersand",
			input: "models&data",
			err:   fmt.Errorf(`project subdirectory path should not contain file characters ("%s"): "models&data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:  "invalid subdirectory - contains curly braces",
			input: "models{data}",
			err:   fmt.Errorf(`project subdirectory path should not contain file characters ("%s"): "models{data}"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:  "invalid subdirectory - contains angle brackets",
			input: "models<data>",
			err:   fmt.Errorf(`project subdirectory path should not contain file characters ("%s"): "models<data>"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:  "invalid subdirectory - contains asterisk",
			input: "models*data",
			err:   fmt.Errorf(`project subdirectory path should not contain file characters ("%s"): "models*data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:  "invalid subdirectory - contains question mark",
			input: "models?data",
			err:   fmt.Errorf(`project subdirectory path should not contain file characters ("%s"): "models?data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:  "invalid subdirectory - contains dollar sign",
			input: "models$data",
			err:   fmt.Errorf(`project subdirectory path should not contain file characters ("%s"): "models$data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:  "invalid subdirectory - contains exclamation",
			input: "models!data",
			err:   fmt.Errorf(`project subdirectory path should not contain file characters ("%s"): "models!data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:  "invalid subdirectory - contains single quote",
			input: "models'data",
			err:   fmt.Errorf(`project subdirectory path should not contain file characters ("%s"): "models'data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:  "invalid subdirectory - contains double quote",
			input: "models\"data",
			err:   fmt.Errorf(`project subdirectory path should not contain file characters ("%s"): "models"data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:  "invalid subdirectory - contains colon",
			input: "models:data",
			err:   fmt.Errorf(`project subdirectory path should not contain file characters ("%s"): "models:data"`, dbt_cloud.InvalidFileCharacters),
		},
		{
			name:  "invalid subdirectory - contains at symbol",
			input: "models@data",
			err:   fmt.Errorf(`project subdirectory path should not contain file characters ("%s"): "models@data"`, dbt_cloud.InvalidFileCharacters),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dbt_cloud.IsValidSubdirectory(tt.input)

			assert.Equal(t, tt.err, err)
		})
	}
}
