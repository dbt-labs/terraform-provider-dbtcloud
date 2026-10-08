package dbt_cloud

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestUpdateBigQueryCredentialGlobConn_SendsDatasetInCredentialDetails checks the payload
// shape of a v1 credential update. Those credentials keep the dataset in
// credential_details rather than in `schema`, and the endpoint rejects the other
// top-level fields of the full credential object. threads is the exception: the
// credential keeps its own copy of it, which only gets refreshed when the payload
// carries threads at the top level too.
func TestUpdateBigQueryCredentialGlobConn_SendsDatasetInCredentialDetails(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"data":{"id":42,"account_id":123,"project_id":7,"type":"adapter","state":1,"threads":8},"status":{"code":200,"is_success":true}}`))
	}))
	defer srv.Close()

	c := newRetryTestClient(t, srv.URL, 1, nil)

	credentialDetails, err := GenerateBigQueryCredentialDetails(
		"my_new_dataset",
		8,
		"external-oauth-wif",
		"projects/1/locations/global/workloadIdentityPools/p/providers/pr",
		"",
	)
	if err != nil {
		t.Fatalf("GenerateBigQueryCredentialDetails: %v", err)
	}

	credential, err := c.UpdateBigQueryCredentialGlobConn(7, 42, BigQueryCredentialGlobConnPatch{
		Threads:           8,
		CredentialDetails: credentialDetails,
	})
	if err != nil {
		t.Fatalf("UpdateBigQueryCredentialGlobConn: %v", err)
	}
	if credential.ID == nil || *credential.ID != 42 {
		t.Fatalf("expected the credential from the response, got %+v", credential)
	}

	if gotMethod != http.MethodPatch {
		t.Errorf("expected PATCH, got %s", gotMethod)
	}
	if want := "/v3/accounts/123/projects/7/credentials/42/"; gotPath != want {
		t.Errorf("expected path %s, got %s", want, gotPath)
	}

	// the API rejects these at the top level for v1 credentials
	for _, field := range []string{
		"account_id",
		"adapter_version",
		"id",
		"project_id",
		"schema",
		"state",
		"type",
		"unencrypted_credential_details",
	} {
		if _, found := gotBody[field]; found {
			t.Errorf("%s should not be sent for a v1 credential update", field)
		}
	}

	// threads is kept on the credential as well as in credential_details, and the copy
	// on the credential goes stale when the payload leaves it out
	if got := gotBody["threads"]; got != float64(8) {
		t.Errorf("expected threads to be sent at the top level, got %v", got)
	}

	fields, ok := gotBody["credential_details"].(map[string]any)
	if !ok {
		t.Fatalf("expected credential_details in the payload, got %v", gotBody)
	}
	values, ok := fields["fields"].(map[string]any)
	if !ok {
		t.Fatalf("expected credential_details.fields in the payload, got %v", fields)
	}

	for field, want := range map[string]any{
		"schema":                      "my_new_dataset",
		"threads":                     float64(8),
		"auth_type":                   "external-oauth-wif",
		"workload_pool_provider_path": "projects/1/locations/global/workloadIdentityPools/p/providers/pr",
	} {
		value, ok := values[field].(map[string]any)
		if !ok {
			t.Fatalf("expected credential_details.fields.%s in the payload, got %v", field, values)
		}
		if value["value"] != want {
			t.Errorf("expected credential_details.fields.%s to be %v, got %v", field, want, value["value"])
		}
	}
}
