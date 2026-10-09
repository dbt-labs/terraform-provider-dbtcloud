package dbt_cloud

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestUpdateAuthProvider_PreservesUseAuth0 checks that an update carries use_auth0
// back to the API. The auth provider update endpoint is a POST that replaces the
// record rather than merging into it, and the field defaults to false server-side,
// so dropping it from the payload un-migrates an account that already moved to
// Auth0 and brings the "Begin migration" banner back in the UI.
func TestUpdateAuthProvider_PreservesUseAuth0(t *testing.T) {
	var gotBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if err := json.Unmarshal(body, &gotBody); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"id":5,"account_id":123,"type":"azure_single_tenant","slug":"acme","state":1,"use_auth0":true},"status":{"code":200,"is_success":true}}`))
	}))
	defer srv.Close()

	c := newRetryTestClient(t, srv.URL, 1, nil)

	current, err := c.GetAuthProvider(5)
	if err != nil {
		t.Fatalf("GetAuthProvider returned an error: %v", err)
	}
	if current.UseAuth0 == nil || !*current.UseAuth0 {
		t.Fatalf("expected use_auth0 to be read back as true, got %v", current.UseAuth0)
	}

	// A migrated account changing an unrelated field, which is what every apply
	// against an existing auth provider does.
	domain := "acme.example.com"
	current.Domain = &domain

	if _, err := c.UpdateAuthProvider(5, *current); err != nil {
		t.Fatalf("UpdateAuthProvider returned an error: %v", err)
	}

	if gotBody["use_auth0"] != true {
		t.Errorf("update payload must keep use_auth0 true, got %v", gotBody["use_auth0"])
	}
}

// TestUpdateAuthProvider_OmitsUnsetUseAuth0 checks that an account which never
// migrated sends no use_auth0 at all, leaving the server default in place.
func TestUpdateAuthProvider_OmitsUnsetUseAuth0(t *testing.T) {
	var gotBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &gotBody)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"id":5,"account_id":123,"type":"azure_single_tenant","slug":"acme","state":1,"use_auth0":false},"status":{"code":200,"is_success":true}}`))
	}))
	defer srv.Close()

	c := newRetryTestClient(t, srv.URL, 1, nil)

	if _, err := c.UpdateAuthProvider(5, AuthProvider{Type: "azure_single_tenant"}); err != nil {
		t.Fatalf("UpdateAuthProvider returned an error: %v", err)
	}

	if _, present := gotBody["use_auth0"]; present {
		t.Errorf("use_auth0 must be omitted when it was never set, got %v", gotBody["use_auth0"])
	}
}
