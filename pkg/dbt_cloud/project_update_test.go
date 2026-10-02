package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateProject_ExcludesConnectionID(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal(body, &gotBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"id":123,"account_id":123,"name":"test_project","type":1,"state":1},"status":{"code":200,"is_success":true}}`))
	}))
	defer srv.Close()

	c := newRetryTestClient(t, srv.URL, 1, nil)

	connID := 999
	projID := 123
	project := Project{
		ID:           &projID,
		Name:         "test_project",
		AccountID:    c.AccountID,
		ConnectionID: &connID,
		State:        STATE_ACTIVE,
	}

	updated, err := c.UpdateProject("123", project)
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, fmt.Sprintf("/v3/accounts/%d/projects/123/", c.AccountID), gotPath)
	assert.NotNil(t, updated)

	// Verify connection_id was stripped from request body so it doesn't cascade to environments
	_, exists := gotBody["connection_id"]
	assert.False(t, exists, "connection_id should not be included in the UpdateProject payload")
}
