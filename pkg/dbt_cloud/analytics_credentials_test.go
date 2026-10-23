package dbt_cloud

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// Rule: a credential is fetched from the account's analytics-credentials endpoint.
// given: a credential exists on the server
// when: it is fetched by id
// then: the request goes to the client's host and account and the response is decoded.
func TestGetAnalyticsCredential_RequestAndDecode(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_, _ = w.Write([]byte(`{"data":{"id":7,"account_id":123,"connection_id":42,"adapter_version":"snowflake_v0","config":{"auth_type":"password","user":"u","warehouse":"w"}},"status":{"code":200}}`))
	}))
	defer srv.Close()
	c := newRetryTestClient(t, srv.URL, 1, nil)

	cred, err := c.GetAnalyticsCredential(7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/v3/accounts/123/analytics-credentials/7/" {
		t.Fatalf("unexpected request: %s %s", gotMethod, gotPath)
	}
	if cred.ID == nil || *cred.ID != 7 || cred.ConnectionID != 42 ||
		cred.AdapterVersion != "snowflake_v0" || cred.Config.User != "u" ||
		cred.Config.Warehouse != "w" || cred.Config.AuthType != "password" {
		t.Fatalf("unexpected credential: %+v", cred)
	}
}

// Rule: a missing credential is reported as an error.
// given: the server answers 404
// when: the credential is fetched
// then: an error is returned.
func TestGetAnalyticsCredential_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":{"code":404}}`))
	}))
	defer srv.Close()
	c := newRetryTestClient(t, srv.URL, 1, nil)

	if _, err := c.GetAnalyticsCredential(7); err == nil {
		t.Fatal("expected an error on 404")
	}
}

// Rule: listing without filters sends no query string.
// given: no filters
// when: credentials are listed
// then: the request has no query string and the items are decoded.
func TestListAnalyticsCredentials_NoFilters(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
		_, _ = w.Write([]byte(`{"data":[{"id":1,"connection_id":42},{"id":2,"connection_id":43}],"extra":{"pagination":{"count":2,"total_count":2}}}`))
	}))
	defer srv.Close()
	c := newRetryTestClient(t, srv.URL, 1, nil)

	creds, err := c.ListAnalyticsCredentials(nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/v3/accounts/123/analytics-credentials/" || gotQuery != "" {
		t.Fatalf("unexpected request: %s ?%s", gotPath, gotQuery)
	}
	if len(creds) != 2 || *creds[0].ID != 1 || creds[1].ConnectionID != 43 {
		t.Fatalf("unexpected credentials: %+v", creds)
	}
}

// Rule: every page of a filtered list is read, keeping the filters.
// given: both filters set and a result spanning two pages
// when: credentials are listed
// then: the second request keeps the filters and adds the offset, and the items come back in order.
func TestListAnalyticsCredentials_FiltersAndPagination(t *testing.T) {
	var queries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.RawQuery)
		if r.URL.Query().Get("offset") == "" {
			_, _ = w.Write([]byte(`{"data":[{"id":1}],"extra":{"pagination":{"count":1,"total_count":2}}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":2}],"extra":{"pagination":{"count":1,"total_count":2}}}`))
	}))
	defer srv.Close()
	c := newRetryTestClient(t, srv.URL, 1, nil)

	connectionID := int64(42)
	creds, err := c.ListAnalyticsCredentials(&connectionID, "snowflake_v0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(queries) != 2 {
		t.Fatalf("expected 2 requests, got %v", queries)
	}
	for _, q := range queries {
		v, _ := url.ParseQuery(q)
		if v.Get("connection_id") != "42" || v.Get("adapter_version") != "snowflake_v0" {
			t.Fatalf("filters missing from query %q", q)
		}
	}
	if v, _ := url.ParseQuery(queries[1]); v.Get("offset") != "1" {
		t.Fatalf("expected offset=1 on the second request, got %q", queries[1])
	}
	if len(creds) != 2 || *creds[0].ID != 1 || *creds[1].ID != 2 {
		t.Fatalf("unexpected credentials: %+v", creds)
	}
}

