package dbt_cloud

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Rule: the list of connections carries the analytics credential attached to each one.
// given: the API lists one connection with an analytics credential and one without
// when: all connections are read
// then: the first shows the credential id and the second shows none.
func TestGetAllConnections_AnalyticsCredentialsID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":1,"name":"a","analytics_credentials_id":42},{"id":2,"name":"b","analytics_credentials_id":null}],"extra":{"pagination":{"count":2,"total_count":2}}}`))
	}))
	defer srv.Close()
	c := newRetryTestClient(t, srv.URL, 1, nil)

	connections, err := c.GetAllConnections()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(connections) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(connections))
	}
	if connections[0].AnalyticsCredentialsID == nil || *connections[0].AnalyticsCredentialsID != 42 {
		t.Fatalf("expected credential id 42, got %v", connections[0].AnalyticsCredentialsID)
	}
	if connections[1].AnalyticsCredentialsID != nil {
		t.Fatalf("expected no credential id, got %v", *connections[1].AnalyticsCredentialsID)
	}
}
