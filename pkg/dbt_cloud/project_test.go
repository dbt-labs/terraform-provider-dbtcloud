package dbt_cloud_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
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

// newProjectTestClient builds a Client pointed at the given httptest.Server,
// mirroring the acceptance-test client setup but with retries/timeouts
// zeroed out so unit tests never block on backoff.
func newProjectTestClient(t *testing.T, serverURL string) *dbt_cloud.Client {
	t.Helper()

	accountID := int64(123)
	token := "test_token"
	maxRetries := 0
	retryInterval := 0
	timeout := 5

	c, err := dbt_cloud.NewClient(
		&accountID,
		&token,
		&serverURL,
		&maxRetries,
		&retryInterval,
		nil,
		true, // skipCredentialsValidation
		&timeout,
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

// TestUpdateProject_DeleteSkipsSubdirectoryValidation locks in a fix for a bug
// where UpdateProject validated DbtProjectSubdirectory even on a soft-delete
// (State = STATE_DELETED). Since deleting a project never changes that field,
// a legacy project whose subdirectory predates today's validation rules (e.g.
// a leading "/") could never be destroyed via terraform destroy or the API -
// UpdateProject returned a client-side validation error before the request
// was ever sent. Observed in production against the "Terraform TEST"
// acceptance-test account: several "Test Project" entries with
// dbt_project_subdirectory="/dbt" were permanently stuck, e.g.
//
//	FAILED to delete project 452153 (Test Project): project subdirectory path should not start with a slash: "/dbt"
func TestUpdateProject_DeleteSkipsSubdirectoryValidation(t *testing.T) {
	invalidSubdirectory := "/dbt"

	t.Run("delete with legacy-invalid subdirectory succeeds", func(t *testing.T) {
		var requests atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{},"status":{"code":200}}`))
		}))
		defer srv.Close()

		client := newProjectTestClient(t, srv.URL)
		project := dbt_cloud.Project{
			DbtProjectSubdirectory: &invalidSubdirectory,
			State:                  dbt_cloud.STATE_DELETED,
		}

		_, err := client.UpdateProject("123", project)

		assert.NoError(t, err)
		assert.Equal(t, int32(1), requests.Load(), "expected the delete request to actually reach the API")
	})

	t.Run("non-delete update with the same invalid subdirectory is still rejected", func(t *testing.T) {
		var requests atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{},"status":{"code":200}}`))
		}))
		defer srv.Close()

		client := newProjectTestClient(t, srv.URL)
		project := dbt_cloud.Project{
			DbtProjectSubdirectory: &invalidSubdirectory,
			State:                  0,
		}

		_, err := client.UpdateProject("123", project)

		assert.Error(t, err)
		assert.Equal(t, int32(0), requests.Load(), "invalid subdirectory should be rejected client-side before any request is sent")
	})
}