// Rule: a failing page is reported, not fatal.
// given: the second page answers 404
// when: credentials are listed
// then: an error is returned and the process keeps running.
func TestListAnalyticsCredentials_FailingPageReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("offset") != "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":1}],"extra":{"pagination":{"count":1,"total_count":2}}}`))
	}))
	defer srv.Close()
	c := newRetryTestClient(t, srv.URL, 1, nil)

	if _, err := c.ListAnalyticsCredentials(nil, ""); err == nil {
		t.Fatal("expected an error when a page fails")
	}
}

// assertBodyKeys fails unless the JSON request body has exactly the wanted top-level keys.
func assertBodyKeys(t *testing.T, r *http.Request, want ...string) map[string]any {
	t.Helper()
	body := map[string]any{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Errorf("decode request body: %v", err)
		return body
	}
	if len(body) != len(want) {
		t.Errorf("expected body keys %v, got %v", want, body)
	}
	for _, k := range want {
		if _, ok := body[k]; !ok {
			t.Errorf("missing key %q in body %v", k, body)
		}
	}
	return body
}

// Rule: the adapter is fixed by the connection.
// given: a create call
// when: the request is sent
// then: the request body, decoded into map[string]any, has exactly the keys account_id, connection_id and config.
func TestCreateAnalyticsCredential_BodyAndDecode(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		body := assertBodyKeys(t, r, "account_id", "connection_id", "config")
		if body["account_id"] != float64(123) || body["connection_id"] != float64(42) {
			t.Errorf("unexpected ids in body: %v", body)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"data":{"id":9,"account_id":123,"connection_id":42,"adapter_version":"snowflake_v0","config":{"user":"u"}}}`))
	}))
	defer srv.Close()
	c := newRetryTestClient(t, srv.URL, 1, nil)

	cred, err := c.CreateAnalyticsCredential(42, AnalyticsCredentialConfig{User: "u", Password: "p"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/v3/accounts/123/analytics-credentials/" {
		t.Fatalf("unexpected request: %s %s", gotMethod, gotPath)
	}
	if cred.ID == nil || *cred.ID != 9 || cred.AdapterVersion != "snowflake_v0" {
		t.Fatalf("unexpected credential: %+v", cred)
	}
}

// Rule: a rejected create is reported.
// given: the server answers 400
// when: a credential is created
// then: an error is returned.
func TestCreateAnalyticsCredential_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":{"code":400}}`))
	}))
	defer srv.Close()
	c := newRetryTestClient(t, srv.URL, 1, nil)

	if _, err := c.CreateAnalyticsCredential(42, AnalyticsCredentialConfig{}); err == nil {
		t.Fatal("expected an error on 400")
	}
}

// Rule: the adapter of a credential cannot change.
// given: an update call
// when: the request is sent
// then: the request body, decoded into map[string]any, has exactly the keys account_id and config.
func TestUpdateAnalyticsCredential_BodyAndDecode(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		assertBodyKeys(t, r, "account_id", "config")
		_, _ = w.Write([]byte(`{"data":{"id":9,"account_id":123,"connection_id":42,"adapter_version":"snowflake_v0","config":{"user":"new"}}}`))
	}))
	defer srv.Close()
	c := newRetryTestClient(t, srv.URL, 1, nil)

	cred, err := c.UpdateAnalyticsCredential(9, AnalyticsCredentialConfig{User: "new"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPatch || gotPath != "/v3/accounts/123/analytics-credentials/9/" {
		t.Fatalf("unexpected request: %s %s", gotMethod, gotPath)
	}
	if cred.ID == nil || *cred.ID != 9 || cred.Config.User != "new" {
		t.Fatalf("unexpected credential: %+v", cred)
	}
}

// Rule: a rejected update is reported.
// given: the server answers 400
// when: a credential is updated
// then: an error is returned.
func TestUpdateAnalyticsCredential_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":{"code":400}}`))
	}))
	defer srv.Close()
	c := newRetryTestClient(t, srv.URL, 1, nil)

	if _, err := c.UpdateAnalyticsCredential(9, AnalyticsCredentialConfig{}); err == nil {
		t.Fatal("expected an error on 400")
	}
}

// Rule: a credential is deleted through its own endpoint.
// given: the server answers 200 with the deleted id
// when: the credential is deleted
// then: a DELETE goes to the credential's URL and no error is returned.
func TestDeleteAnalyticsCredential(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_, _ = w.Write([]byte(`{"data":{"id":9}}`))
	}))
	defer srv.Close()
	c := newRetryTestClient(t, srv.URL, 1, nil)

	if err := c.DeleteAnalyticsCredential(9); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/v3/accounts/123/analytics-credentials/9/" {
		t.Fatalf("unexpected request: %s %s", gotMethod, gotPath)
	}
}

// Rule: deleting a missing credential is reported.
// given: the server answers 404
// when: the credential is deleted
// then: an error is returned.
func TestDeleteAnalyticsCredential_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":{"code":404}}`))
	}))
	defer srv.Close()
	c := newRetryTestClient(t, srv.URL, 1, nil)

	if err := c.DeleteAnalyticsCredential(9); err == nil {
		t.Fatal("expected an error on 404")
	}
}

// Rule: the analytics credential config sends only the fields that are set, under the server's key names.
// given: a Redshift config with a username and a password
// when: it is serialized
// then: the JSON has exactly the keys username and password.
func TestAnalyticsCredentialConfig_RedshiftKeys(t *testing.T) {
	body, err := json.Marshal(AnalyticsCredentialConfig{Username: "analytics_user", Password: "secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := map[string]any{}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got["username"] != "analytics_user" || got["password"] != "secret" {
		t.Fatalf("unexpected keys: %v", got)
	}
}

// Rule: the adapter version decides the BigQuery config keys.
// given: a v0 service account config, a v1 service account config and a v1 workload identity federation config
// when: they are serialized
// then: v0 has no auth_type, and the WIF body has the provider path and no service account keys.
func TestAnalyticsCredentialConfig_BigQueryKeys(t *testing.T) {
	keys := func(config AnalyticsCredentialConfig) map[string]any {
		body, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := map[string]any{}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return got
	}

	v0 := keys(AnalyticsCredentialConfig{
		PrivateKeyID: "id", PrivateKey: "k", ClientEmail: "e", ClientID: "c", AuthURI: "a",
		TokenURI: "t", AuthProviderX509CertURL: "p", ClientX509CertURL: "x", ExecutionProject: "proj",
	})
	if _, ok := v0["auth_type"]; ok || len(v0) != 9 {
		t.Fatalf("unexpected v0 keys: %v", v0)
	}
	for _, key := range []string{"private_key_id", "private_key", "client_email", "client_id", "auth_uri",
		"token_uri", "auth_provider_x509_cert_url", "client_x509_cert_url", "execution_project"} {
		if _, ok := v0[key]; !ok {
			t.Fatalf("missing key %s in %v", key, v0)
		}
	}

	wif := keys(AnalyticsCredentialConfig{
		AuthType:                       "external-oauth-wif",
		WorkloadPoolProviderPath:       "//iam.googleapis.com/pool",
		ServiceAccountImpersonationURL: "https://example.com/impersonate",
	})
	if len(wif) != 3 || wif["auth_type"] != "external-oauth-wif" ||
		wif["workload_pool_provider_path"] != "//iam.googleapis.com/pool" ||
		wif["service_account_impersonation_url"] != "https://example.com/impersonate" {
		t.Fatalf("unexpected WIF keys: %v", wif)
	}
}
